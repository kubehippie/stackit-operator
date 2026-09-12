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
	"github.com/kubehippie/stackit-operator/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceAccountCredentialsSpec defines the desired state of ServiceAccountCredentials
type ServiceAccountCredentialsSpec struct {
	// projectId is the STACKIT project ID that resources referencing these
	// credentials are associated with.
	// +required
	ProjectID string `json:"projectId"`

	// region is the default STACKIT region used for regional services
	// (e.g. Postgres Flex) when a resource does not override it.
	// +optional
	Region *string `json:"region,omitempty"`

	// serviceAccountKey references a secret or direct value containing the
	// full JSON service account key (Key Flow), as generated via the STACKIT
	// Portal or the stackit_service_account_key resource. Either
	// serviceAccountKey or token must be set.
	// +optional
	ServiceAccountKey *common.SecretKeyRefOrVal `json:"serviceAccountKey,omitempty"`

	// privateKey references a secret or direct value containing a custom RSA
	// private key matching the public key uploaded for the service account
	// key. Only relevant when serviceAccountKey uses a custom key pair.
	// +optional
	PrivateKey *common.SecretKeyRefOrVal `json:"privateKey,omitempty"`

	// token references a secret or direct value containing a long-lived
	// STACKIT service account access token (Token Flow). This flow is
	// deprecated upstream but supported for simple setups. Either
	// serviceAccountKey or token must be set.
	// +optional
	Token *common.SecretKeyRefOrVal `json:"token,omitempty"`
}

// ServiceAccountCredentialsStatus defines the observed state of ServiceAccountCredentials.
type ServiceAccountCredentialsStatus struct {
	// connected shows if the STACKIT API was reached successfully using
	// these credentials.
	Connected bool `json:"connected"`

	// conditions represent the current state of the ServiceAccountCredentials resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ProjectID",type=string,JSONPath=`.spec.projectId`
// +kubebuilder:printcolumn:name="Connected",type=boolean,JSONPath=`.status.connected`

// ServiceAccountCredentials is the Schema for the serviceaccountcredentials API
type ServiceAccountCredentials struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ServiceAccountCredentials
	// +required
	Spec ServiceAccountCredentialsSpec `json:"spec"`

	// status defines the observed state of ServiceAccountCredentials
	// +optional
	Status ServiceAccountCredentialsStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ServiceAccountCredentialsList contains a list of ServiceAccountCredentials
type ServiceAccountCredentialsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ServiceAccountCredentials `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ServiceAccountCredentials{}, &ServiceAccountCredentialsList{})
		return nil
	})
}

// GetConnected returns whether the last reconciliation successfully reached
// the STACKIT API using these credentials.
func (s *ServiceAccountCredentials) GetConnected() bool { return s.Status.Connected }

// SetConnected records whether the last reconciliation successfully reached
// the STACKIT API using these credentials.
func (s *ServiceAccountCredentials) SetConnected(connected bool) { s.Status.Connected = connected }
