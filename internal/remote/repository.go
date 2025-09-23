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

	"github.com/oras-project/oras-mcp/internal/version"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// DefaultClient is the default oras-mcp auth client shared by all repositories.
// This is intended for performance optimization to reduce repeated
// authentication and token exchange requests for the MCP server in the stdio
// mode.
var DefaultClient *auth.Client

func init() {
	var err error
	DefaultClient, err = authClient()
	if err != nil {
		panic(err)
	}
}

// NewRepository assembles an oras-mcp remote repository.
func NewRepository(ref registry.Reference) *remote.Repository {
	return &remote.Repository{
		Client:          DefaultClient,
		Reference:       ref,
		PlainHTTP:       isPlainHttp(ref.Registry),
		SkipReferrersGC: true,
	}
}

// isPlainHttp determines whether to use plain HTTP for the given registry.
func isPlainHttp(registry string) bool {
	host, _, _ := net.SplitHostPort(registry)
	return host == "localhost" || registry == "localhost"
}

// authClient assembles an oras-mcp auth client.
func authClient() (*auth.Client, error) {
	client := &auth.Client{
		Client: &http.Client{
			// http.RoundTripper with a retry using the DefaultPolicy
			// see: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/retry#Policy
			Transport: retry.NewTransport(http.DefaultTransport),
		},
		Cache: auth.NewCache(),
	}
	client.SetUserAgent("oras-mcp/" + version.GetVersion())

	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return nil, err
	}
	client.Credential = credentials.Credential(store)
	return client, nil
}
