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
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestListReferrers_InvalidInput(t *testing.T) {
	testCases := []struct {
		name  string
		input InputListReferrers
	}{
		{
			name: "missing registry",
			input: InputListReferrers{
				Repository: "repo",
				Tag:        "latest",
			},
		},
		{
			name: "missing repository",
			input: InputListReferrers{
				Registry: "localhost:5000",
				Tag:      "latest",
			},
		},
		{
			name: "missing reference",
			input: InputListReferrers{
				Registry:   "localhost:5000",
				Repository: "repo",
			},
		},
		{
			name: "invalid repository name",
			input: InputListReferrers{
				Registry:   "localhost:5000",
				Repository: "INVALID_REPO",
				Tag:        "latest",
			},
		},
		{
			name: "invalid digest",
			input: InputListReferrers{
				Registry:   "localhost:5000",
				Repository: "repo",
				Digest:     "invalid-digest",
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, output, err := ListReferrers(ctx, nil, tc.input)
			if err == nil {
				t.Fatalf("ListReferrers() error = nil, want error")
			}
			if result != nil {
				t.Fatalf("expected MCP result to be nil, got %v", result)
			}
			if output.Root != nil {
				t.Fatalf("expected empty output on error, got root %v", output.Root)
			}
		})
	}
}

func TestListReferrers_ResolveError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputListReferrers{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Tag:        "latest",
	}

	result, output, err := ListReferrers(ctx, nil, input)
	if err == nil {
		t.Fatal("ListReferrers() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if output.Root != nil {
		t.Fatalf("expected empty output on error, got root %v", output.Root)
	}
}

