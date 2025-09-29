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

import "testing"

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name          string
		registry      string
		wantPlainHTTP bool
	}{
		{
			name:          "localhost uses plain HTTP",
			registry:      "localhost:5000",
			wantPlainHTTP: true,
		},
		{
			name:          "remote host uses HTTPS",
			registry:      "example.com",
			wantPlainHTTP: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, err := NewRegistry(tt.registry)
			if err != nil {
				t.Fatalf("NewRegistry() error = %v", err)
			}
			if reg == nil {
				t.Fatal("NewRegistry() returned nil registry")
			}
			if got := reg.PlainHTTP; got != tt.wantPlainHTTP {
				t.Errorf("PlainHTTP = %v, want %v", got, tt.wantPlainHTTP)
			}
			if reg.Client != DefaultClient {
				t.Errorf("Client = %v, want DefaultClient", reg.Client)
			}
			if reg.Reference.Registry != tt.registry {
				t.Errorf("Registry = %q, want %q", reg.Reference.Registry, tt.registry)
			}
		})
	}
}

func TestNewRegistryInvalid(t *testing.T) {
	if _, err := NewRegistry("https://example.com"); err == nil {
		t.Fatal("expected error for registry with scheme, got nil")
	}
}
