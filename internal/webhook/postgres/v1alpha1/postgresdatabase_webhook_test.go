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

const testCredentialsName = "test-credentials"

var _ = Describe("PostgresDatabase Webhook", func() {
	var (
		obj       *postgresv1alpha1.PostgresDatabase
		oldObj    *postgresv1alpha1.PostgresDatabase
		validator PostgresDatabaseCustomValidator
		defaulter PostgresDatabaseCustomDefaulter
	)

	BeforeEach(func() {
		obj = &postgresv1alpha1.PostgresDatabase{
			Spec: postgresv1alpha1.PostgresDatabaseSpec{
				CredentialsRef: &common.StackitCredentialsRef{Name: testCredentialsName},
				InstanceName:   "test-instance",
				Name:           "test_database",
				Owner:          "test_owner",
			},
		}
		oldObj = obj.DeepCopy()
		validator = PostgresDatabaseCustomValidator{}
		defaulter = PostgresDatabaseCustomDefaulter{}
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	Context("When creating PostgresDatabase under Defaulting Webhook", func() {
		It("Should not modify the object", func() {
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(obj.Spec.Name).To(Equal("test_database"))
		})
	})

	Context("When creating or updating PostgresDatabase under Validating Webhook", func() {
		It("Should deny creation when credentialsRef is missing", func() {
			obj.Spec.CredentialsRef = nil
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when credentialsRef.namespace is set for a cluster-scoped kind", func() {
			obj.Spec.CredentialsRef = &common.StackitCredentialsRef{
				Kind:      "ClusterServiceAccountCredentials",
				Name:      testCredentialsName,
				Namespace: "default",
			}
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when instanceName is missing", func() {
			obj.Spec.InstanceName = ""
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when owner is missing", func() {
			obj.Spec.Owner = ""
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should admit creation when all required fields are present", func() {
			Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
		})

		It("Should deny updates that change instanceName", func() {
			obj.Spec.InstanceName = "other-instance"
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().To(HaveOccurred())
		})

		It("Should deny updates that change name", func() {
			obj.Spec.Name = "other_database"
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().To(HaveOccurred())
		})

		It("Should admit valid updates", func() {
			obj.Spec.Owner = "other_owner"
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().NotTo(HaveOccurred())
		})
	})
})
