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
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/oras-project/oras-mcp/internal/remote"
	"oras.land/oras-go/v2/registry"
)

// MetadataListReferrers describes the ListReferrers tool.
var MetadataListReferrers = &mcp.Tool{
	Name:        "list_referrers",
	Description: "List referrers of a container image or an OCI artifact.",
}

// InputListReferrers is the input for the ListReferrers tool.
type InputListReferrers struct {
	Registry     string `json:"registry" jsonschema:"registry name"`
	Repository   string `json:"repository" jsonschema:"repository name"`
	Tag          string `json:"tag,omitempty" jsonschema:"tag name"`
	Digest       string `json:"digest,omitempty" jsonschema:"manifest digest"`
	ArtifactType string `json:"artifactType,omitempty" jsonschema:"filter by artifact type"`
}

// OutputListReferrers is the output for the ListReferrers tool.
//
// MCP Go SDK rejects cyclic schemas. Referrers form a recursive tree, so we
// pre-marshal it and inline the JSON to keep the schema acyclic.
type OutputListReferrers struct {
	Data json.RawMessage `json:",inline" jsonschema:"referrers of the requested artifact"`
}

type ListReferrersNode struct {
	ocispec.Descriptor
	Referrers []*ListReferrersNode `json:"referrers"`
}

// ListReferrers lists referrers of a container image or an OCI artifact.
func ListReferrers(ctx context.Context, _ *mcp.CallToolRequest, input InputListReferrers) (*mcp.CallToolResult, OutputListReferrers, error) {
	// validate input
	if input.Registry == "" || input.Repository == "" {
		return nil, OutputListReferrers{}, fmt.Errorf("registry and repository names are required")
	}
	if input.Tag == "" && input.Digest == "" {
		return nil, OutputListReferrers{}, fmt.Errorf("either tag or digest is required")
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
		return nil, OutputListReferrers{}, err
	}
	repo := remote.NewRepository(ref)

	// resolve the reference to get the descriptor
	desc, err := repo.Manifests().Resolve(ctx, ref.Reference)
	if err != nil {
		return nil, OutputListReferrers{}, err
	}

	// fetch referrers
	root := &ListReferrersNode{
		Descriptor: desc,
	}
	if err := fetchAllReferrers(ctx, repo, root, input.ArtifactType); err != nil {
		return nil, OutputListReferrers{}, err
	}

	// json.Marshal on ListReferrersNode never fails because the structure only
	// contains JSON-serializable fields; safe to ignore the error.
	rootJSON, _ := json.Marshal(root)

	output := OutputListReferrers{
		Data: json.RawMessage(rootJSON),
	}
	return nil, output, nil
}

// fetchAllReferrers fetches all referrers of the root node recursively.
func fetchAllReferrers(ctx context.Context, repo registry.ReferrerLister, root *ListReferrersNode, artifactType string) error {
	// referrers forms a strict tree although nodes for artifacts form a DAG.
	// `visited` is a fail-safe mechanism in case the server malfunctions.
	// The logic still works if `visited` is removed.
	stack := []*ListReferrersNode{root}
	visited := map[digest.Digest]struct{}{
		root.Digest: {},
	}

	for len(stack) > 0 {
		// pop for depth-first search
		idx := len(stack) - 1
		current := stack[idx]
		stack = stack[:idx]

		// fetch referrers of the current node
		var children []*ListReferrersNode
		if err := repo.Referrers(ctx, current.Descriptor, artifactType, func(referrers []ocispec.Descriptor) error {
			for _, node := range referrers {
				if _, ok := visited[node.Digest]; ok {
					continue
				}
				child := &ListReferrersNode{
					Descriptor: node,
				}
				current.Referrers = append(current.Referrers, child)
				visited[node.Digest] = struct{}{}
				children = append(children, child)
			}
			return nil
		}); err != nil {
			return err
		}
		for _, child := range slices.Backward(children) {
			stack = append(stack, child)
		}
	}
	return nil
}
