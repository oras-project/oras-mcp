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
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MetadataListReferrers describes the ListReferrers tool.
var MetadataListReferrers = &mcp.Tool{
	Name:        "list_referrers",
	Description: "List referrers of a container image or an OCI artifact.",
}

// InputListReferrers is the input for the ListReferrers tool.
type InputListReferrers struct {
	Registry   string `json:"registry" jsonschema:"registry name"`
	Repository string `json:"repository" jsonschema:"repository name"`
	Tag        string `json:"tag,omitempty" jsonschema:"tag name"`
	Digest     string `json:"digest,omitempty" jsonschema:"manifest digest"`
}

// OutputListReferrers is the output for the ListReferrers tool.
type OutputListReferrers struct {
	Data json.RawMessage `json:",inline" jsonschema:"referrers data in JSON format"`
}

// ListReferrers lists referrers of a container image or an OCI artifact.
func ListReferrers(ctx context.Context, _ *mcp.CallToolRequest, input InputListReferrers) (*mcp.CallToolResult, OutputListReferrers, error) {
	if input.Registry == "" || input.Repository == "" {
		return nil, OutputListReferrers{}, fmt.Errorf("registry and repository names are required")
	}
	if input.Tag == "" && input.Digest == "" {
		return nil, OutputListReferrers{}, fmt.Errorf("either tag or digest is required")
	}

	// construct the reference string
	reference := fmt.Sprintf("%s/%s", input.Registry, input.Repository)
	if input.Tag != "" {
		reference += ":" + input.Tag
	} else if input.Digest != "" {
		reference += "@" + input.Digest
	}

	// list referrers using oras CLI
	cmd := exec.CommandContext(ctx, "oras", "discover", "--format", "json", reference)
	result, err := cmd.Output()
	if err != nil {
		return nil, OutputListReferrers{}, err
	}

	output := OutputListReferrers{
		Data: json.RawMessage(result),
	}

	return nil, output, nil
}
