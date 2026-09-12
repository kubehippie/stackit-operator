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

// PostgresUserSpec defines the desired state of PostgresUser
type PostgresUserSpec struct {
	// credentialsRef is a reference to the ServiceAccountCredentials or
	// ClusterServiceAccountCredentials resource used to authenticate
	// against the STACKIT API.
	// +required
	CredentialsRef *common.StackitCredentialsRef `json:"credentialsRef"`

	// instanceName is the name of the Postgres Flex instance the user
	// should be created on. The instance's STACKIT-assigned ID is resolved
	// automatically and stored in status.instanceID.
	// +required
	InstanceName string `json:"instanceName"`

	// username is the login name of the user. This field is immutable
	// after creation.
	// +required
	Username string `json:"username"`

	// roles lists the database access levels granted to the user (e.g.
	// "login", "createdb", "read", "readwrite"). Refer to the Postgres Flex
	// List Roles endpoint for the roles available on a given instance.
	// +required
	// +kubebuilder:validation:MinItems=1
	Roles []string `json:"roles"`

	// secretName specifies the name of the Secret this operator creates (or
	// updates) with the connection details (username, password, host, port,
	// database) for this user. Defaults to the name of this resource.
	// +optional
	SecretName *string `json:"secretName,omitempty"`
}

// PostgresUserStatus defines the observed state of PostgresUser.
type PostgresUserStatus struct {
	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// instanceID is the STACKIT-assigned UUID of the Postgres Flex instance
	// resolved from spec.instanceName.
	// +optional
	InstanceID *string `json:"instanceID,omitempty"`

	// userID is the STACKIT-assigned ID of this user. It is stored here so
	// that update and delete operations can reference the user directly
	// without an additional lookup by username.
	// +optional
	UserID *int64 `json:"userID,omitempty"`

	// secretName records the name of the Secret this operator most recently
	// wrote the connection details to.
	// +optional
	SecretName *string `json:"secretName,omitempty"`

	// conditions represent the current state of the PostgresUser resource.
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
// +kubebuilder:printcolumn:name="Instance",type=string,JSONPath=`.spec.instanceName`
// +kubebuilder:printcolumn:name="Username",type=string,JSONPath=`.spec.username`
// +kubebuilder:printcolumn:name="UserID",type=string,JSONPath=`.status.userID`

// PostgresUser is the Schema for the postgresusers API
type PostgresUser struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PostgresUser
	// +required
	Spec PostgresUserSpec `json:"spec"`

	// status defines the observed state of PostgresUser
	// +optional
	Status PostgresUserStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PostgresUserList contains a list of PostgresUser
type PostgresUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PostgresUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PostgresUser{}, &PostgresUserList{})
		return nil
	})
}

// GetInstanceID returns the STACKIT-assigned instance UUID stored in the status.
func (p *PostgresUser) GetInstanceID() *string { return p.Status.InstanceID }

// SetInstanceID stores the STACKIT-assigned instance UUID in the status.
func (p *PostgresUser) SetInstanceID(id *string) { p.Status.InstanceID = id }
