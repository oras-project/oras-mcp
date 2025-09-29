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
	"bytes"
	"context"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	t.Parallel()

	cmd := New()

	if cmd.Use != "oras-mcp [command]" {
		t.Fatalf("unexpected use: %q", cmd.Use)
	}

	if !cmd.SilenceUsage {
		t.Fatalf("expected SilenceUsage to be true")
	}

	subCommands := cmd.Commands()
	if len(subCommands) != 2 {
		t.Fatalf("expected two subcommands, got %d", len(subCommands))
	}

	serveCmd, _, err := cmd.Find([]string{"serve"})
	if err != nil {
		t.Fatalf("serve command not registered: %v", err)
	}
	if serveCmd.Use != "serve" {
		t.Fatalf("unexpected serve command use: %q", serveCmd.Use)
	}

	versionCmd, _, err := cmd.Find([]string{"version"})
	if err != nil {
		t.Fatalf("version command not registered: %v", err)
	}
	if versionCmd.Use != "version" {
		t.Fatalf("unexpected version command use: %q", versionCmd.Use)
	}

	cmd.SetArgs([]string{"version"})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected version subcommand to execute successfully, got error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Fatalf("expected output from version subcommand")
	}

	if stderr.Len() != 0 {
		t.Fatalf("did not expect stderr output, got %q", stderr.String())
	}
}
