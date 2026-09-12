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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubehippie/stackit-operator/api/common"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	stackitoperatorwebhippiedev1alpha1 "github.com/kubehippie/stackit-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var serviceaccountcredentialslog = logf.Log.WithName("serviceaccountcredentials-resource")

// SetupServiceAccountCredentialsWebhookWithManager registers the webhook for ServiceAccountCredentials in the manager.
func SetupServiceAccountCredentialsWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials{}).
		WithValidator(&ServiceAccountCredentialsCustomValidator{}).
		WithDefaulter(&ServiceAccountCredentialsCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-stackit-operator-webhippie-de-v1alpha1-serviceaccountcredentials,mutating=true,failurePolicy=fail,sideEffects=None,groups=stackit-operator.webhippie.de,resources=serviceaccountcredentials,verbs=create;update,versions=v1alpha1,name=mserviceaccountcredentials-v1alpha1.kb.io,admissionReviewVersions=v1

// ServiceAccountCredentialsCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ServiceAccountCredentials when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ServiceAccountCredentialsCustomDefaulter struct{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ServiceAccountCredentials.
func (d *ServiceAccountCredentialsCustomDefaulter) Default(_ context.Context, obj *stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials) error {
	_ = obj
	return nil
}

// +kubebuilder:webhook:path=/validate-stackit-operator-webhippie-de-v1alpha1-serviceaccountcredentials,mutating=false,failurePolicy=fail,sideEffects=None,groups=stackit-operator.webhippie.de,resources=serviceaccountcredentials,verbs=create;update,versions=v1alpha1,name=vserviceaccountcredentials-v1alpha1.kb.io,admissionReviewVersions=v1

// ServiceAccountCredentialsCustomValidator struct is responsible for validating the ServiceAccountCredentials resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ServiceAccountCredentialsCustomValidator struct{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ServiceAccountCredentials.
func (v *ServiceAccountCredentialsCustomValidator) ValidateCreate(_ context.Context, obj *stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials) (admission.Warnings, error) {
	serviceaccountcredentialslog.Info("Validation for ServiceAccountCredentials upon creation", "name", obj.GetName())

	return nil, validateStackitCredentialsSpec(obj.Spec.ProjectID, obj.Spec.ServiceAccountKey, obj.Spec.PrivateKey, obj.Spec.Token, obj.Namespace)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ServiceAccountCredentials.
func (v *ServiceAccountCredentialsCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials) (admission.Warnings, error) {
	serviceaccountcredentialslog.Info("Validation for ServiceAccountCredentials upon update", "name", newObj.GetName())

	if err := validateStackitCredentialsSpec(newObj.Spec.ProjectID, newObj.Spec.ServiceAccountKey, newObj.Spec.PrivateKey, newObj.Spec.Token, newObj.Namespace); err != nil {
		return nil, err
	}

	if oldObj.Spec.ProjectID != newObj.Spec.ProjectID {
		return nil, fmt.Errorf("spec.projectId is immutable and cannot be changed after creation")
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ServiceAccountCredentials.
func (v *ServiceAccountCredentialsCustomValidator) ValidateDelete(_ context.Context, obj *stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials) (admission.Warnings, error) {
	_ = obj
	return nil, nil
}

// validateStackitCredentialsSpec validates the fields shared by
// ServiceAccountCredentials and ClusterServiceAccountCredentials. namespace
// is the namespace the CR lives in, or empty for cluster-scoped resources,
// in which case any secretKeyRef must set an explicit namespace.
func validateStackitCredentialsSpec(projectID string, serviceAccountKey, privateKey, token *common.SecretKeyRefOrVal, namespace string) error {
	if strings.TrimSpace(projectID) == "" {
		return fmt.Errorf("spec.projectId must be set")
	}

	if (serviceAccountKey == nil) == (token == nil) {
		return fmt.Errorf("exactly one of spec.serviceAccountKey or spec.token must be set")
	}

	if serviceAccountKey != nil {
		if err := validateSecretKeyRefOrVal("spec.serviceAccountKey", serviceAccountKey, namespace); err != nil {
			return err
		}
	}

	if privateKey != nil {
		if err := validateSecretKeyRefOrVal("spec.privateKey", privateKey, namespace); err != nil {
			return err
		}
	}

	if token != nil {
		if err := validateSecretKeyRefOrVal("spec.token", token, namespace); err != nil {
			return err
		}
	}

	return nil
}

// validateSecretKeyRefOrVal validates a common.SecretKeyRefOrVal field.
// namespace is the namespace of the owning CR; when empty (cluster-scoped
// resources) any secretKeyRef must set an explicit namespace.
func validateSecretKeyRefOrVal(field string, ref *common.SecretKeyRefOrVal, namespace string) error {
	if strings.TrimSpace(ref.Value) != "" {
		return nil
	}

	if ref.SecretKeyRef == nil {
		return fmt.Errorf("%s must set value or secretKeyRef", field)
	}

	if strings.TrimSpace(ref.SecretKeyRef.Name) == "" || strings.TrimSpace(ref.SecretKeyRef.Key) == "" {
		return fmt.Errorf("%s.secretKeyRef must set name and key", field)
	}

	if namespace == "" && strings.TrimSpace(ref.SecretKeyRef.Namespace) == "" {
		return fmt.Errorf("%s.secretKeyRef.namespace must be set for cluster-scoped resources", field)
	}

	return nil
}
