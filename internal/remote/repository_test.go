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

package remote

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// TestNewRepository tests the NewRepository function for creating a repository
func TestNewRepository(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Parse the host and port from the test server URL
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse test server URL: %v", err)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("Failed to parse test server host: %v", err)
	}

	tests := []struct {
		name        string
		reference   string
		checkFields func(*testing.T, *remote.Repository)
	}{
		{
			name:      "valid reference",
			reference: host + ":" + port + "/test-repo",
			checkFields: func(t *testing.T, repo *remote.Repository) {
				if repo.PlainHTTP { // PlainHTTP should be false for non-localhost hosts, even if using HTTP
					t.Errorf("Expected PlainHTTP to be false for non-localhost host")
				}
				if repo.Client != DefaultClient {
					t.Errorf("Expected Client to be DefaultClient")
				}
				if repo.SkipReferrersGC != true {
					t.Errorf("Expected SkipReferrersGC to be true")
				}
				if repo.Reference.Registry != host+":"+port {
					t.Errorf("Expected registry to be %s, got %s", host+":"+port, repo.Reference.Registry)
				}
				if repo.Reference.Repository != "test-repo" {
					t.Errorf("Expected repository to be test-repo, got %s", repo.Reference.Repository)
				}
			},
		},
		{
			name:      "localhost reference",
			reference: "localhost:5000/test-repo",
			checkFields: func(t *testing.T, repo *remote.Repository) {
				if repo.PlainHTTP != true {
					t.Errorf("Expected PlainHTTP to be true for localhost")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse reference first
			ref, err := registry.ParseReference(tt.reference)
			if err != nil {
				t.Fatalf("Failed to parse reference: %v", err)
			}

			// Create repository with parsed reference
			repo := NewRepository(ref)
			if repo == nil {
				t.Fatal("NewRepository() returned nil repository")
			}

			tt.checkFields(t, repo)
		})
	}
}
