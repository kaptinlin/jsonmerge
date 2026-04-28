package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_PrintsMergedStructAndTypedAccess(t *testing.T) {
	// Capturing stdout mutates process-wide state.
	output := captureStdout(t, main)

	assert.Contains(t, output, "=== Struct Merge Example ===")
	assert.Contains(t, output, `Original: {"name":"John Doe","email":"john@example.com","age":30}`)
	assert.Contains(t, output, `Patch:    {"name":"Jane Doe","age":25}`)
	assert.Contains(t, output, `Result:   {"name":"Jane Doe","email":"john@example.com","age":25}`)
	assert.Contains(t, output, "Type-safe access: Jane Doe is 25 years old")
}

func TestPrettyJSON_FormatsStructs(t *testing.T) {
	t.Parallel()

	got := prettyJSON(User{Name: "Jane", Email: "jane@example.com", Age: 31})

	assert.Equal(t, `{"name":"Jane","email":"jane@example.com","age":31}`, got)
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
