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

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/oras-project/oras-mcp/internal/remote"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
)

// maxBlobSize defines the maximum blob size that can be fetched.
const maxBlobSize = 4 * 1024 * 1024 // 4 MiB

// MetadataFetchBlob describes the FetchBlob tool.
var MetadataFetchBlob = &mcp.Tool{
	Name:        "fetch_blob",
	Description: "Fetch blob referenced by a digest in a manifest.",
	OutputSchema: &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: &jsonschema.Schema{},
		Description:          "Blob data in JSON format.",
	},
}

// InputFetchBlob is the input for the FetchBlob tool.
type InputFetchBlob struct {
	Registry   string `json:"registry" jsonschema:"registry name"`
	Repository string `json:"repository" jsonschema:"repository name"`
	Digest     string `json:"digest" jsonschema:"blob digest"`
}

// OutputFetchBlob is the output for the FetchBlob tool.
type OutputFetchBlob struct {
	blob json.RawMessage
}

// MarshalJSON ensures the tool response is the raw blob document without extra
// wrapping fields so agents receive the exact JSON payload fetched.
func (o OutputFetchBlob) MarshalJSON() ([]byte, error) {
	if len(o.blob) == 0 {
		return []byte("null"), nil
	}
	return o.blob, nil
}

// Raw returns the underlying blob bytes.
func (o OutputFetchBlob) Raw() []byte {
	return o.blob
}

// FetchBlob fetches blob referenced by a digest in a manifest.
func FetchBlob(ctx context.Context, _ *mcp.CallToolRequest, input InputFetchBlob) (*mcp.CallToolResult, OutputFetchBlob, error) {
	// validate input
	if input.Registry == "" {
		return nil, OutputFetchBlob{}, fmt.Errorf("registry name is required")
	}
	if input.Repository == "" {
		return nil, OutputFetchBlob{}, fmt.Errorf("repository name is required")
	}
	if input.Digest == "" {
		return nil, OutputFetchBlob{}, fmt.Errorf("blob digest is required")
	}
	ref := registry.Reference{
		Registry:   input.Registry,
		Repository: input.Repository,
		Reference:  input.Digest,
	}
	if err := ref.Validate(); err != nil {
		return nil, OutputFetchBlob{}, err
	}
	repo := remote.NewRepository(ref)

	// fetch the blob
	desc, rc, err := repo.Blobs().FetchReference(ctx, ref.Reference)
	if err != nil {
		return nil, OutputFetchBlob{}, err
	}
	defer rc.Close()
	if desc.Size > maxBlobSize {
		return nil, OutputFetchBlob{}, fmt.Errorf("blob too large: %d", desc.Size)
	}
	blobBytes, err := content.ReadAll(rc, desc)
	if err != nil {
		return nil, OutputFetchBlob{}, err
	}

	// only JSON blob is returned to the agent
	if !json.Valid(blobBytes) {
		return nil, OutputFetchBlob{}, fmt.Errorf("non-JSON blob is unsupported")
	}

	// Create the output
	output := OutputFetchBlob{
		blob: json.RawMessage(blobBytes),
	}
	return nil, output, nil
}
