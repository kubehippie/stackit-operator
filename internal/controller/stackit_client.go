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
	"fmt"

	"github.com/kubehippie/stackit-operator/api/common"
	v1alpha1 "github.com/kubehippie/stackit-operator/api/v1alpha1"
	stackitconfig "github.com/stackitcloud/stackit-sdk-go/core/config"
	resourcemanager "github.com/stackitcloud/stackit-sdk-go/services/resourcemanager/v0api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StackitCredentials holds the resolved STACKIT API configuration derived
// from a ServiceAccountCredentials or ClusterServiceAccountCredentials
// resource.
type StackitCredentials struct {
	// Options are the SDK authentication configuration options to pass to
	// any STACKIT service API client constructor. Deliberately excludes the
	// region: not every STACKIT API accepts a region in its client
	// configuration (e.g. the global Resource Manager API rejects it
	// outright), so callers that need a region should append
	// stackitconfig.WithRegion(creds.Region) themselves when constructing a
	// region-scoped client.
	Options []stackitconfig.ConfigurationOption

	// ProjectID is the STACKIT project ID configured on the referenced
	// credentials resource.
	ProjectID string

	// Region is the STACKIT region configured on the referenced credentials
	// resource, or the empty string when unset.
	Region string
}

// ResolveStackitCredentialsRef resolves a StackitCredentialsRef to the
// STACKIT API configuration options, following either a namespaced
// ServiceAccountCredentials or a cluster-scoped
// ClusterServiceAccountCredentials resource. defaultNamespace is used when
// ref.Namespace is empty and ref.Kind is ServiceAccountCredentials (or left
// empty).
func ResolveStackitCredentialsRef(ctx context.Context, c client.Client, ref *common.StackitCredentialsRef, defaultNamespace string) (*StackitCredentials, error) {
	if ref == nil {
		return nil, fmt.Errorf("credentialsRef must be set")
	}

	switch ref.Kind {
	case "", "ServiceAccountCredentials":
		ns := ref.Namespace
		if ns == "" {
			ns = defaultNamespace
		}

		creds := &v1alpha1.ServiceAccountCredentials{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: ref.Name}, creds); err != nil {
			return nil, fmt.Errorf("unable to fetch ServiceAccountCredentials %s/%s: %w", ns, ref.Name, err)
		}

		return resolveStackitCredentials(ctx, c, creds.Spec.ProjectID, creds.Spec.Region,
			creds.Spec.ServiceAccountKey, creds.Spec.PrivateKey, creds.Spec.Token, ns)

	case "ClusterServiceAccountCredentials":
		creds := &v1alpha1.ClusterServiceAccountCredentials{}
		if err := c.Get(ctx, types.NamespacedName{Name: ref.Name}, creds); err != nil {
			return nil, fmt.Errorf("unable to fetch ClusterServiceAccountCredentials %s: %w", ref.Name, err)
		}

		return resolveStackitCredentials(ctx, c, creds.Spec.ProjectID, creds.Spec.Region,
			creds.Spec.ServiceAccountKey, creds.Spec.PrivateKey, creds.Spec.Token, "")

	default:
		return nil, fmt.Errorf("unsupported credentialsRef kind %q", ref.Kind)
	}
}

// resolveStackitCredentials builds the SDK configuration options common to
// both the namespaced and cluster-scoped credentials resources.
// defaultNamespace is used to resolve secretKeyRefs that don't carry an
// explicit namespace; it is empty for cluster-scoped credentials, in which
// case every secretKeyRef must set its own namespace.
func resolveStackitCredentials(ctx context.Context, c client.Client, projectID string, region *string,
	serviceAccountKey, privateKey, token *common.SecretKeyRefOrVal, defaultNamespace string) (*StackitCredentials, error) {
	var opts []stackitconfig.ConfigurationOption

	switch {
	case serviceAccountKey != nil:
		key, err := ResolveSecretKeyRefOrVal(ctx, c, serviceAccountKey, defaultNamespace)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve serviceAccountKey: %w", err)
		}
		opts = append(opts, stackitconfig.WithServiceAccountKey(key))

		if privateKey != nil {
			pk, err := ResolveSecretKeyRefOrVal(ctx, c, privateKey, defaultNamespace)
			if err != nil {
				return nil, fmt.Errorf("unable to resolve privateKey: %w", err)
			}
			opts = append(opts, stackitconfig.WithPrivateKey(pk))
		}

	case token != nil:
		t, err := ResolveSecretKeyRefOrVal(ctx, c, token, defaultNamespace)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve token: %w", err)
		}
		opts = append(opts, stackitconfig.WithToken(t))

	default:
		return nil, fmt.Errorf("either serviceAccountKey or token must be set")
	}

	regionVal := ""
	if region != nil {
		regionVal = *region
	}

	return &StackitCredentials{Options: opts, ProjectID: projectID, Region: regionVal}, nil
}

// ResolveSecretKeyRefOrVal returns the string value from a SecretKeyRefOrVal.
// If the inline Value field is non-empty it is returned directly; otherwise
// the value is read from the referenced Kubernetes Secret. defaultNamespace
// is used when the secretKeyRef carries no explicit namespace; if
// defaultNamespace is also empty, an explicit namespace is required (this is
// the case for cluster-scoped credentials resources).
func ResolveSecretKeyRefOrVal(ctx context.Context, c client.Client, ref *common.SecretKeyRefOrVal, defaultNamespace string) (string, error) {
	if ref == nil {
		return "", fmt.Errorf("ref is nil")
	}

	if ref.Value != "" {
		return ref.Value, nil
	}

	if ref.SecretKeyRef == nil {
		return "", fmt.Errorf("either value or secretKeyRef must be set")
	}

	if ref.SecretKeyRef.Name == "" {
		return "", fmt.Errorf("secretKeyRef.name must be set")
	}

	if ref.SecretKeyRef.Key == "" {
		return "", fmt.Errorf("secretKeyRef.key must be set")
	}

	ns := ref.SecretKeyRef.Namespace
	if ns == "" {
		ns = defaultNamespace
	}

	if ns == "" {
		return "", fmt.Errorf("secretKeyRef.namespace must be set")
	}

	secret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: ref.SecretKeyRef.Name}, secret); err != nil {
		return "", fmt.Errorf("unable to get secret %s/%s: %w", ns, ref.SecretKeyRef.Name, err)
	}

	val, ok := secret.Data[ref.SecretKeyRef.Key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret %s/%s", ref.SecretKeyRef.Key, ns, ref.SecretKeyRef.Name)
	}

	return string(val), nil
}

// VerifyStackitConnectivity checks that the resolved credentials can
// successfully authenticate against and reach the STACKIT API, by fetching
// the referenced project via the Resource Manager API. It is used by the
// ServiceAccountCredentials and ClusterServiceAccountCredentials controllers
// to populate status.connected.
func VerifyStackitConnectivity(ctx context.Context, creds *StackitCredentials) error {
	rmClient, err := resourcemanager.NewAPIClient(creds.Options...)
	if err != nil {
		return fmt.Errorf("failed to build resourcemanager client: %w", err)
	}

	if _, err := rmClient.DefaultAPI.GetProject(ctx, creds.ProjectID).Execute(); err != nil {
		return fmt.Errorf("failed to reach STACKIT API: %w", err)
	}

	return nil
}
