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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/opencontainers/go-digest"
)

func TestFetchManifest_OutputSchema(t *testing.T) {
	rawSchema := MetadataFetchManifest.OutputSchema
	if rawSchema == nil {
		t.Fatal("OutputSchema is nil")
	}

	schema, ok := rawSchema.(*jsonschema.Schema)
	if !ok {
		t.Fatalf("OutputSchema has unexpected type %T", rawSchema)
	}

	if schema.Type != "object" {
		t.Fatalf("unexpected schema type: got %q, want %q", schema.Type, "object")
	}
	if schema.Description != "Manifest data in JSON format." {
		t.Fatalf("unexpected schema description: got %q", schema.Description)
	}
	if schema.AdditionalProperties == nil {
		t.Fatal("AdditionalProperties is nil")
	}
	if schema.AdditionalProperties.Type != "" || schema.AdditionalProperties.Description != "" {
		t.Fatalf("AdditionalProperties should be unconstrained, got %+v", schema.AdditionalProperties)
	}
}

func TestOutputFetchManifest_MarshalJSON(t *testing.T) {
	manifest := []byte(`{"schemaVersion":2}`)
	output := OutputFetchManifest{manifest: json.RawMessage(manifest)}

	got, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !bytes.Equal(got, manifest) {
		t.Fatalf("unexpected marshal output: got %s, want %s", string(got), string(manifest))
	}

	var zero OutputFetchManifest
	got, err = json.Marshal(zero)
	if err != nil {
		t.Fatalf("json.Marshal() zero error = %v", err)
	}
	if string(got) != "null" {
		t.Fatalf("unexpected zero marshal output: got %s, want null", string(got))
	}
}

func TestFetchManifest_SuccessWithTag(t *testing.T) {
	manifest := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json"}`)
	dgst := digest.FromBytes(manifest)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/test-repo/manifests/latest" {
			t.Fatalf("unexpected path accessed: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			t.Fatalf("unexpected method accessed: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(manifest); err != nil {
			t.Fatalf("failed to write manifest: %v", err)
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchManifest{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Tag:        "latest",
	}

	result, output, err := FetchManifest(ctx, nil, input)
	if err != nil {
		t.Fatalf("FetchManifest() error = %v", err)
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if !bytes.Equal(output.Raw(), manifest) {
		t.Fatalf("unexpected manifest data: got %s, want %s", string(output.Raw()), string(manifest))
	}
}

func TestFetchManifest_SuccessWithDigest(t *testing.T) {
	manifest := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json"}`)
	dgst := digest.FromBytes(manifest)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/test-repo/manifests/"+dgst.String() {
			t.Fatalf("unexpected path accessed: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method accessed: %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(manifest); err != nil {
			t.Fatalf("failed to write manifest: %v", err)
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchManifest{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Digest:     dgst.String(),
	}

	result, output, err := FetchManifest(ctx, nil, input)
	if err != nil {
		t.Fatalf("FetchManifest() error = %v", err)
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if !bytes.Equal(output.Raw(), manifest) {
		t.Fatalf("unexpected manifest data: got %s, want %s", string(output.Raw()), string(manifest))
	}
}

func TestFetchManifest_InvalidInput(t *testing.T) {
	testCases := []struct {
		name  string
		input InputFetchManifest
	}{
		{
			name: "missing registry",
			input: InputFetchManifest{
				Repository: "repo",
				Tag:        "latest",
			},
		},
		{
			name: "missing repository",
			input: InputFetchManifest{
				Registry: "localhost:5000",
				Tag:      "latest",
			},
		},
		{
			name: "missing reference",
			input: InputFetchManifest{
				Registry:   "localhost:5000",
				Repository: "repo",
			},
		},
		{
			name: "invalid repository name",
			input: InputFetchManifest{
				Registry:   "localhost:5000",
				Repository: "INVALID_REPO",
				Tag:        "latest",
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, output, err := FetchManifest(ctx, nil, tc.input)
			if err == nil {
				t.Fatalf("FetchManifest() error = nil, want error")
			}
			if result != nil {
				t.Fatalf("expected MCP result to be nil, got %v", result)
			}
			if len(output.Raw()) != 0 {
				t.Fatalf("expected empty output on error, got %s", string(output.Raw()))
			}
		})
	}
}

func TestFetchManifest_RemoteError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchManifest{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Tag:        "latest",
	}

	result, output, err := FetchManifest(ctx, nil, input)
	if err == nil {
		t.Fatal("FetchManifest() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if len(output.Raw()) != 0 {
		t.Fatalf("expected empty output on error, got %s", string(output.Raw()))
	}
}

func TestFetchManifest_ContentMismatch(t *testing.T) {
	manifest := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json"}`)
	dgst := digest.Digest("sha256:" + strings.Repeat("a", 64))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(manifest); err != nil {
			t.Fatalf("failed to write manifest: %v", err)
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchManifest{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Tag:        "latest",
	}

	result, output, err := FetchManifest(ctx, nil, input)
	if err == nil {
		t.Fatal("FetchManifest() error = nil, want error due to digest mismatch")
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if len(output.Raw()) != 0 {
		t.Fatalf("expected empty output on error, got %s", string(output.Raw()))
	}
}
