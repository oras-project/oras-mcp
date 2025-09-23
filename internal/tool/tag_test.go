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
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// getLocalhostServerURL extracts the port from a test server URL and returns a localhost URL
func getLocalhostServerURL(serverURL string) string {
	_, port, _ := net.SplitHostPort(serverURL[7:]) // Split the host:port
	return "localhost:" + port
}

func TestListTags_ValidInput(t *testing.T) {
	// Setup a test server that simulates a registry
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/test-repo/tags/list" {
			t.Errorf("unexpected access: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Return a mock tags list
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test-repo","tags":["v1.0","latest","v1.1"]}`))
	}))
	defer ts.Close()

	// Get localhost server URL
	serverURL := getLocalhostServerURL(ts.URL)

	// Create test context and input
	ctx := context.Background()
	input := InputListTags{
		Registry:   serverURL,
		Repository: "test-repo",
	}

	// Call the function under test
	result, output, err := ListTags(ctx, nil, input)

	// Check the results
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}

	// Check that the output contains the expected tags
	expectedTags := []string{"v1.0", "latest", "v1.1"}
	if !reflect.DeepEqual(output.Tags, expectedTags) {
		t.Errorf("ListTags() = %v, want %v", output.Tags, expectedTags)
	}
}

func TestListTags_InvalidInput(t *testing.T) {
	// Test cases for invalid inputs
	testCases := []struct {
		name     string
		input    InputListTags
		wantErr  bool
		errorMsg string
	}{
		{
			name: "empty registry",
			input: InputListTags{
				Registry:   "",
				Repository: "test-repo",
			},
			wantErr:  true,
			errorMsg: "empty registry",
		},
		{
			name: "empty repository",
			input: InputListTags{
				Registry:   "localhost:5000",
				Repository: "",
			},
			wantErr:  true,
			errorMsg: "empty repository",
		},
		{
			name: "invalid registry",
			input: InputListTags{
				Registry:   "://invalid:registry:format",
				Repository: "test-repo",
			},
			wantErr:  true,
			errorMsg: "invalid registry",
		},
		{
			name: "invalid repository",
			input: InputListTags{
				Registry:   "localhost:5000",
				Repository: "INVALID_REPO_NAME",
			},
			wantErr:  true,
			errorMsg: "invalid repository",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, output, err := ListTags(ctx, nil, tt.input)

			// Check if we got an error when we expected one
			if (err != nil) != tt.wantErr {
				t.Errorf("ListTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For error cases, we expect result and output to be empty
			if tt.wantErr {
				if result != nil {
					t.Errorf("Expected result to be nil for error case, got %v", result)
				}
				if len(output.Tags) != 0 {
					t.Errorf("Expected empty output for error case, got %v", output)
				}
			}
		})
	}
}

func TestListTags_RegistryError(t *testing.T) {
	// Setup a test server that simulates a registry error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/test-repo/tags/list" {
			t.Errorf("unexpected access: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Return an error response
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code":"SERVER_ERROR","message":"Internal server error"}]}`))
	}))
	defer ts.Close()

	// Get localhost server URL
	serverURL := getLocalhostServerURL(ts.URL)

	// Create test context and input
	ctx := context.Background()
	input := InputListTags{
		Registry:   serverURL,
		Repository: "test-repo",
	}

	// Call the function under test
	result, output, err := ListTags(ctx, nil, input)

	// Check the results
	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if len(output.Tags) != 0 {
		t.Errorf("Expected empty output for error case, got %v", output)
	}
}

func TestListTags_EmptyTagList(t *testing.T) {
	// Setup a test server that returns an empty tag list
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/test-repo/tags/list" {
			t.Errorf("unexpected access: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Return a mock empty tags list
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test-repo","tags":[]}`))
	}))
	defer ts.Close()

	// Get localhost server URL
	serverURL := getLocalhostServerURL(ts.URL)

	// Create test context and input
	ctx := context.Background()
	input := InputListTags{
		Registry:   serverURL,
		Repository: "test-repo",
	}

	// Call the function under test
	result, output, err := ListTags(ctx, nil, input)

	// Check the results
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}

	// Check that the output contains an empty tags list
	if len(output.Tags) != 0 {
		t.Errorf("ListTags() = %v, want empty tag list", output.Tags)
	}
}

func TestListTags_RepositoryNotFound(t *testing.T) {
	// Setup a test server that simulates a repository not found error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test-repo/tags/list" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors":[{"code":"NAME_UNKNOWN","message":"Repository not found"}]}`))
			return
		}
	}))
	defer ts.Close()

	// Get localhost server URL
	serverURL := getLocalhostServerURL(ts.URL)

	// Create test context and input
	ctx := context.Background()
	input := InputListTags{
		Registry:   serverURL,
		Repository: "test-repo",
	}

	// Call the function under test
	result, output, err := ListTags(ctx, nil, input)

	// Check the results - we expect an error
	if err == nil {
		t.Fatalf("Expected error for repository not found, but got nil")
	}
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if len(output.Tags) != 0 {
		t.Errorf("Expected empty output for error case, got %v", output)
	}
}
