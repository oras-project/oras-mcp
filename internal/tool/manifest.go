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

// MetadataFetchManifest describes the FetchManifest tool.
var MetadataFetchManifest = &mcp.Tool{
	Name:        "fetch_manifest",
	Description: "Fetch manifest of a container image or an OCI artifact.",
	OutputSchema: &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: &jsonschema.Schema{},
		Description:          "Manifest data in JSON format.",
	},
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
	manifest json.RawMessage
}

// MarshalJSON implements json.Marshaler and emits the raw manifest payload
// without any additional wrapping so the response matches the manifest
// document exactly.
func (o OutputFetchManifest) MarshalJSON() ([]byte, error) {
	if len(o.manifest) == 0 {
		return []byte("null"), nil
	}
	return o.manifest, nil
}

// Raw returns the manifest bytes.
func (o OutputFetchManifest) Raw() []byte {
	return o.manifest
}

// FetchManifest fetches manifest of a container image or an OCI artifact.
func FetchManifest(ctx context.Context, _ *mcp.CallToolRequest, input InputFetchManifest) (*mcp.CallToolResult, OutputFetchManifest, error) {
	// validate input
	if input.Registry == "" || input.Repository == "" {
		return nil, OutputFetchManifest{}, fmt.Errorf("registry and repository names are required")
	}
	if input.Tag == "" && input.Digest == "" {
		return nil, OutputFetchManifest{}, fmt.Errorf("either tag or digest is required")
	}
	ref := registry.Reference{
		Registry:   input.Registry,
		Repository: input.Repository,
		Reference:  input.Tag,
	}
	if input.Digest != "" {
		ref.Reference = input.Digest
	}
	if err := ref.Validate(); err != nil {
		return nil, OutputFetchManifest{}, err
	}
	repo := remote.NewRepository(ref)

	// fetch the manifest
	desc, rc, err := repo.FetchReference(ctx, ref.Reference)
	if err != nil {
		return nil, OutputFetchManifest{}, err
	}
	defer rc.Close()

	manifestBytes, err := content.ReadAll(rc, desc)
	if err != nil {
		return nil, OutputFetchManifest{}, err
	}

	// output direct as manifests are already in JSON
	output := OutputFetchManifest{
		manifest: json.RawMessage(manifestBytes),
	}
	return nil, output, nil
}
