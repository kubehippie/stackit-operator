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

	// secret configures the Secret this operator creates (or updates) with
	// the connection details (username, password, host, port) for this
	// user.
	// +optional
	Secret *PostgresUserSecretSpec `json:"secret,omitempty"`

	// rotation configures automatic and/or triggered rotation of this
	// user's password, mirroring the `rotate_when_changed` mechanism of the
	// STACKIT Terraform provider.
	// +optional
	Rotation *PostgresUserRotationSpec `json:"rotation,omitempty"`
}

// PostgresUserSecretKeys allows overriding the individual keys written to
// the generated connection Secret, e.g. to directly match the environment
// variable names required by a consuming application. Any key left unset
// falls back to its default name.
type PostgresUserSecretKeys struct {
	// username overrides the key used for the username. Defaults to "username".
	// +optional
	Username *string `json:"username,omitempty"`

	// password overrides the key used for the password. Defaults to "password".
	// +optional
	Password *string `json:"password,omitempty"`

	// host overrides the key used for the host. Defaults to "host".
	// +optional
	Host *string `json:"host,omitempty"`

	// port overrides the key used for the port. Defaults to "port".
	// +optional
	Port *string `json:"port,omitempty"`
}

// PostgresUserSecretSpec configures the Secret this operator writes with
// the connection details for this user.
type PostgresUserSecretSpec struct {
	// name specifies the name of the Secret this operator creates (or
	// updates) with the connection details for this user. Defaults to the
	// name of this resource.
	// +optional
	Name *string `json:"name,omitempty"`

	// keys allows renaming the individual keys written to the Secret. Any
	// key left unset uses its default name.
	// +optional
	Keys *PostgresUserSecretKeys `json:"keys,omitempty"`
}

// PostgresUserRotationSpec configures automatic and/or triggered password
// rotation for a PostgresUser.
type PostgresUserRotationSpec struct {
	// trigger is an arbitrary, user-controlled value. Whenever it changes
	// compared to the last reconciled value (recorded in
	// status.rotation.trigger), the operator resets the user's password in
	// STACKIT and rewrites the connection Secret. Use this to force
	// rotation on demand, e.g. by bumping a timestamp or random value from
	// a CronJob or GitOps pipeline.
	// +optional
	Trigger *string `json:"trigger,omitempty"`

	// interval, when set, causes the operator to automatically reset the
	// password on this fixed schedule (e.g. "720h" for 30 days), in
	// addition to any trigger-based rotation. Must be a valid Go duration
	// string.
	// +optional
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	Interval *string `json:"interval,omitempty"`
}

// PostgresUserRotationStatus records the last observed password rotation
// state for a PostgresUser.
type PostgresUserRotationStatus struct {
	// trigger records the last spec.rotation.trigger value that was
	// reconciled, used to detect subsequent changes.
	// +optional
	Trigger *string `json:"trigger,omitempty"`

	// lastRotatedTime records when the password was last rotated by this
	// operator, either due to a trigger change, an elapsed interval, or
	// initial user creation.
	// +optional
	LastRotatedTime *metav1.Time `json:"lastRotatedTime,omitempty"`
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

	// rotation records the last observed password rotation state.
	// +optional
	Rotation *PostgresUserRotationStatus `json:"rotation,omitempty"`

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
