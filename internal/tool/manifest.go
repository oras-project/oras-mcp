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

// MetadataFetchManifest describes the FetchManifest tool.
var MetadataFetchManifest = &mcp.Tool{
	Name:        "fetch_manifest",
	Description: "Fetch manifest of a container image or an OCI artifact.",
}

// InputFetchManifest is the input for the FetchManifest tool.
type InputFetchManifest struct {
	Registry   string `json:"registry" jsonschema:"registry name"`
	Repository string `json:"repository" jsonschema:"repository name"`
	Tag        string `json:"tag,omitempty" jsonschema:"tag name"`
	Digest     string `json:"digest,omitempty" jsonschema:"manifest digest"`
}

// OutputFetchManifest is the output for the FetchManifest tool.
type OutputFetchManifest struct {
	Data json.RawMessage `json:",inline" jsonschema:"manifest data in JSON format"`
}

// FetchManifest fetches manifest of a container image or an OCI artifact.
func FetchManifest(ctx context.Context, _ *mcp.CallToolRequest, input InputFetchManifest) (*mcp.CallToolResult, OutputFetchManifest, error) {
	if input.Registry == "" || input.Repository == "" {
		return nil, OutputFetchManifest{}, fmt.Errorf("registry and repository names are required")
	}
	if input.Tag == "" && input.Digest == "" {
		return nil, OutputFetchManifest{}, fmt.Errorf("either tag or digest is required")
	}

	// construct the reference string
	reference := fmt.Sprintf("%s/%s", input.Registry, input.Repository)
	if input.Tag != "" {
		reference += ":" + input.Tag
	} else if input.Digest != "" {
		reference += "@" + input.Digest
	}

	// fetch manifest using oras CLI
	cmd := exec.CommandContext(ctx, "oras", "manifest", "fetch", "--format", "json", reference)
	result, err := cmd.Output()
	if err != nil {
		return nil, OutputFetchManifest{}, err
	}

	output := OutputFetchManifest{
		Data: json.RawMessage(result),
	}

	return nil, output, nil
}
