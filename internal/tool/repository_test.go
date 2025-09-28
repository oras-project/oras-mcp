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
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestListRepositoriesSuccess(t *testing.T) {
	handler := func() http.Handler {
		var call int
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/_catalog" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			switch call {
			case 0:
				w.Header().Set("Link", "</v2/_catalog?last=repo1>; rel=\"next\"")
				if err := json.NewEncoder(w).Encode(map[string]any{"repositories": []string{"repo1"}}); err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}
			default:
				if err := json.NewEncoder(w).Encode(map[string]any{"repositories": []string{"repo2", "repo3"}}); err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}
			}
			call++
		})
	}

	ts := httptest.NewServer(handler())
	defer ts.Close()

	registry := getLocalhostServerURL(ts.URL)
	_, output, err := ListRepositories(context.Background(), nil, InputListRepositories{Registry: registry})
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}

	expected := []string{"repo1", "repo2", "repo3"}
	if !slices.Equal(output.Repositories, expected) {
		t.Fatalf("Repositories = %v, want %v", output.Repositories, expected)
	}
}

func TestListRepositoriesMissingRegistry(t *testing.T) {
	if _, _, err := ListRepositories(context.Background(), nil, InputListRepositories{}); err == nil {
		t.Fatal("expected error for missing registry, got nil")
	}
}

func TestListRepositoriesServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/_catalog" {
			http.NotFound(w, r)
			return
		}
		// use 418 to avoid triggering retry logic in the client
		w.WriteHeader(http.StatusTeapot)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{
				"code":    "UNKNOWN",
				"message": "internal",
			}},
		})
	}))
	defer ts.Close()

	registry := getLocalhostServerURL(ts.URL)
	if _, _, err := ListRepositories(context.Background(), nil, InputListRepositories{Registry: registry}); err == nil {
		t.Fatal("expected error when registry returns failure")
	}
}
