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
	"strings"
	"testing"

	"github.com/oras-project/oras-mcp/cmd/oras-mcp/internal/output"
	"github.com/oras-project/oras-mcp/internal/version"
)

func TestRunVersionIncludesMetadata(t *testing.T) {
	originalVersion := version.Version
	originalBuild := version.BuildMetadata
	originalCommit := version.GitCommit
	originalTreeState := version.GitTreeState

	version.Version = "1.2.3"
	version.BuildMetadata = ""
	version.GitCommit = "abc1234"
	version.GitTreeState = "dirty"

	t.Cleanup(func() {
		version.Version = originalVersion
		version.BuildMetadata = originalBuild
		version.GitCommit = originalCommit
		version.GitTreeState = originalTreeState
	})

	var out bytes.Buffer
	var errOut bytes.Buffer
	printer := output.NewPrinter(&out, &errOut)

	if err := runVersion(printer); err != nil {
		t.Fatalf("runVersion returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Version:") || !strings.Contains(output, "1.2.3") {
		t.Fatalf("expected version information in output, got %q", output)
	}
	if !strings.Contains(output, "Git commit:") || !strings.Contains(output, "abc1234") {
		t.Fatalf("expected git commit in output, got %q", output)
	}
	if !strings.Contains(output, "Git tree state:") || !strings.Contains(output, "dirty") {
		t.Fatalf("expected git tree state in output, got %q", output)
	}

	if errOut.Len() != 0 {
		t.Fatalf("expected no error output, got %q", errOut.String())
	}
}

func TestVersionCommandExecution(t *testing.T) {
	cmd := versionCmd()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version command returned error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Fatalf("expected output from version command execution")
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}
