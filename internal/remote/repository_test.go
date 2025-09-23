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

	"github.com/oras-project/oras-mcp/internal/version"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
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
				if repo.PlainHTTP != false { // PlainHTTP should be false for non-localhost hosts, even if using HTTP
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

// TestIsPlainHttp tests the isPlainHttp function
func TestIsPlainHttp(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		want     bool
	}{
		{
			name:     "localhost",
			registry: "localhost",
			want:     true,
		},
		{
			name:     "localhost with port",
			registry: "localhost:5000",
			want:     true,
		},
		{
			name:     "non-localhost registry",
			registry: "example.com",
			want:     false,
		},
		{
			name:     "non-localhost registry with port",
			registry: "example.com:5000",
			want:     false,
		},
		{
			name:     "IP with port",
			registry: "192.168.1.1:5000",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPlainHttp(tt.registry)
			if got != tt.want {
				t.Errorf("isPlainHttp() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAuthClient tests the authClient function
func TestAuthClient(t *testing.T) {
	// Save the original version values to restore after test
	originalVersion := version.Version
	originalBuildMetadata := version.BuildMetadata
	defer func() {
		version.Version = originalVersion
		version.BuildMetadata = originalBuildMetadata
	}()

	// Set test values
	version.Version = "1.0.0"
	version.BuildMetadata = "test"

	// Test creating the auth client
	client, err := authClient()
	if err != nil {
		t.Fatalf("authClient() error = %v", err)
	}

	// Check client properties
	if client == nil {
		t.Fatal("authClient() returned nil client")
	}

	// Check if UserAgent is set correctly
	expectedUserAgent := "oras-mcp/1.0.0+test"
	userAgentSet := false

	// Since we can't directly access the client's user-agent, let's make a test request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == expectedUserAgent {
			userAgentSet = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	_, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make test request: %v", err)
	}

	if !userAgentSet {
		t.Errorf("Expected User-Agent header to be set to %s", expectedUserAgent)
	}

	// Check if cache is initialized
	if client.Cache == nil {
		t.Error("Expected Cache to be initialized")
	}

	// Check if credential store is set
	if client.Credential == nil {
		t.Error("Expected Credential function to be set")
	}
}

// TestDefaultClient checks that the DefaultClient is properly initialized
func TestDefaultClient(t *testing.T) {
	if DefaultClient == nil {
		t.Fatal("DefaultClient should not be nil")
	}

	// Check that it's an auth.Client
	if _, ok := interface{}(DefaultClient).(*auth.Client); !ok {
		t.Errorf("DefaultClient is not of type *auth.Client")
	}

	// Test basic functionality of the client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to use DefaultClient: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
