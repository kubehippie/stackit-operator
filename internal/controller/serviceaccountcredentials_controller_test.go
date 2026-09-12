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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubehippie/stackit-operator/api/common"
	stackitoperatorwebhippiedev1alpha1 "github.com/kubehippie/stackit-operator/api/v1alpha1"
)

var _ = Describe("ServiceAccountCredentials Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			resourceNamespace = "default"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}
		serviceaccountcredentials := &stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ServiceAccountCredentials")
			err := k8sClient.Get(ctx, typeNamespacedName, serviceaccountcredentials)
			if err != nil && errors.IsNotFound(err) {
				resource := &stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: resourceNamespace,
					},
					Spec: stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentialsSpec{
						ProjectID: "00000000-0000-0000-0000-000000000000",
						Token:     &common.SecretKeyRefOrVal{Value: "test-token"},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ServiceAccountCredentials")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ServiceAccountCredentialsReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the connectivity status was recorded")
			updated := &stackitoperatorwebhippiedev1alpha1.ServiceAccountCredentials{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updated)).To(Succeed())
			Expect(updated.Status.Connected).To(BeFalse())
		})
	})
})
