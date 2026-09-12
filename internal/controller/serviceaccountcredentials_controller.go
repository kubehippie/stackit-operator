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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/kubehippie/stackit-operator/api/common"
	v1alpha1 "github.com/kubehippie/stackit-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// FailedStackitConnectionRetryPeriod is the requeue delay after a failed
	// connectivity check against the STACKIT API.
	FailedStackitConnectionRetryPeriod = time.Second * 30

	// successStackitConnectionRetryPeriod is the requeue delay after a
	// successful connectivity check, so credential rotation or expiry is
	// still detected periodically.
	successStackitConnectionRetryPeriod = time.Minute * 30
)

// ServiceAccountCredentialsReconciler reconciles a ServiceAccountCredentials object
type ServiceAccountCredentialsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=stackit-operator.webhippie.de,resources=serviceaccountcredentials,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stackit-operator.webhippie.de,resources=serviceaccountcredentials/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stackit-operator.webhippie.de,resources=serviceaccountcredentials/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ServiceAccountCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling")

	instance := &v1alpha1.ServiceAccountCredentials{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to fetch: %w", err)
	}

	connected := true
	if err := r.verifyConnectivity(ctx, instance); err != nil {
		log.Error(err, "Unable to reach STACKIT API")
		connected = false
	}

	if instance.Status.Connected != connected {
		instance.Status.Connected = connected
		if err := r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
		}
	}

	if !connected {
		return ctrl.Result{RequeueAfter: FailedStackitConnectionRetryPeriod}, nil
	}

	return ctrl.Result{RequeueAfter: successStackitConnectionRetryPeriod}, nil
}

func (r *ServiceAccountCredentialsReconciler) verifyConnectivity(ctx context.Context, instance *v1alpha1.ServiceAccountCredentials) error {
	creds, err := ResolveStackitCredentialsRef(ctx, r.Client, &common.StackitCredentialsRef{
		Kind:      "ServiceAccountCredentials",
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}, instance.Namespace)
	if err != nil {
		return err
	}

	return VerifyStackitConnectivity(ctx, creds)
}

const serviceAccountCredentialsSecretIndexField = ".spec.secretRefs"

// serviceAccountCredentialsSecretRefKeys returns the RefIndexKey values for
// every Secret the given ServiceAccountCredentials instance may reference.
func serviceAccountCredentialsSecretRefKeys(obj client.Object) []string {
	creds, ok := obj.(*v1alpha1.ServiceAccountCredentials)
	if !ok {
		return nil
	}

	var keys []string
	appendSecretRef := func(ref *common.SecretKeyRefOrVal) {
		if ref == nil || ref.SecretKeyRef == nil {
			return
		}
		ns := ref.SecretKeyRef.Namespace
		if ns == "" {
			ns = creds.Namespace
		}
		keys = append(keys, RefIndexKey(ns, ref.SecretKeyRef.Name))
	}

	appendSecretRef(creds.Spec.ServiceAccountKey)
	appendSecretRef(creds.Spec.PrivateKey)
	appendSecretRef(creds.Spec.Token)

	return keys
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceAccountCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := RegisterRefIndex(mgr, &v1alpha1.ServiceAccountCredentials{}, serviceAccountCredentialsSecretIndexField, serviceAccountCredentialsSecretRefKeys); err != nil {
		return err
	}

	newList := func() client.ObjectList { return &v1alpha1.ServiceAccountCredentialsList{} }

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ServiceAccountCredentials{}).
		Watches(&corev1.Secret{}, RefEventHandler(mgr.GetClient(), newList, serviceAccountCredentialsSecretIndexField)).
		Named("serviceaccountcredentials").
		Complete(r)
}
