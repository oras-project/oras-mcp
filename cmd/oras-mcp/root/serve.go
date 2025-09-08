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

package root

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/oras-project/oras-mcp/internal/tool"
	"github.com/oras-project/oras-mcp/internal/version"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the ORAS MCP server",
		Long: `Start the ORAS MCP server

Example - start the server in the stdio mode:
  oras serve
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd)
		},
	}

	return cmd
}

func runServe(cmd *cobra.Command) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "oras-mcp",
		Title:   "ORAS",
		Version: version.GetVersion(),
	}, nil)

	mcp.AddTool(server, tool.MetadataListTags, tool.ListTags)

	return server.Run(cmd.Context(), &mcp.StdioTransport{})
}
