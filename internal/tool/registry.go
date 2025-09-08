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

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MetadataListWellknownRegistries describes the ListWellknownRegistries tool.
var MetadataListWellknownRegistries = &mcp.Tool{
	Name:        "list_wellknown_registries",
	Description: "List well-known public registries with catalog support.",
}

// InputListWellknownRegistries is the input for the ListWellknownRegistries tool.
type InputListWellknownRegistries struct {
	// No input parameters
}

// Registry represents a container registry.
type Registry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OutputListWellknownRegistries is the output for the ListWellknownRegistries tool.
type OutputListWellknownRegistries struct {
	Registries []Registry `json:"registries" jsonschema:"list of well-known registries"`
}

// ListWellknownRegistries lists well-known public registries with catalog support.
func ListWellknownRegistries(ctx context.Context, _ *mcp.CallToolRequest, _ InputListWellknownRegistries) (*mcp.CallToolResult, OutputListWellknownRegistries, error) {
	registries := []Registry{
		{
			Name:        "mcr.microsoft.com",
			Description: "Microsoft Container Registry",
		},
	}

	output := OutputListWellknownRegistries{
		Registries: registries,
	}

	return nil, output, nil
}
