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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	stackitoperatorwebhippiedev1alpha1 "github.com/kubehippie/stackit-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var clusterserviceaccountcredentialslog = logf.Log.WithName("clusterserviceaccountcredentials-resource")

// SetupClusterServiceAccountCredentialsWebhookWithManager registers the webhook for ClusterServiceAccountCredentials in the manager.
func SetupClusterServiceAccountCredentialsWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials{}).
		WithValidator(&ClusterServiceAccountCredentialsCustomValidator{}).
		WithDefaulter(&ClusterServiceAccountCredentialsCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-stackit-operator-webhippie-de-v1alpha1-clusterserviceaccountcredentials,mutating=true,failurePolicy=fail,sideEffects=None,groups=stackit-operator.webhippie.de,resources=clusterserviceaccountcredentials,verbs=create;update,versions=v1alpha1,name=mclusterserviceaccountcredentials-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterServiceAccountCredentialsCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ClusterServiceAccountCredentials when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ClusterServiceAccountCredentialsCustomDefaulter struct{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ClusterServiceAccountCredentials.
func (d *ClusterServiceAccountCredentialsCustomDefaulter) Default(_ context.Context, obj *stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials) error {
	_ = obj
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/validate-stackit-operator-webhippie-de-v1alpha1-clusterserviceaccountcredentials,mutating=false,failurePolicy=fail,sideEffects=None,groups=stackit-operator.webhippie.de,resources=clusterserviceaccountcredentials,verbs=create;update,versions=v1alpha1,name=vclusterserviceaccountcredentials-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterServiceAccountCredentialsCustomValidator struct is responsible for validating the ClusterServiceAccountCredentials resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ClusterServiceAccountCredentialsCustomValidator struct{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ClusterServiceAccountCredentials.
func (v *ClusterServiceAccountCredentialsCustomValidator) ValidateCreate(_ context.Context, obj *stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials) (admission.Warnings, error) {
	clusterserviceaccountcredentialslog.Info("Validation for ClusterServiceAccountCredentials upon creation", "name", obj.GetName())

	return nil, validateStackitCredentialsSpec(obj.Spec.ProjectID, obj.Spec.ServiceAccountKey, obj.Spec.PrivateKey, obj.Spec.Token, "")
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ClusterServiceAccountCredentials.
func (v *ClusterServiceAccountCredentialsCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials) (admission.Warnings, error) {
	clusterserviceaccountcredentialslog.Info("Validation for ClusterServiceAccountCredentials upon update", "name", newObj.GetName())

	if err := validateStackitCredentialsSpec(newObj.Spec.ProjectID, newObj.Spec.ServiceAccountKey, newObj.Spec.PrivateKey, newObj.Spec.Token, ""); err != nil {
		return nil, err
	}

	if oldObj.Spec.ProjectID != newObj.Spec.ProjectID {
		return nil, fmt.Errorf("spec.projectId is immutable and cannot be changed after creation")
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ClusterServiceAccountCredentials.
func (v *ClusterServiceAccountCredentialsCustomValidator) ValidateDelete(_ context.Context, obj *stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials) (admission.Warnings, error) {
	_ = obj
	return nil, nil
}