func TestListReferrers_ReferrersError(t *testing.T) {
	manifest := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json"}`)
	rootDigest := digest.FromBytes(manifest)

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/test-repo/manifests/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead && r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
		w.Header().Set("Docker-Content-Digest", rootDigest.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v2/test-repo/referrers/", func(w http.ResponseWriter, r *http.Request) {
		// use a 418 error for testing instead of 500 to avoid retry logic
		w.WriteHeader(http.StatusTeapot)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx := context.Background()
	input := InputListReferrers{
		Registry:     getLocalhostServerURL(ts.URL),
		Repository:   "test-repo",
		Tag:          "latest",
		ArtifactType: "application/vnd.test",
	}

	result, output, err := ListReferrers(ctx, nil, input)
	if err == nil {
		t.Fatal("ListReferrers() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if output.Root != nil {
		t.Fatalf("expected empty output on error, got root %v", output.Root)
	}
}

func TestListReferrers_Success(t *testing.T) {
	manifest := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json"}`)
	rootDigest := digest.FromBytes(manifest)

	child1Digest := digest.FromString("child-1")
	child2Digest := digest.FromString("child-2")
	grandchildDigest := digest.FromString("grand-child")

	child1Desc := ocispec.Descriptor{
		MediaType:    ocispec.MediaTypeImageManifest,
		Digest:       child1Digest,
		Size:         123,
		ArtifactType: "application/vnd.test",
	}
	child2Desc := ocispec.Descriptor{
		MediaType:    ocispec.MediaTypeImageIndex,
		Digest:       child2Digest,
		Size:         456,
		ArtifactType: "application/vnd.test",
	}
	grandchildDesc := ocispec.Descriptor{
		MediaType:    ocispec.MediaTypeImageManifest,
		Digest:       grandchildDigest,
		Size:         789,
		ArtifactType: "application/vnd.test",
	}

	referrersByDigest := map[string][]ocispec.Descriptor{
		rootDigest.String():       {child1Desc, child2Desc},
		child1Digest.String():     {grandchildDesc},
		child2Digest.String():     {},
		grandchildDigest.String(): {},
	}

	var requestedDigests []string
	var artifactTypeRequests []string

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/test-repo/manifests/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead && r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
		w.Header().Set("Docker-Content-Digest", rootDigest.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(manifest); err != nil {
				t.Fatalf("failed to write manifest: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v2/test-repo/referrers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		digestStr := strings.TrimPrefix(r.URL.Path, "/v2/test-repo/referrers/")
		requestedDigests = append(requestedDigests, digestStr)
		artifactType := r.URL.Query().Get("artifactType")
		artifactTypeRequests = append(artifactTypeRequests, artifactType)

		manifests, ok := referrersByDigest[digestStr]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", ocispec.MediaTypeImageIndex)
		if artifactType != "" {
			w.Header().Set("OCI-Filters-Applied", "artifactType")
		}
		index := ocispec.Index{
			Versioned: specs.Versioned{SchemaVersion: 2},
			MediaType: ocispec.MediaTypeImageIndex,
			Manifests: manifests,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(index); err != nil {
			t.Fatalf("failed to encode index: %v", err)
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx := context.Background()
	input := InputListReferrers{
		Registry:     getLocalhostServerURL(ts.URL),
		Repository:   "test-repo",
		Tag:          "latest",
		ArtifactType: "application/vnd.test",
	}

	result, output, err := ListReferrers(ctx, nil, input)
	if err != nil {
		t.Fatalf("ListReferrers() error = %v", err)
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if output.Root == nil {
		t.Fatal("expected root descriptor, got nil")
	}

	if output.Root.Digest != rootDigest {
		t.Fatalf("unexpected root digest: got %s, want %s", output.Root.Digest, rootDigest)
	}
	if len(output.Root.Referrers) != 2 {
		t.Fatalf("unexpected root referrers count: got %d, want 2", len(output.Root.Referrers))
	}
	if output.Root.Referrers[0].Digest != child1Digest {
		t.Fatalf("unexpected first child digest: got %s, want %s", output.Root.Referrers[0].Digest, child1Digest)
	}
	if output.Root.Referrers[1].Digest != child2Digest {
		t.Fatalf("unexpected second child digest: got %s, want %s", output.Root.Referrers[1].Digest, child2Digest)
	}

	child1Node := output.Root.Referrers[0]
	if len(child1Node.Referrers) != 1 {
		t.Fatalf("unexpected child1 referrers count: got %d, want 1", len(child1Node.Referrers))
	}
	if child1Node.Referrers[0].Digest != grandchildDigest {
		t.Fatalf("unexpected grandchild digest: got %s, want %s", child1Node.Referrers[0].Digest, grandchildDigest)
	}

	child2Node := output.Root.Referrers[1]
	if len(child2Node.Referrers) != 0 {
		t.Fatalf("unexpected child2 referrers count: got %d, want 0", len(child2Node.Referrers))
	}

	if len(requestedDigests) != 4 {
		t.Fatalf("unexpected referrers requests count: got %d, want 4", len(requestedDigests))
	}
	expectedOrder := []string{
		rootDigest.String(),
		child1Digest.String(),
		grandchildDigest.String(),
		child2Digest.String(),
	}
	for i, digestStr := range requestedDigests {
		if digestStr != expectedOrder[i] {
			t.Fatalf("unexpected digest requested at position %d: got %s, want %s", i, digestStr, expectedOrder[i])
		}
	}
	for i, artifactType := range artifactTypeRequests {
		if artifactType != input.ArtifactType {
			t.Fatalf("unexpected artifact type at position %d: got %q, want %q", i, artifactType, input.ArtifactType)
		}
	}
}

func TestFetchAllReferrers_BuildsTree(t *testing.T) {
	ctx := context.Background()
	rootDigest := digest.FromString("root")
	child1Digest := digest.FromString("child1")
	child2Digest := digest.FromString("child2")
	grandchildDigest := digest.FromString("grandchild")

	child1Desc := ocispec.Descriptor{Digest: child1Digest}
	child2Desc := ocispec.Descriptor{Digest: child2Digest}
	grandchildDesc := ocispec.Descriptor{Digest: grandchildDigest}

	lister := &fakeReferrerLister{
		responses: map[string]fakeReferrersResponse{
			rootDigest.String(): {
				refs: []ocispec.Descriptor{child1Desc, child2Desc, child1Desc},
			},
			child1Digest.String(): {
				refs: []ocispec.Descriptor{grandchildDesc, {Digest: rootDigest}},
			},
			child2Digest.String(): {
				refs: []ocispec.Descriptor{child1Desc},
			},
			grandchildDigest.String(): {},
		},
	}

	root := &ListReferrersNode{Descriptor: ocispec.Descriptor{Digest: rootDigest}}

	if err := fetchAllReferrers(ctx, lister, root, ""); err != nil {
		t.Fatalf("fetchAllReferrers() error = %v", err)
	}

	if len(root.Referrers) != 2 {
		t.Fatalf("unexpected root referrers count: got %d, want 2", len(root.Referrers))
	}
	if root.Referrers[0].Digest != child1Digest {
		t.Fatalf("unexpected first child digest: got %s, want %s", root.Referrers[0].Digest, child1Digest)
	}
	if root.Referrers[1].Digest != child2Digest {
		t.Fatalf("unexpected second child digest: got %s, want %s", root.Referrers[1].Digest, child2Digest)
	}
	if len(root.Referrers[0].Referrers) != 1 {
		t.Fatalf("unexpected grandchildren count: got %d, want 1", len(root.Referrers[0].Referrers))
	}
	if root.Referrers[0].Referrers[0].Digest != grandchildDigest {
		t.Fatalf("unexpected grandchild digest: got %s, want %s", root.Referrers[0].Referrers[0].Digest, grandchildDigest)
	}
	if len(root.Referrers[1].Referrers) != 0 {
		t.Fatalf("expected no referrers for child2, got %d", len(root.Referrers[1].Referrers))
	}

	expectedCalls := []string{
		rootDigest.String(),
		child1Digest.String(),
		grandchildDigest.String(),
		child2Digest.String(),
	}
	if len(lister.calls) != len(expectedCalls) {
		t.Fatalf("unexpected number of calls: got %d, want %d", len(lister.calls), len(expectedCalls))
	}
	for i, call := range lister.calls {
		if call.digest != expectedCalls[i] {
			t.Fatalf("unexpected digest at call %d: got %s, want %s", i, call.digest, expectedCalls[i])
		}
		if call.artifactType != "" {
			t.Fatalf("unexpected artifact type at call %d: got %q, want %q", i, call.artifactType, "")
		}
	}
}

type fakeReferrersCall struct {
	digest       string
	artifactType string
}

type fakeReferrersResponse struct {
	refs []ocispec.Descriptor
	err  error
}

type fakeReferrerLister struct {
	responses map[string]fakeReferrersResponse
	calls     []fakeReferrersCall
}

func (f *fakeReferrerLister) Referrers(ctx context.Context, desc ocispec.Descriptor, artifactType string, fn func(referrers []ocispec.Descriptor) error) error {
	resp := f.responses[desc.Digest.String()]
	f.calls = append(f.calls, fakeReferrersCall{digest: desc.Digest.String(), artifactType: artifactType})
	if resp.err != nil {
		return resp.err
	}
	if len(resp.refs) == 0 {
		return nil
	}
	return fn(resp.refs)
}

func TestFetchAllReferrers_ErrorPropagation(t *testing.T) {
	ctx := context.Background()
	rootDigest := digest.FromString("root")

	lister := &fakeReferrerLister{
		responses: map[string]fakeReferrersResponse{
			rootDigest.String(): {
				err: errors.New("boom"),
			},
		},
	}
	root := &ListReferrersNode{Descriptor: ocispec.Descriptor{Digest: rootDigest}}

	err := fetchAllReferrers(ctx, lister, root, "test/artifact")
	if err == nil {
		t.Fatal("fetchAllReferrers() error = nil, want error")
	}
	if !errors.Is(err, lister.responses[rootDigest.String()].err) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(root.Referrers) != 0 {
		t.Fatalf("expected no referrers on error, got %d", len(root.Referrers))
	}
	if len(lister.calls) != 1 {
		t.Fatalf("expected 1 call before error, got %d", len(lister.calls))
	}
	if lister.calls[0].artifactType != "test/artifact" {
		t.Fatalf("unexpected artifact type: got %q, want %q", lister.calls[0].artifactType, "test/artifact")
	}
}
