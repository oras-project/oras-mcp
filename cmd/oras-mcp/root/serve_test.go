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
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/oras-project/oras-mcp/internal/tool"
	"github.com/spf13/cobra"
)

func TestServeCommandDefinition(t *testing.T) {
	t.Parallel()

	cmd := serveCmd()

	if cmd.Use != "serve" {
		t.Fatalf("unexpected use: %q", cmd.Use)
	}

	if cmd.Args == nil {
		t.Fatalf("expected Args validation function to be defined")
	}

	if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
		t.Fatalf("expected error when arguments are provided")
	}

	if cmd.RunE == nil {
		t.Fatalf("expected RunE to be defined")
	}
}

func TestRunServeReturnsErrorOnCanceledContext(t *testing.T) {
	originalSchema := tool.MetadataListReferrers.OutputSchema
	tool.MetadataListReferrers.OutputSchema = &jsonschema.Schema{Type: "object"}
	t.Cleanup(func() {
		tool.MetadataListReferrers.OutputSchema = originalSchema
	})

	originalStdin := os.Stdin
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	t.Cleanup(func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	if err := stdinWriter.Close(); err != nil {
		t.Fatalf("failed to close stdin writer: %v", err)
	}
	os.Stdin = stdinReader
	t.Cleanup(func() {
		if err := stdinReader.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Errorf("failed to close stdin reader: %v", err)
		}
	})

	tempDir := t.TempDir()

	// MCP stdio transport requires real file descriptors, so we use temp files
	// instead of in-memory buffers.
	stdoutFile, err := os.CreateTemp(tempDir, "stdout")
	if err != nil {
		t.Fatalf("failed to create stdout file: %v", err)
	}
	t.Cleanup(func() {
		if err := stdoutFile.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Errorf("failed to close stdout file: %v", err)
		}
	})

	stderrFile, err := os.CreateTemp(tempDir, "stderr")
	if err != nil {
		t.Fatalf("failed to create stderr file: %v", err)
	}
	t.Cleanup(func() {
		if err := stderrFile.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Errorf("failed to close stderr file: %v", err)
		}
	})

	os.Stdout = stdoutFile
	os.Stderr = stderrFile

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	errCh := make(chan error, 1)
	go func() {
		errCh <- runServe(cmd)
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatalf("expected error when context is canceled")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("runServe did not return within timeout")
	}
}
