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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubehippie/stackit-operator/api/common"
	postgresv1alpha1 "github.com/kubehippie/stackit-operator/api/postgres/v1alpha1"
)

var _ = Describe("PostgresUser Webhook", func() {
	var (
		obj       *postgresv1alpha1.PostgresUser
		oldObj    *postgresv1alpha1.PostgresUser
		validator PostgresUserCustomValidator
		defaulter PostgresUserCustomDefaulter
	)

	BeforeEach(func() {
		obj = &postgresv1alpha1.PostgresUser{
			Spec: postgresv1alpha1.PostgresUserSpec{
				CredentialsRef: &common.StackitCredentialsRef{Name: testCredentialsName},
				InstanceName:   "test-instance",
				Username:       "test_user",
				Roles:          []string{"login"},
			},
		}
		obj.SetName("test-resource")
		oldObj = obj.DeepCopy()
		validator = PostgresUserCustomValidator{}
		defaulter = PostgresUserCustomDefaulter{}
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	Context("When creating PostgresUser under Defaulting Webhook", func() {
		It("Should default secret.name to the object name when unset", func() {
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(obj.Spec.Secret).NotTo(BeNil())
			Expect(obj.Spec.Secret.Name).NotTo(BeNil())
			Expect(*obj.Spec.Secret.Name).To(Equal("test-resource"))
		})

		It("Should not override an explicitly set secret.name", func() {
			secretName := "custom-secret"
			obj.Spec.Secret = &postgresv1alpha1.PostgresUserSecretSpec{Name: &secretName}
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(*obj.Spec.Secret.Name).To(Equal("custom-secret"))
		})
	})

	Context("When creating or updating PostgresUser under Validating Webhook", func() {
		It("Should deny creation when credentialsRef is missing", func() {
			obj.Spec.CredentialsRef = nil
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when instanceName is missing", func() {
			obj.Spec.InstanceName = ""
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when username is missing", func() {
			obj.Spec.Username = ""
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when roles is empty", func() {
			obj.Spec.Roles = nil
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should admit creation when all required fields are present", func() {
			Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
		})

		It("Should deny creation when custom secret keys collide", func() {
			username := "conn"
			password := "conn"
			obj.Spec.Secret = &postgresv1alpha1.PostgresUserSecretSpec{
				Keys: &postgresv1alpha1.PostgresUserSecretKeys{
					Username: &username,
					Password: &password,
				},
			}
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should admit creation with distinct custom secret keys", func() {
			username := "DB_USER"
			password := "DB_PASSWORD"
			host := "DB_HOST"
			port := "DB_PORT"
			obj.Spec.Secret = &postgresv1alpha1.PostgresUserSecretSpec{
				Keys: &postgresv1alpha1.PostgresUserSecretKeys{
					Username: &username,
					Password: &password,
					Host:     &host,
					Port:     &port,
				},
			}
			Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
		})

		It("Should deny creation when rotation.interval is not a valid duration", func() {
			interval := "not-a-duration"
			obj.Spec.Rotation = &postgresv1alpha1.PostgresUserRotationSpec{Interval: &interval}
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should admit creation when rotation.interval is a valid duration", func() {
			interval := "720h"
			obj.Spec.Rotation = &postgresv1alpha1.PostgresUserRotationSpec{Interval: &interval}
			Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
		})

		It("Should deny updates that change instanceName", func() {
			obj.Spec.InstanceName = "other-instance"
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().To(HaveOccurred())
		})

		It("Should deny updates that change username", func() {
			obj.Spec.Username = "other_user"
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().To(HaveOccurred())
		})

		It("Should admit valid updates", func() {
			obj.Spec.Roles = []string{"login", "createdb"}
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().NotTo(HaveOccurred())
		})
	})
})
