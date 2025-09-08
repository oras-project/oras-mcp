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

// MetadataFetchBlob describes the FetchBlob tool.
var MetadataFetchBlob = &mcp.Tool{
	Name:        "fetch_blob",
	Description: "Fetch blob referenced by a digest in a manifest.",
}

// InputFetchBlob is the input for the FetchBlob tool.
type InputFetchBlob struct {
	Registry   string `json:"registry" jsonschema:"registry name"`
	Repository string `json:"repository" jsonschema:"repository name"`
	Digest     string `json:"digest" jsonschema:"blob digest"`
}

// OutputFetchBlob is the output for the FetchBlob tool.
type OutputFetchBlob struct {
	Data json.RawMessage `json:",inline" jsonschema:"blob data in JSON format"`
}

// FetchBlob fetches blob referenced by a digest in a manifest.
func FetchBlob(ctx context.Context, _ *mcp.CallToolRequest, input InputFetchBlob) (*mcp.CallToolResult, OutputFetchBlob, error) {
	if input.Registry == "" || input.Repository == "" || input.Digest == "" {
		return nil, OutputFetchBlob{}, fmt.Errorf("registry, repository, and digest are required")
	}

	// construct the reference string
	reference := fmt.Sprintf("%s/%s@%s", input.Registry, input.Repository, input.Digest)

	// fetch blob using oras CLI
	cmd := exec.CommandContext(ctx, "oras", "blob", "fetch", "-o-", reference)
	result, err := cmd.Output()
	if err != nil {
		return nil, OutputFetchBlob{}, err
	}

	output := OutputFetchBlob{
		Data: json.RawMessage(result),
	}

	return nil, output, nil
}
