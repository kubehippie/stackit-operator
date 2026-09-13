/*
Copyright 2026 Thomas Boerger <thomas@webhippie.de>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	postgresv1alpha1 "github.com/kubehippie/stackit-operator/api/postgres/v1alpha1"
	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
	postgresflex "github.com/stackitcloud/stackit-sdk-go/services/postgresflex/v3api"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const postgresUserFinalizer = "stackit-operator.webhippie.de/postgresuser"

// PostgresUserReconciler reconciles a PostgresUser object
type PostgresUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=postgres.stackit-operator.webhippie.de,resources=postgresusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgres.stackit-operator.webhippie.de,resources=postgresusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgres.stackit-operator.webhippie.de,resources=postgresusers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=create;delete;get;list;patch;update;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PostgresUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling")

	instance := &postgresv1alpha1.PostgresUser{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to fetch: %w", err)
	}

	session, err := NewPostgresFlexSession(ctx, r.Client, instance.Spec.CredentialsRef, instance.Namespace)
	if err != nil {
		if !instance.DeletionTimestamp.IsZero() {
			log.Info("Unable to build Postgres Flex session during deletion, dropping finalizer", "error", err.Error())
			return r.removeFinalizer(ctx, instance)
		}
		return ctrl.Result{}, fmt.Errorf("unable to build Postgres Flex session: %w", err)
	}

	instanceID, err := session.ResolveInstanceID(ctx, instance.Spec.InstanceName)
	if err != nil {
		if !instance.DeletionTimestamp.IsZero() {
			log.Info("Unable to resolve instance during deletion, dropping finalizer", "error", err.Error())
			return r.removeFinalizer(ctx, instance)
		}
		return ctrl.Result{}, fmt.Errorf("unable to resolve instance: %w", err)
	}

	if err := r.updateInstanceIDStatus(ctx, instance, &instanceID); err != nil {
		return ctrl.Result{}, err
	}

	if !instance.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, instance, session, instanceID)
	}

	if !controllerutil.ContainsFinalizer(instance, postgresUserFinalizer) {
		controllerutil.AddFinalizer(instance, postgresUserFinalizer)
		if err := r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer: %w", err)
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	return r.reconcileUser(ctx, instance, session, instanceID)
}

func (r *PostgresUserReconciler) reconcileUser(ctx context.Context, instance *postgresv1alpha1.PostgresUser, session *PostgresFlexSession, instanceID string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	existing, err := session.FindUserByName(ctx, instanceID, instance.Spec.Username)
	if err != nil {
		return ctrl.Result{}, err
	}

	secretName := instance.Name
	if instance.Spec.Secret != nil && instance.Spec.Secret.Name != nil && *instance.Spec.Secret.Name != "" {
		secretName = *instance.Spec.Secret.Name
	}
	secretKeys := resolveSecretKeys(instance.Spec.Secret)

	if existing == nil {
		resp, err := session.Client.CreateUser(ctx, session.ProjectID, session.Region, instanceID).
			CreateUserPayload(postgresflex.CreateUserPayload{
				Name:  instance.Spec.Username,
				Roles: instance.Spec.Roles,
			}).Execute()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create user in Postgres Flex: %w", err)
		}
		log.Info("User created in Postgres Flex", "id", resp.Id)

		if err := r.updateUserIDStatus(ctx, instance, &resp.Id); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.writeSecret(ctx, instance, session, instanceID, secretName, secretKeys, resp.Name, resp.Password); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.updateRotationStatus(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}

		return r.rotationResult(instance)
	}

	if err := r.updateUserIDStatus(ctx, instance, &existing.Id); err != nil {
		return ctrl.Result{}, err
	}

	if !rolesEqual(existing.Roles, instance.Spec.Roles) {
		if err := session.Client.UpdateUser(ctx, session.ProjectID, session.Region, instanceID, existing.Id).
			UpdateUserPayload(postgresflex.UpdateUserPayload{
				Name:  &instance.Spec.Username,
				Roles: instance.Spec.Roles,
			}).Execute(); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update user roles in Postgres Flex: %w", err)
		}
		log.Info("User roles updated in Postgres Flex")
	}

	secret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: secretName}, secret)
	switch {
	case apierrors.IsNotFound(err):
		log.Info("Connection secret missing, resetting password to recreate it")
		resp, err := session.Client.ResetUserPassword(ctx, session.ProjectID, session.Region, instanceID, existing.Id).Execute()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to reset user password in Postgres Flex: %w", err)
		}
		if err := r.writeSecret(ctx, instance, session, instanceID, secretName, secretKeys, existing.Name, resp.Password); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.updateRotationStatus(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	case err != nil:
		return ctrl.Result{}, fmt.Errorf("failed to check for connection secret: %w", err)
	default:
		if needsRotation(instance) {
			log.Info("Rotating user password in Postgres Flex")
			resp, err := session.Client.ResetUserPassword(ctx, session.ProjectID, session.Region, instanceID, existing.Id).Execute()
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to reset user password in Postgres Flex: %w", err)
			}
			if err := r.writeSecret(ctx, instance, session, instanceID, secretName, secretKeys, existing.Name, resp.Password); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.updateRotationStatus(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return r.rotationResult(instance)
}

// resolveSecretKeys returns the effective Secret data key names, applying
// any overrides from spec.secret.keys on top of the defaults.
func resolveSecretKeys(secret *postgresv1alpha1.PostgresUserSecretSpec) postgresUserSecretKeys {
	keys := postgresUserSecretKeys{
		Username: "username",
		Password: "password",
		Host:     "host",
		Port:     "port",
	}
	if secret == nil || secret.Keys == nil {
		return keys
	}
	if secret.Keys.Username != nil && *secret.Keys.Username != "" {
		keys.Username = *secret.Keys.Username
	}
	if secret.Keys.Password != nil && *secret.Keys.Password != "" {
		keys.Password = *secret.Keys.Password
	}
	if secret.Keys.Host != nil && *secret.Keys.Host != "" {
		keys.Host = *secret.Keys.Host
	}
	if secret.Keys.Port != nil && *secret.Keys.Port != "" {
		keys.Port = *secret.Keys.Port
	}
	return keys
}

// postgresUserSecretKeys holds the resolved Secret data key names for a
// PostgresUser's connection details.
type postgresUserSecretKeys struct {
	Username string
	Password string
	Host     string
	Port     string
}

// needsRotation reports whether instance's password should be rotated,
// based on spec.rotation.trigger having changed since the last reconciled
// value, or spec.rotation.interval having elapsed since the last rotation.
func needsRotation(instance *postgresv1alpha1.PostgresUser) bool {
	rotation := instance.Spec.Rotation
	if rotation == nil {
		return false
	}

	status := instance.Status.Rotation

	if rotation.Trigger != nil {
		if status == nil || status.Trigger == nil || *status.Trigger != *rotation.Trigger {
			return true
		}
	}

	if rotation.Interval != nil {
		interval, err := time.ParseDuration(*rotation.Interval)
		if err != nil {
			return false
		}
		if status == nil || status.LastRotatedTime == nil {
			return true
		}
		if time.Since(status.LastRotatedTime.Time) >= interval {
			return true
		}
	}

	return false
}

// rotationResult returns a Result requesting a requeue shortly before the
// next interval-based rotation is due, so scheduled rotation happens even
// without any other trigger for reconciliation.
func (r *PostgresUserReconciler) rotationResult(instance *postgresv1alpha1.PostgresUser) (ctrl.Result, error) {
	rotation := instance.Spec.Rotation
	if rotation == nil || rotation.Interval == nil {
		return ctrl.Result{}, nil
	}

	interval, err := time.ParseDuration(*rotation.Interval)
	if err != nil {
		return ctrl.Result{}, nil
	}

	status := instance.Status.Rotation
	if status == nil || status.LastRotatedTime == nil {
		return ctrl.Result{RequeueAfter: interval}, nil
	}

	remaining := interval - time.Since(status.LastRotatedTime.Time)
	if remaining < 0 {
		remaining = 0
	}

	return ctrl.Result{RequeueAfter: remaining}, nil
}

// updateRotationStatus records the current spec.rotation.trigger value and
// the current time as the last rotation, after a password create or reset.
func (r *PostgresUserReconciler) updateRotationStatus(ctx context.Context, instance *postgresv1alpha1.PostgresUser) error {
	now := metav1.Now()
	status := &postgresv1alpha1.PostgresUserRotationStatus{LastRotatedTime: &now}
	if instance.Spec.Rotation != nil {
		status.Trigger = instance.Spec.Rotation.Trigger
	}
	instance.Status.Rotation = status
	if err := r.Status().Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update rotation status: %w", err)
	}
	return nil
}

// writeSecret creates or updates the Secret holding the connection details
// for the given user, owned by instance so it is garbage collected together
// with it.
func (r *PostgresUserReconciler) writeSecret(ctx context.Context, instance *postgresv1alpha1.PostgresUser, session *PostgresFlexSession, instanceID, secretName string, secretKeys postgresUserSecretKeys, username, password string) error {
	host, port, err := session.InstanceConnectionInfo(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to resolve instance connection info: %w", err)
	}

	secret := &corev1.Secret{}
	secret.Namespace = instance.Namespace
	secret.Name = secretName

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}
		secret.Data[secretKeys.Username] = []byte(username)
		secret.Data[secretKeys.Password] = []byte(password)
		secret.Data[secretKeys.Host] = []byte(host)
		secret.Data[secretKeys.Port] = []byte(strconv.Itoa(int(port)))
		return controllerutil.SetControllerReference(instance, secret, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("failed to write connection secret: %w", err)
	}

	if instance.Status.SecretName == nil || *instance.Status.SecretName != secretName {
		instance.Status.SecretName = &secretName
		if err := r.Status().Update(ctx, instance); err != nil {
			return fmt.Errorf("failed to update secret name status: %w", err)
		}
	}

	return nil
}

func (r *PostgresUserReconciler) handleDeletion(ctx context.Context, instance *postgresv1alpha1.PostgresUser, session *PostgresFlexSession, instanceID string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(instance, postgresUserFinalizer) {
		return ctrl.Result{}, nil
	}

	if instance.Status.UserID != nil {
		log.Info("Deleting user from Postgres Flex", "id", *instance.Status.UserID)
		if err := session.Client.DeleteUser(ctx, session.ProjectID, session.Region, instanceID, *instance.Status.UserID).Execute(); err != nil {
			var apiErr *oapierror.GenericOpenAPIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
				return ctrl.Result{}, fmt.Errorf("failed to delete user from Postgres Flex: %w", err)
			}
			log.Info("User already absent in Postgres Flex, skipping delete")
		}
	}

	return r.removeFinalizer(ctx, instance)
}

func (r *PostgresUserReconciler) removeFinalizer(ctx context.Context, instance *postgresv1alpha1.PostgresUser) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(instance, postgresUserFinalizer)
	if err := r.Update(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *PostgresUserReconciler) updateInstanceIDStatus(ctx context.Context, instance *postgresv1alpha1.PostgresUser, id *string) error {
	if instance.Status.InstanceID != nil && *instance.Status.InstanceID == *id {
		return nil
	}
	instance.Status.InstanceID = id
	if err := r.Status().Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance ID status: %w", err)
	}
	return nil
}

func (r *PostgresUserReconciler) updateUserIDStatus(ctx context.Context, instance *postgresv1alpha1.PostgresUser, id *int64) error {
	if instance.Status.UserID != nil && *instance.Status.UserID == *id {
		return nil
	}
	instance.Status.UserID = id
	if err := r.Status().Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update user ID status: %w", err)
	}
	return nil
}

// rolesEqual compares two role lists ignoring order.
func rolesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sortedA := slices.Clone(a)
	sortedB := slices.Clone(b)
	slices.Sort(sortedA)
	slices.Sort(sortedB)
	return slices.Equal(sortedA, sortedB)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&postgresv1alpha1.PostgresUser{}).
		Owns(&corev1.Secret{}).
		Named("postgres-postgresuser").
		Complete(r)
}
