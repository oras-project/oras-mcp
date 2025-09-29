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
	"oras.land/oras-go/v2/registry"
)

// MetadataParseReference describes the ParseReference tool.
var MetadataParseReference = &mcp.Tool{
	Name:        "parse_reference",
	Description: "Parse a reference string into its components of registry, repository, tag, and digest.",
}

// InputParseReference is the input for the ParseReference tool.
type InputParseReference struct {
	Reference string `json:"reference" jsonschema:"reference string"`
}

// OutputParseReference is the output for the ParseReference tool.
type OutputParseReference struct {
	Registry   string `json:"registry" jsonschema:"registry name"`
	Repository string `json:"repository" jsonschema:"repository name"`
	Tag        string `json:"tag,omitempty" jsonschema:"tag name"`
	Digest     string `json:"digest,omitempty" jsonschema:"manifest digest"`
}

// ParseReference parses a reference string into its components.
func ParseReference(ctx context.Context, _ *mcp.CallToolRequest, input InputParseReference) (*mcp.CallToolResult, OutputParseReference, error) {
	if input.Reference == "" {
		return nil, OutputParseReference{}, fmt.Errorf("reference string is required")
	}

	// parse the reference
	ref, err := registry.ParseReference(input.Reference)
	if err != nil {
		return nil, OutputParseReference{}, fmt.Errorf("invalid reference string format: %w", err)
	}

	output := OutputParseReference{
		Registry:   ref.Registry,
		Repository: ref.Repository,
	}

	// set tag and digest if present
	if ref.Reference != "" {
		if ref.ValidateReferenceAsDigest() == nil {
			output.Digest = ref.Reference
		} else {
			output.Tag = ref.Reference
		}
	}

	return nil, output, nil
}
