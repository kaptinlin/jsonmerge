package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_PrintsMergedJSONBytesAndGeneratedPatch(t *testing.T) {
	// Capturing stdout mutates process-wide state.
	output := captureStdout(t, main)

	assert.Contains(t, output, "=== JSON Bytes Merge Example ===")
	result := resultSection(output)
	assert.Contains(t, result, `"phone":"+1-555-0123"`)
	assert.Contains(t, result, `"theme":"dark"`)
	assert.Contains(t, result, `"version":"1.1"`)
	assert.NotContains(t, result, `"email":"john@example.com"`)
	assert.Contains(t, output, "=== Generate Patch ===")
	assert.JSONEq(t, `{"age":26,"city":"Boston"}`, generatedPatch(output))
}

func generatedPatch(output string) string {
	_, patch, _ := strings.Cut(output, "Generated Patch: ")
	patch, _, _ = strings.Cut(patch, "\n")
	return patch
}

func resultSection(output string) string {
	_, result, _ := strings.Cut(output, "\nResult:\n")
	result, _, _ = strings.Cut(result, "\n\n=== Generate Patch ===")
	return result
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	require.NoError(t, writer.Close())
	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())

	return string(output)
}
