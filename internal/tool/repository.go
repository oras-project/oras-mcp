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
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	if input.Registry == "" {
		return nil, OutputListRepositories{}, fmt.Errorf("registry name is required")
	}

	// list repositories using oras CLI
	cmd := exec.CommandContext(ctx, "oras", "repo", "list", input.Registry)
	result, err := cmd.Output()
	if err != nil {
		return nil, OutputListRepositories{}, err
	}

	// parse output
	repositories := []string{}
	for line := range strings.SplitSeq(string(result), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			repositories = append(repositories, line)
		}
	}

	output := OutputListRepositories{
		Repositories: repositories,
	}

	return nil, output, nil
}
