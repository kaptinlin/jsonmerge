package jsonmerge

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestGeneratePreservesRawJSONDocumentTypes(t *testing.T) {
	t.Parallel()

	t.Run("bytes", func(t *testing.T) {
		t.Parallel()
		patch, err := Generate([]byte(`{"name":"John","age":30}`), []byte(`{"name":"Jane","age":30}`))
		require.NoError(t, err)

		var got map[string]any
		require.NoError(t, json.Unmarshal(patch, &got))
		if diff := cmp.Diff(map[string]any{"name": "Jane"}, got); diff != "" {
			t.Errorf("Generate() patch mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		patch, err := Generate(`{"name":"John","age":30}`, `{"name":"Jane","age":30}`)
		require.NoError(t, err)

		var got map[string]any
		require.NoError(t, json.Unmarshal([]byte(patch), &got))
		if diff := cmp.Diff(map[string]any{"name": "Jane"}, got); diff != "" {
			t.Errorf("Generate() patch mismatch (-want +got):\n%s", diff)
		}
	})
}
