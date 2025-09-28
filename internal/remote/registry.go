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

import "oras.land/oras-go/v2/registry/remote"

// NewRegistry assembles an oras-mcp remote registry client.
func NewRegistry(name string) (*remote.Registry, error) {
	reg, err := remote.NewRegistry(name)
	if err != nil {
		return nil, err
	}

	reg.Client = DefaultClient
	reg.PlainHTTP = isPlainHttp(name)

	return reg, nil
}
