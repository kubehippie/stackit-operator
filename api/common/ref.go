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

// Package common holds shared types referenced across the different API
// groups of stackit-operator.
package common

// StackitCredentialsRef is a reference to the credentials used to
// authenticate against the STACKIT API. It points either to a namespaced
// ServiceAccountCredentials resource or to a cluster-scoped
// ClusterServiceAccountCredentials resource.
// +kubebuilder:object:generate=true
type StackitCredentialsRef struct {
	// kind specifies whether the referenced credentials are namespaced
	// (ServiceAccountCredentials) or cluster-scoped
	// (ClusterServiceAccountCredentials).
	// +kubebuilder:validation:Enum=ServiceAccountCredentials;ClusterServiceAccountCredentials
	// +kubebuilder:default=ServiceAccountCredentials
	// +optional
	Kind string `json:"kind,omitempty"`

	// name specifies the name of the referenced credentials resource.
	// +required
	Name string `json:"name,omitempty"`

	// namespace specifies the namespace of the referenced
	// ServiceAccountCredentials resource. It is ignored (and must be empty)
	// when kind is ClusterServiceAccountCredentials, and defaults to the
	// namespace of the referencing resource when kind is
	// ServiceAccountCredentials and namespace is left empty.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// SecretKeySelector selects a key of a Secret.
// +kubebuilder:object:generate=true
type SecretKeySelector struct {
	// name is the name of the Secret.
	// +required
	Name string `json:"name"`

	// key is the key of the Secret to select from.
	// +required
	Key string `json:"key"`

	// namespace of the Secret.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// SecretKeyRefOrVal holds either an inline value or a reference to a Secret
// key. Exactly one of value or secretKeyRef must be set.
// +kubebuilder:object:generate=true
type SecretKeyRefOrVal struct {
	// secretKeyRef selects a key of a Kubernetes Secret.
	// +optional
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`

	// value directly specifies the value as a plain string.
	// +optional
	Value string `json:"value,omitempty"`
}
