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
	"reflect"
	"testing"
)

func TestListWellknownRegistries(t *testing.T) {
	// Create test context and input
	ctx := context.Background()
	input := InputListWellknownRegistries{}

	// Call the function under test
	result, output, err := ListWellknownRegistries(ctx, nil, input)

	// Check the results
	if err != nil {
		t.Fatalf("ListWellknownRegistries() error = %v", err)
	}
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}

	// Check that the output contains the expected registries
	expectedRegistries := []Registry{
		{
			Name:        "mcr.microsoft.com",
			Description: "Microsoft Container Registry",
		},
	}
	if !reflect.DeepEqual(output.Registries, expectedRegistries) {
		t.Errorf("ListWellknownRegistries() = %v, want %v", output.Registries, expectedRegistries)
	}
}
