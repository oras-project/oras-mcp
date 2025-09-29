/*
Copyright The ORAS Authors.
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

package tool

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/oras-project/oras-mcp/internal/remote"
	"oras.land/oras-go/v2/registry"
)

// MetadataListRepositories describes the ListRepositories tool.
var MetadataListRepositories = &mcp.Tool{
	Name:        "list_repositories",
	Description: "List repositories of a container registry.",
}

// InputListRepositories is the input for the ListRepositories tool.
type InputListRepositories struct {
	Registry string `json:"registry" jsonschema:"registry name"`
}

// OutputListRepositories is the output for the ListRepositories tool.
type OutputListRepositories struct {
	Repositories []string `json:"repositories" jsonschema:"list of repositories"`
}

// ListRepositories lists repositories of a container registry.
func ListRepositories(ctx context.Context, _ *mcp.CallToolRequest, input InputListRepositories) (*mcp.CallToolResult, OutputListRepositories, error) {
	// validate input
	if input.Registry == "" {
		return nil, OutputListRepositories{}, fmt.Errorf("registry name is required")
	}
	reg, err := remote.NewRegistry(input.Registry)
	if err != nil {
		return nil, OutputListRepositories{}, err
	}

	// list repositories
	repositories, err := registry.Repositories(ctx, reg)
	if err != nil {
		return nil, OutputListRepositories{}, err
	}

	output := OutputListRepositories{
		Repositories: repositories,
	}
	return nil, output, nil
}
