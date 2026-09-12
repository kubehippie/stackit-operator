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
	stackitoperatorwebhippiedev1alpha1 "github.com/kubehippie/stackit-operator/api/v1alpha1"
)

const testSecretName = "test-secret"

var _ = Describe("ClusterServiceAccountCredentials Webhook", func() {
	var (
		obj       *stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials
		oldObj    *stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials
		validator ClusterServiceAccountCredentialsCustomValidator
		defaulter ClusterServiceAccountCredentialsCustomDefaulter
	)

	BeforeEach(func() {
		obj = &stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentials{
			Spec: stackitoperatorwebhippiedev1alpha1.ClusterServiceAccountCredentialsSpec{
				ProjectID: "00000000-0000-0000-0000-000000000000",
				Token:     &common.SecretKeyRefOrVal{Value: "test-token"},
			},
		}
		oldObj = obj.DeepCopy()
		validator = ClusterServiceAccountCredentialsCustomValidator{}
		defaulter = ClusterServiceAccountCredentialsCustomDefaulter{}
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	Context("When creating ClusterServiceAccountCredentials under Defaulting Webhook", func() {
		It("Should not modify the object", func() {
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			Expect(obj.Spec.ProjectID).To(Equal("00000000-0000-0000-0000-000000000000"))
		})
	})

	Context("When creating or updating ClusterServiceAccountCredentials under Validating Webhook", func() {
		It("Should deny creation when projectId is missing", func() {
			obj.Spec.ProjectID = ""
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when neither serviceAccountKey nor token is set", func() {
			obj.Spec.Token = nil
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should deny creation when a secretKeyRef has no explicit namespace", func() {
			obj.Spec.Token = &common.SecretKeyRefOrVal{
				SecretKeyRef: &common.SecretKeySelector{Name: testSecretName, Key: "token"},
			}
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should admit creation when a secretKeyRef sets an explicit namespace", func() {
			obj.Spec.Token = &common.SecretKeyRefOrVal{
				SecretKeyRef: &common.SecretKeySelector{Name: testSecretName, Key: "token", Namespace: "default"},
			}
			Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
		})

		It("Should admit creation when all required fields are present", func() {
			Expect(validator.ValidateCreate(ctx, obj)).Error().NotTo(HaveOccurred())
		})

		It("Should deny updates that change projectId", func() {
			obj.Spec.ProjectID = "11111111-1111-1111-1111-111111111111"
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().To(HaveOccurred())
		})

		It("Should admit valid updates", func() {
			obj.Spec.Token = &common.SecretKeyRefOrVal{Value: "rotated-token"}
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().NotTo(HaveOccurred())
		})
	})
})
