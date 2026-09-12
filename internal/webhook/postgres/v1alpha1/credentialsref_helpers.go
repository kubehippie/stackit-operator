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
	"fmt"
	"strings"

	"github.com/kubehippie/stackit-operator/api/common"
)

// validateCredentialsRef validates the shared common.StackitCredentialsRef
// field embedded in the postgres CRDs.
func validateCredentialsRef(ref *common.StackitCredentialsRef) error {
	if ref == nil {
		return fmt.Errorf("spec.credentialsRef must be set")
	}

	if strings.TrimSpace(ref.Name) == "" {
		return fmt.Errorf("spec.credentialsRef.name must be set")
	}

	switch ref.Kind {
	case "", "ServiceAccountCredentials":
		// namespace defaults to the referencing resource's namespace, both
		// empty and set values are valid here.
	case "ClusterServiceAccountCredentials":
		if strings.TrimSpace(ref.Namespace) != "" {
			return fmt.Errorf("spec.credentialsRef.namespace must be empty when kind is ClusterServiceAccountCredentials")
		}
	default:
		return fmt.Errorf("spec.credentialsRef.kind must be one of ServiceAccountCredentials, ClusterServiceAccountCredentials")
	}

	return nil
}
