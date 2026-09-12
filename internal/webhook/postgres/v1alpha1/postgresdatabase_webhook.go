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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	postgresv1alpha1 "github.com/kubehippie/stackit-operator/api/postgres/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var postgresdatabaselog = logf.Log.WithName("postgresdatabase-resource")

// SetupPostgresDatabaseWebhookWithManager registers the webhook for PostgresDatabase in the manager.
func SetupPostgresDatabaseWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &postgresv1alpha1.PostgresDatabase{}).
		WithValidator(&PostgresDatabaseCustomValidator{}).
		WithDefaulter(&PostgresDatabaseCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-postgres-stackit-operator-webhippie-de-v1alpha1-postgresdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=postgres.stackit-operator.webhippie.de,resources=postgresdatabases,verbs=create;update,versions=v1alpha1,name=mpostgresdatabase-v1alpha1.kb.io,admissionReviewVersions=v1

// PostgresDatabaseCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind PostgresDatabase when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PostgresDatabaseCustomDefaulter struct{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind PostgresDatabase.
func (d *PostgresDatabaseCustomDefaulter) Default(_ context.Context, obj *postgresv1alpha1.PostgresDatabase) error {
	_ = obj
	return nil
}

// +kubebuilder:webhook:path=/validate-postgres-stackit-operator-webhippie-de-v1alpha1-postgresdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=postgres.stackit-operator.webhippie.de,resources=postgresdatabases,verbs=create;update,versions=v1alpha1,name=vpostgresdatabase-v1alpha1.kb.io,admissionReviewVersions=v1

// PostgresDatabaseCustomValidator struct is responsible for validating the PostgresDatabase resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type PostgresDatabaseCustomValidator struct{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type PostgresDatabase.
func (v *PostgresDatabaseCustomValidator) ValidateCreate(_ context.Context, obj *postgresv1alpha1.PostgresDatabase) (admission.Warnings, error) {
	postgresdatabaselog.Info("Validation for PostgresDatabase upon creation", "name", obj.GetName())

	return nil, v.validate(obj)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type PostgresDatabase.
func (v *PostgresDatabaseCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *postgresv1alpha1.PostgresDatabase) (admission.Warnings, error) {
	postgresdatabaselog.Info("Validation for PostgresDatabase upon update", "name", newObj.GetName())

	if err := v.validate(newObj); err != nil {
		return nil, err
	}

	if oldObj.Spec.InstanceName != newObj.Spec.InstanceName {
		return nil, fmt.Errorf("spec.instanceName is immutable and cannot be changed after creation")
	}

	if oldObj.Spec.Name != newObj.Spec.Name {
		return nil, fmt.Errorf("spec.name is immutable and cannot be changed after creation")
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type PostgresDatabase.
func (v *PostgresDatabaseCustomValidator) ValidateDelete(_ context.Context, obj *postgresv1alpha1.PostgresDatabase) (admission.Warnings, error) {
	_ = obj
	return nil, nil
}

func (v *PostgresDatabaseCustomValidator) validate(obj *postgresv1alpha1.PostgresDatabase) error {
	if err := validateCredentialsRef(obj.Spec.CredentialsRef); err != nil {
		return err
	}

	if strings.TrimSpace(obj.Spec.InstanceName) == "" {
		return fmt.Errorf("spec.instanceName must be set")
	}

	if strings.TrimSpace(obj.Spec.Name) == "" {
		return fmt.Errorf("spec.name must be set")
	}

	if strings.TrimSpace(obj.Spec.Owner) == "" {
		return fmt.Errorf("spec.owner must be set")
	}

	return nil
}
