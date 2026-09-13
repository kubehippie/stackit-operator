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
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	postgresv1alpha1 "github.com/kubehippie/stackit-operator/api/postgres/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var postgresuserlog = logf.Log.WithName("postgresuser-resource")

// SetupPostgresUserWebhookWithManager registers the webhook for PostgresUser in the manager.
func SetupPostgresUserWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &postgresv1alpha1.PostgresUser{}).
		WithValidator(&PostgresUserCustomValidator{}).
		WithDefaulter(&PostgresUserCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-postgres-stackit-operator-webhippie-de-v1alpha1-postgresuser,mutating=true,failurePolicy=fail,sideEffects=None,groups=postgres.stackit-operator.webhippie.de,resources=postgresusers,verbs=create;update,versions=v1alpha1,name=mpostgresuser-v1alpha1.kb.io,admissionReviewVersions=v1

// PostgresUserCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind PostgresUser when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PostgresUserCustomDefaulter struct{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind PostgresUser.
func (d *PostgresUserCustomDefaulter) Default(_ context.Context, obj *postgresv1alpha1.PostgresUser) error {
	if obj.Spec.Secret == nil {
		obj.Spec.Secret = &postgresv1alpha1.PostgresUserSecretSpec{}
	}

	if obj.Spec.Secret.Name == nil || strings.TrimSpace(*obj.Spec.Secret.Name) == "" {
		postgresuserlog.Info("Defaulting for PostgresUser", "name", obj.GetName())
		name := obj.GetName()
		obj.Spec.Secret.Name = &name
	}

	return nil
}

// +kubebuilder:webhook:path=/validate-postgres-stackit-operator-webhippie-de-v1alpha1-postgresuser,mutating=false,failurePolicy=fail,sideEffects=None,groups=postgres.stackit-operator.webhippie.de,resources=postgresusers,verbs=create;update,versions=v1alpha1,name=vpostgresuser-v1alpha1.kb.io,admissionReviewVersions=v1

// PostgresUserCustomValidator struct is responsible for validating the PostgresUser resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type PostgresUserCustomValidator struct{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type PostgresUser.
func (v *PostgresUserCustomValidator) ValidateCreate(_ context.Context, obj *postgresv1alpha1.PostgresUser) (admission.Warnings, error) {
	postgresuserlog.Info("Validation for PostgresUser upon creation", "name", obj.GetName())

	return nil, v.validate(obj)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type PostgresUser.
func (v *PostgresUserCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *postgresv1alpha1.PostgresUser) (admission.Warnings, error) {
	postgresuserlog.Info("Validation for PostgresUser upon update", "name", newObj.GetName())

	if err := v.validate(newObj); err != nil {
		return nil, err
	}

	if oldObj.Spec.InstanceName != newObj.Spec.InstanceName {
		return nil, fmt.Errorf("spec.instanceName is immutable and cannot be changed after creation")
	}

	if oldObj.Spec.Username != newObj.Spec.Username {
		return nil, fmt.Errorf("spec.username is immutable and cannot be changed after creation")
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type PostgresUser.
func (v *PostgresUserCustomValidator) ValidateDelete(_ context.Context, obj *postgresv1alpha1.PostgresUser) (admission.Warnings, error) {
	_ = obj
	return nil, nil
}

func (v *PostgresUserCustomValidator) validate(obj *postgresv1alpha1.PostgresUser) error {
	if err := validateCredentialsRef(obj.Spec.CredentialsRef); err != nil {
		return err
	}

	if strings.TrimSpace(obj.Spec.InstanceName) == "" {
		return fmt.Errorf("spec.instanceName must be set")
	}

	if strings.TrimSpace(obj.Spec.Username) == "" {
		return fmt.Errorf("spec.username must be set")
	}

	if len(obj.Spec.Roles) == 0 {
		return fmt.Errorf("spec.roles must set at least one role")
	}

	if err := validateSecretKeys(obj.Spec.Secret); err != nil {
		return err
	}

	if err := validateRotation(obj.Spec.Rotation); err != nil {
		return err
	}

	return nil
}

// validateSecretKeys ensures that, if custom secret key names are set, they
// do not collide with one another.
func validateSecretKeys(secret *postgresv1alpha1.PostgresUserSecretSpec) error {
	if secret == nil || secret.Keys == nil {
		return nil
	}

	keys := secret.Keys
	names := map[string]string{
		"username": "username",
		"password": "password",
		"host":     "host",
		"port":     "port",
	}
	if keys.Username != nil && strings.TrimSpace(*keys.Username) != "" {
		names["username"] = *keys.Username
	}
	if keys.Password != nil && strings.TrimSpace(*keys.Password) != "" {
		names["password"] = *keys.Password
	}
	if keys.Host != nil && strings.TrimSpace(*keys.Host) != "" {
		names["host"] = *keys.Host
	}
	if keys.Port != nil && strings.TrimSpace(*keys.Port) != "" {
		names["port"] = *keys.Port
	}

	seen := map[string]string{}
	for field, key := range names {
		if other, ok := seen[key]; ok {
			return fmt.Errorf("spec.secret.keys.%s and spec.secret.keys.%s must not use the same key %q", other, field, key)
		}
		seen[key] = field
	}

	return nil
}

// validateRotation ensures spec.rotation.interval, when set, is a valid Go
// duration string.
func validateRotation(rotation *postgresv1alpha1.PostgresUserRotationSpec) error {
	if rotation == nil || rotation.Interval == nil {
		return nil
	}

	if _, err := time.ParseDuration(*rotation.Interval); err != nil {
		return fmt.Errorf("spec.rotation.interval must be a valid duration: %w", err)
	}

	return nil
}
