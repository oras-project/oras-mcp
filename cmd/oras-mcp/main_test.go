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

package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureCombinedOutput(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	originalStdout := os.Stdout
	originalStderr := os.Stderr

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = writer
	os.Stderr = writer

	outputCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		if _, copyErr := io.Copy(&buf, reader); copyErr != nil {
			// io.Copy only returns an error if the read or write fails; close errors are expected
			t.Logf("io.Copy error: %v", copyErr)
		}
		outputCh <- buf.String()
	}()

	errRun := fn()

	if closeErr := writer.Close(); closeErr != nil {
		t.Fatalf("failed to close writer: %v", closeErr)
	}

	os.Stdout = originalStdout
	os.Stderr = originalStderr

	output := <-outputCh

	if closeErr := reader.Close(); closeErr != nil {
		t.Fatalf("failed to close reader: %v", closeErr)
	}

	return output, errRun
}

func TestRunVersionCommand(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"oras-mcp", "version"}
	defer func() {
		os.Args = originalArgs
	}()

	output, err := captureCombinedOutput(t, run)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(output, "Version:") {
		t.Fatalf("expected output to contain version information, got: %q", output)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"oras-mcp", "nonexistent"}
	defer func() {
		os.Args = originalArgs
	}()

	_, err := captureCombinedOutput(t, run)
	if err == nil {
		t.Fatalf("expected an error for unknown command")
	}
}
