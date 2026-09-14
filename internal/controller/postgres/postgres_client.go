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
	"fmt"

	"github.com/kubehippie/stackit-operator/api/common"
	"github.com/kubehippie/stackit-operator/internal/controller"
	postgresflex "github.com/stackitcloud/stackit-sdk-go/services/postgresflex/v3api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// listInstancesPageSize is the page size used when listing instances to
// resolve an instance name to its STACKIT-assigned ID. It is set high
// enough to cover the number of instances found in a typical project in a
// single request.
const listInstancesPageSize = 1000

// PostgresFlexSession holds an authenticated Postgres Flex API client
// together with the STACKIT project ID and region resolved from a
// StackitCredentialsRef.
type PostgresFlexSession struct {
	Client    postgresflex.DefaultAPI
	ProjectID string
	Region    string
}

// NewPostgresFlexSession resolves credentialsRef to an authenticated
// Postgres Flex API client. defaultNamespace is used when the ref carries no
// explicit namespace.
func NewPostgresFlexSession(ctx context.Context, c client.Client, credentialsRef *common.StackitCredentialsRef, defaultNamespace string) (*PostgresFlexSession, error) {
	creds, err := controller.ResolveStackitCredentialsRef(ctx, c, credentialsRef, defaultNamespace)
	if err != nil {
		return nil, err
	}

	if creds.Region == "" {
		return nil, fmt.Errorf("region must be set on the referenced credentials to use the Postgres Flex API")
	}

	// The Postgres Flex v3 API expects the region as a per-call function
	// parameter (see ResolveInstanceID, InstanceConnectionInfo, etc. below),
	// not as client configuration. Passing WithRegion here makes the SDK
	// reject the client with "this API does not support setting a region in
	// the client configuration".
	pgClient, err := postgresflex.NewAPIClient(creds.Options...)
	if err != nil {
		return nil, fmt.Errorf("failed to build Postgres Flex client: %w", err)
	}

	return &PostgresFlexSession{
		Client:    pgClient.DefaultAPI,
		ProjectID: creds.ProjectID,
		Region:    creds.Region,
	}, nil
}

// ResolveInstanceID looks up the STACKIT-assigned ID of the Postgres Flex
// instance with the given name in the session's project/region.
func (s *PostgresFlexSession) ResolveInstanceID(ctx context.Context, instanceName string) (string, error) {
	resp, err := s.Client.ListInstances(ctx, s.ProjectID, s.Region).Size(listInstancesPageSize).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Postgres Flex instances: %w", err)
	}

	for _, instance := range resp.Instances {
		if instance.Name == instanceName {
			return instance.Id, nil
		}
	}

	return "", fmt.Errorf("no Postgres Flex instance named %q found in project %s/%s", instanceName, s.ProjectID, s.Region)
}

// InstanceConnectionInfo returns the host and port used to connect to the
// given Postgres Flex instance.
func (s *PostgresFlexSession) InstanceConnectionInfo(ctx context.Context, instanceID string) (host string, port int32, err error) {
	resp, err := s.Client.GetInstance(ctx, s.ProjectID, s.Region, instanceID).Execute()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get Postgres Flex instance: %w", err)
	}

	return resp.ConnectionInfo.Write.Host, resp.ConnectionInfo.Write.Port, nil
}

// FindDatabaseByName looks up an existing database by name on the given
// instance. It returns nil, nil when no matching database exists.
func (s *PostgresFlexSession) FindDatabaseByName(ctx context.Context, instanceID, name string) (*postgresflex.ListDatabase, error) {
	resp, err := s.Client.ListDatabases(ctx, s.ProjectID, s.Region, instanceID).Size(listInstancesPageSize).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list Postgres Flex databases: %w", err)
	}

	for _, db := range resp.Databases {
		if db.Name == name {
			return &db, nil
		}
	}

	return nil, nil
}

// FindUserByName looks up an existing user by name on the given instance. It
// returns nil, nil when no matching user exists.
func (s *PostgresFlexSession) FindUserByName(ctx context.Context, instanceID, name string) (*postgresflex.GetUserResponse, error) {
	resp, err := s.Client.ListUsers(ctx, s.ProjectID, s.Region, instanceID).Size(listInstancesPageSize).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list Postgres Flex users: %w", err)
	}

	for _, user := range resp.Users {
		if user.Name != name {
			continue
		}

		full, err := s.Client.GetUser(ctx, s.ProjectID, s.Region, instanceID, user.Id).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to get Postgres Flex user: %w", err)
		}
		return full, nil
	}

	return nil, nil
}
