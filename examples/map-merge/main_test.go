package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_PrintsRFC7386MapBehaviors(t *testing.T) {
	// Capturing stdout mutates process-wide state.
	output := captureStdout(t, main)

	assert.Contains(t, output, "=== Map Merge Example ===")
	result := resultSection(output)
	assert.Contains(t, result, `"city": "Boston"`)
	assert.Contains(t, result, `"street": "123 Main St"`)
	assert.Contains(t, result, `"email": "john@example.com"`)
	assert.Contains(t, result, `"hobbies": [`)
	assert.Contains(t, result, `"hiking"`)
	assert.Contains(t, result, `"coding"`)
	assert.NotContains(t, result, `"country": "USA"`)
	assert.Contains(t, output, "Objects are merged recursively")
	assert.Contains(t, output, "null values delete fields")
	assert.Contains(t, output, "Arrays are replaced entirely")
	assert.Contains(t, output, "New fields are added")
}

func TestPrettyJSON_FormatsMaps(t *testing.T) {
	t.Parallel()

	got := prettyJSON(map[string]any{"name": "Jane", "age": 31})

	assert.Contains(t, got, "\n")
	assert.Contains(t, got, `"name": "Jane"`)
	assert.Contains(t, got, `"age": 31`)
}

func resultSection(output string) string {
	_, result, _ := strings.Cut(output, "\nResult:   ")
	result, _, _ = strings.Cut(result, "\n\n=== Key Behaviors ===")
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
