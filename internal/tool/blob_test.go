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
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"oras.land/oras-go/v2/content"
)

func TestFetchBlob_Success(t *testing.T) {
	blob := []byte(`{"hello":"world"}`)
	dgst := digest.FromBytes(blob)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v2/test-repo/blobs/" + dgst.String()
		if r.URL.Path != expectedPath {
			t.Fatalf("unexpected path accessed: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			t.Fatalf("unexpected method accessed: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(blob)))
		w.Header().Set("Accept-Ranges", "bytes")

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(blob); err != nil {
			t.Fatalf("failed to write blob: %v", err)
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchBlob{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Digest:     dgst.String(),
	}

	result, output, err := FetchBlob(ctx, nil, input)
	if err != nil {
		t.Fatalf("FetchBlob() error = %v", err)
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if !bytes.Equal(output.Data, blob) {
		t.Fatalf("unexpected blob data: got %s, want %s", string(output.Data), string(blob))
	}
}

func TestFetchBlob_InvalidInput(t *testing.T) {
	testCases := []struct {
		name  string
		input InputFetchBlob
	}{
		{
			name: "missing registry",
			input: InputFetchBlob{
				Repository: "repo",
				Digest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		{
			name: "missing repository",
			input: InputFetchBlob{
				Registry: "localhost:5000",
				Digest:   "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		{
			name: "missing digest",
			input: InputFetchBlob{
				Registry:   "localhost:5000",
				Repository: "repo",
			},
		},
		{
			name: "invalid digest format",
			input: InputFetchBlob{
				Registry:   "localhost:5000",
				Repository: "repo",
				Digest:     "sha256:zzzz",
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, output, err := FetchBlob(ctx, nil, tc.input)
			if err == nil {
				t.Fatalf("FetchBlob() error = nil, want error")
			}
			if result != nil {
				t.Fatalf("expected MCP result to be nil, got %v", result)
			}
			if len(output.Data) != 0 {
				t.Fatalf("expected empty output on error, got %s", string(output.Data))
			}
		})
	}
}

func TestFetchBlob_RemoteError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchBlob{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Digest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	result, output, err := FetchBlob(ctx, nil, input)
	if err == nil {
		t.Fatalf("FetchBlob() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if len(output.Data) != 0 {
		t.Fatalf("expected empty output on error, got %s", string(output.Data))
	}
}

func TestFetchBlob_NonJSONBlob(t *testing.T) {
	blob := []byte("not-json")
	dgst := digest.FromBytes(blob)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/test-repo/blobs/"+dgst.String() {
			t.Fatalf("unexpected path accessed: %s", r.URL.Path)
		}
		if r.Method == http.MethodHead {
			w.Header().Set("Docker-Content-Digest", dgst.String())
			w.Header().Set("Content-Length", strconv.Itoa(len(blob)))
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method accessed: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(blob)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(blob); err != nil {
			t.Fatalf("failed to write blob: %v", err)
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchBlob{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Digest:     dgst.String(),
	}

	result, output, err := FetchBlob(ctx, nil, input)
	if err == nil {
		t.Fatalf("FetchBlob() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "non-JSON blob is unsupported") {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if len(output.Data) != 0 {
		t.Fatalf("expected empty output on error, got %s", string(output.Data))
	}
}

func TestFetchBlob_BlobTooLarge(t *testing.T) {
	blob := bytes.Repeat([]byte("a"), maxBlobSize+1)
	dgst := digest.FromBytes(blob)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v2/test-repo/blobs/" + dgst.String()
		if r.URL.Path != expectedPath {
			t.Fatalf("unexpected path accessed: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			t.Fatalf("unexpected method accessed: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Content-Length", strconv.Itoa(len(blob)))
		w.Header().Set("Accept-Ranges", "bytes")

		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			if _, err := w.Write(blob); err != nil {
				t.Fatalf("failed to write blob: %v", err)
			}
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchBlob{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Digest:     dgst.String(),
	}

	result, output, err := FetchBlob(ctx, nil, input)
	if err == nil {
		t.Fatalf("FetchBlob() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "blob too large") {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if len(output.Data) != 0 {
		t.Fatalf("expected empty output on error, got %s", string(output.Data))
	}
}

func TestFetchBlob_DigestMismatch(t *testing.T) {
	expectedBlob := []byte(`{"value":"AAAA"}`)
	actualBlob := []byte(`{"value":"BBBB"}`)
	if len(expectedBlob) != len(actualBlob) {
		t.Fatalf("expected and actual blobs must be the same length for this test")
	}
	dgst := digest.FromBytes(expectedBlob)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v2/test-repo/blobs/" + dgst.String()
		if r.URL.Path != expectedPath {
			t.Fatalf("unexpected path accessed: %s", r.URL.Path)
		}

		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Docker-Content-Digest", dgst.String())
			w.Header().Set("Content-Length", strconv.Itoa(len(actualBlob)))
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Docker-Content-Digest", dgst.String())
			w.Header().Set("Content-Length", strconv.Itoa(len(actualBlob)))
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(actualBlob); err != nil {
				t.Fatalf("failed to write blob: %v", err)
			}
		default:
			t.Fatalf("unexpected method accessed: %s", r.Method)
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	input := InputFetchBlob{
		Registry:   getLocalhostServerURL(ts.URL),
		Repository: "test-repo",
		Digest:     dgst.String(),
	}

	result, output, err := FetchBlob(ctx, nil, input)
	if err == nil {
		t.Fatalf("FetchBlob() error = nil, want error")
	}
	if !strings.Contains(err.Error(), content.ErrMismatchedDigest.Error()) {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected MCP result to be nil, got %v", result)
	}
	if len(output.Data) != 0 {
		t.Fatalf("expected empty output on error, got %s", string(output.Data))
	}
}
