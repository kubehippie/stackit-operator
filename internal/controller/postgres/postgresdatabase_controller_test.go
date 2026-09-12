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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubehippie/stackit-operator/api/common"
	postgresv1alpha1 "github.com/kubehippie/stackit-operator/api/postgres/v1alpha1"
	v1alpha1 "github.com/kubehippie/stackit-operator/api/v1alpha1"
)

var _ = Describe("PostgresDatabase Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			credentialsName   = "test-credentials"
			resourceNamespace = "default"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}
		credentialsNamespacedName := types.NamespacedName{
			Name:      credentialsName,
			Namespace: resourceNamespace,
		}
		postgresdatabase := &postgresv1alpha1.PostgresDatabase{}

		BeforeEach(func() {
			By("creating the referenced ServiceAccountCredentials")
			credentials := &v1alpha1.ServiceAccountCredentials{}
			err := k8sClient.Get(ctx, credentialsNamespacedName, credentials)
			if err != nil && errors.IsNotFound(err) {
				credentials = &v1alpha1.ServiceAccountCredentials{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialsName,
						Namespace: resourceNamespace,
					},
					Spec: v1alpha1.ServiceAccountCredentialsSpec{
						ProjectID: "00000000-0000-0000-0000-000000000000",
						Token:     &common.SecretKeyRefOrVal{Value: "test-token"},
					},
				}
				Expect(k8sClient.Create(ctx, credentials)).To(Succeed())
			}

			By("creating the custom resource for the Kind PostgresDatabase")
			err = k8sClient.Get(ctx, typeNamespacedName, postgresdatabase)
			if err != nil && errors.IsNotFound(err) {
				resource := &postgresv1alpha1.PostgresDatabase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: resourceNamespace,
					},
					Spec: postgresv1alpha1.PostgresDatabaseSpec{
						CredentialsRef: &common.StackitCredentialsRef{
							Name: credentialsName,
						},
						InstanceName: "test-instance",
						Name:         "test_database",
						Owner:        "test_owner",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &postgresv1alpha1.PostgresDatabase{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance PostgresDatabase")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			credentials := &v1alpha1.ServiceAccountCredentials{}
			Expect(k8sClient.Get(ctx, credentialsNamespacedName, credentials)).To(Succeed())
			Expect(k8sClient.Delete(ctx, credentials)).To(Succeed())
		})

		It("should reconcile the resource and fail to resolve the instance ID", func() {
			By("Reconciling the created resource")
			controllerReconciler := &PostgresDatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// No real Postgres Flex instance named "test-instance" exists
			// under the fake credentials, so reconciliation is expected to
			// fail while trying to resolve the instance ID.
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
		})
	})
})
