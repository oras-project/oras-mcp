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
	_ "crypto/sha256"
	"strings"
	"testing"
)

func TestParseReference_Success(t *testing.T) {
	const validDigest = "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	tests := []struct {
		name      string
		reference string
		want      OutputParseReference
	}{
		{
			name:      "repository only",
			reference: "registry.example.com/hello-world",
			want: OutputParseReference{
				Registry:   "registry.example.com",
				Repository: "hello-world",
			},
		},
		{
			name:      "tag reference",
			reference: "registry.example.com/hello-world:v1",
			want: OutputParseReference{
				Registry:   "registry.example.com",
				Repository: "hello-world",
				Tag:        "v1",
			},
		},
		{
			name:      "digest reference",
			reference: "registry.example.com/hello-world@" + validDigest,
			want: OutputParseReference{
				Registry:   "registry.example.com",
				Repository: "hello-world",
				Digest:     validDigest,
			},
		},
		{
			name:      "tag and digest reference",
			reference: "registry.example.com/hello-world:v1@" + validDigest,
			want: OutputParseReference{
				Registry:   "registry.example.com",
				Repository: "hello-world",
				Digest:     validDigest,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result, output, err := ParseReference(ctx, nil, InputParseReference{Reference: tt.reference})
			if err != nil {
				t.Fatalf("ParseReference() unexpected error: %v", err)
			}
			if result != nil {
				t.Fatalf("ParseReference() result = %v, want nil", result)
			}
			if output != tt.want {
				t.Fatalf("ParseReference() output = %+v, want %+v", output, tt.want)
			}
		})
	}
}

func TestParseReference_Error(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		wantErr   string
	}{
		{
			name:      "empty reference",
			reference: "",
			wantErr:   "reference string is required",
		},
		{
			name:      "missing registry",
			reference: "hello-world:v1",
			wantErr:   "invalid reference string format",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result, output, err := ParseReference(ctx, nil, InputParseReference{Reference: tt.reference})
			if err == nil {
				t.Fatalf("ParseReference() error = nil, want non-nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ParseReference() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
			if result != nil {
				t.Fatalf("ParseReference() result = %v, want nil", result)
			}
			if output != (OutputParseReference{}) {
				t.Fatalf("ParseReference() output = %+v, want zero value", output)
			}
		})
	}
}
