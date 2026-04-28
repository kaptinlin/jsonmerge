package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_PrintsMergedJSONStringAndValidation(t *testing.T) {
	// Capturing stdout mutates process-wide state.
	output := captureStdout(t, main)

	assert.Contains(t, output, "=== JSON String Merge Example ===")
	assert.Contains(t, output, `"age":31`)
	assert.Contains(t, output, `"skills":["Go","JavaScript","Rust"]`)
	assert.Contains(t, output, `"country":"USA"`)
	assert.Contains(t, output, `"email":"john@example.com"`)
	assert.Contains(t, output, "=== Validation ===")
	assert.Contains(t, output, "Valid patch: true")
	assert.Contains(t, output, "Invalid patch: true")
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
