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

// PostgresDatabaseSpec defines the desired state of PostgresDatabase
type PostgresDatabaseSpec struct {
	// credentialsRef is a reference to the ServiceAccountCredentials or
	// ClusterServiceAccountCredentials resource used to authenticate
	// against the STACKIT API.
	// +required
	CredentialsRef *common.StackitCredentialsRef `json:"credentialsRef"`

	// instanceName is the name of the Postgres Flex instance the database
	// should be created in. The instance's STACKIT-assigned ID is resolved
	// automatically and stored in status.instanceID.
	// +required
	InstanceName string `json:"instanceName"`

	// name is the name of the database in Postgres Flex. Must be 1-63
	// characters long, start with a lowercase letter or underscore, and
	// contain only lowercase letters, numbers, underscores, or dashes. This
	// field is immutable after creation.
	// +required
	// +kubebuilder:validation:Pattern=`^[a-z_][a-z0-9_-]*$`
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// owner is the username of the database owner. This should typically
	// reference the username of a PostgresUser managed on the same
	// instance.
	// +required
	Owner string `json:"owner"`
}

// PostgresDatabaseStatus defines the observed state of PostgresDatabase.
type PostgresDatabaseStatus struct {
	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// instanceID is the STACKIT-assigned UUID of the Postgres Flex instance
	// resolved from spec.instanceName.
	// +optional
	InstanceID *string `json:"instanceID,omitempty"`

	// databaseID is the STACKIT-assigned ID of this database. It is stored
	// here so that update and delete operations can reference the database
	// directly without an additional lookup by name.
	// +optional
	DatabaseID *int64 `json:"databaseID,omitempty"`

	// conditions represent the current state of the PostgresDatabase resource.
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
// +kubebuilder:printcolumn:name="Database",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.owner`
// +kubebuilder:printcolumn:name="DatabaseID",type=string,JSONPath=`.status.databaseID`

// PostgresDatabase is the Schema for the postgresdatabases API
type PostgresDatabase struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PostgresDatabase
	// +required
	Spec PostgresDatabaseSpec `json:"spec"`

	// status defines the observed state of PostgresDatabase
	// +optional
	Status PostgresDatabaseStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PostgresDatabaseList contains a list of PostgresDatabase
type PostgresDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PostgresDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PostgresDatabase{}, &PostgresDatabaseList{})
		return nil
	})
}

// GetInstanceID returns the STACKIT-assigned instance UUID stored in the status.
func (p *PostgresDatabase) GetInstanceID() *string { return p.Status.InstanceID }

// SetInstanceID stores the STACKIT-assigned instance UUID in the status.
func (p *PostgresDatabase) SetInstanceID(id *string) { p.Status.InstanceID = id }
