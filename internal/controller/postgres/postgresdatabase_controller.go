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
	"time"

	postgresv1alpha1 "github.com/kubehippie/stackit-operator/api/postgres/v1alpha1"
	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
	postgresflex "github.com/stackitcloud/stackit-sdk-go/services/postgresflex/v3api"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const postgresDatabaseFinalizer = "stackit-operator.webhippie.de/postgresdatabase"

// PostgresDatabaseReconciler reconciles a PostgresDatabase object
type PostgresDatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=postgres.stackit-operator.webhippie.de,resources=postgresdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgres.stackit-operator.webhippie.de,resources=postgresdatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgres.stackit-operator.webhippie.de,resources=postgresdatabases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PostgresDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling")

	instance := &postgresv1alpha1.PostgresDatabase{}
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

	if !controllerutil.ContainsFinalizer(instance, postgresDatabaseFinalizer) {
		controllerutil.AddFinalizer(instance, postgresDatabaseFinalizer)
		if err := r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer: %w", err)
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	return r.reconcileDatabase(ctx, instance, session, instanceID)
}

func (r *PostgresDatabaseReconciler) reconcileDatabase(ctx context.Context, instance *postgresv1alpha1.PostgresDatabase, session *PostgresFlexSession, instanceID string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	existing, err := session.FindDatabaseByName(ctx, instanceID, instance.Spec.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	if existing == nil {
		resp, err := session.Client.CreateDatabase(ctx, session.ProjectID, session.Region, instanceID).
			CreateDatabasePayload(postgresflex.CreateDatabasePayload{
				Name:  instance.Spec.Name,
				Owner: &instance.Spec.Owner,
			}).Execute()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create database in Postgres Flex: %w", err)
		}
		log.Info("Database created in Postgres Flex", "id", resp.Id)
		if err := r.updateDatabaseIDStatus(ctx, instance, &resp.Id); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.updateDatabaseIDStatus(ctx, instance, &existing.Id); err != nil {
		return ctrl.Result{}, err
	}

	if existing.Owner != instance.Spec.Owner {
		if err := session.Client.UpdateDatabase(ctx, session.ProjectID, session.Region, instanceID, existing.Id).
			UpdateDatabasePayload(postgresflex.UpdateDatabasePayload{
				Name:  instance.Spec.Name,
				Owner: instance.Spec.Owner,
			}).Execute(); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update database in Postgres Flex: %w", err)
		}
		log.Info("Database owner updated in Postgres Flex")
	}

	return ctrl.Result{}, nil
}

func (r *PostgresDatabaseReconciler) handleDeletion(ctx context.Context, instance *postgresv1alpha1.PostgresDatabase, session *PostgresFlexSession, instanceID string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(instance, postgresDatabaseFinalizer) {
		return ctrl.Result{}, nil
	}

	if instance.Status.DatabaseID != nil {
		log.Info("Deleting database from Postgres Flex", "id", *instance.Status.DatabaseID)
		if err := session.Client.DeleteDatabase(ctx, session.ProjectID, session.Region, instanceID, *instance.Status.DatabaseID).Execute(); err != nil {
			var apiErr *oapierror.GenericOpenAPIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
				return ctrl.Result{}, fmt.Errorf("failed to delete database from Postgres Flex: %w", err)
			}
			log.Info("Database already absent in Postgres Flex, skipping delete")
		}
	}

	return r.removeFinalizer(ctx, instance)
}

func (r *PostgresDatabaseReconciler) removeFinalizer(ctx context.Context, instance *postgresv1alpha1.PostgresDatabase) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(instance, postgresDatabaseFinalizer)
	if err := r.Update(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *PostgresDatabaseReconciler) updateInstanceIDStatus(ctx context.Context, instance *postgresv1alpha1.PostgresDatabase, id *string) error {
	if instance.Status.InstanceID != nil && *instance.Status.InstanceID == *id {
		return nil
	}
	instance.Status.InstanceID = id
	if err := r.Status().Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance ID status: %w", err)
	}
	return nil
}

func (r *PostgresDatabaseReconciler) updateDatabaseIDStatus(ctx context.Context, instance *postgresv1alpha1.PostgresDatabase, id *int64) error {
	if instance.Status.DatabaseID != nil && *instance.Status.DatabaseID == *id {
		return nil
	}
	instance.Status.DatabaseID = id
	if err := r.Status().Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update database ID status: %w", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&postgresv1alpha1.PostgresDatabase{}).
		Named("postgres-postgresdatabase").
		Complete(r)
}
