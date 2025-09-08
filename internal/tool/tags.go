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

var MetadataListTags = &mcp.Tool{
	Name:        "list_tags",
	Description: "List tags in a repository of a container registry.",
}

type InputListTags struct {
	Registry   string `json:"registry" jsonschema:"registry name"`
	Repository string `json:"repository" jsonschema:"repository name"`
}

type OutputListTags struct {
	Tags []string `json:"tags" jsonschema:"list of tags"`
}

func ListTags(ctx context.Context, req *mcp.CallToolRequest, input InputListTags) (*mcp.CallToolResult, OutputListTags, error) {
	return nil, OutputListTags{Tags: []string{"tag1", "tag2"}}, nil
}
