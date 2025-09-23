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
	"github.com/oras-project/oras-mcp/internal/remote"
	"oras.land/oras-go/v2/registry"
)

// MetadataListTags describes the ListTags tool.
var MetadataListTags = &mcp.Tool{
	Name:        "list_tags",
	Description: "List tags in a repository of a container registry.",
}

// InputListTags is the input for the ListTags tool.
type InputListTags struct {
	Registry   string `json:"registry" jsonschema:"registry name"`
	Repository string `json:"repository" jsonschema:"repository name"`
}

// OutputListTags is the output for the ListTags tool.
type OutputListTags struct {
	Tags []string `json:"tags" jsonschema:"list of tags"`
}

// ListTags lists tags in a repository of a container registry.
func ListTags(ctx context.Context, _ *mcp.CallToolRequest, input InputListTags) (*mcp.CallToolResult, OutputListTags, error) {
	// validate input
	ref := registry.Reference{
		Registry:   input.Registry,
		Repository: input.Repository,
	}
	if err := ref.Validate(); err != nil {
		return nil, OutputListTags{}, err
	}
	repo := remote.NewRepository(ref)

	// list tags
	tags, err := registry.Tags(ctx, repo)
	if err != nil {
		return nil, OutputListTags{}, err
	}

	output := OutputListTags{
		Tags: tags,
	}
	return nil, output, nil
}
