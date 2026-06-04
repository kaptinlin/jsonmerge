package jsonmerge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyPreservesJSONTextDocumentTypes(t *testing.T) {
	t.Parallel()

	t.Run("bytes", func(t *testing.T) {
		t.Parallel()

		patch := mustParsePatch(t, `{"name":"Jane"}`)
		got, err := Apply([]byte(`{"name":"John","age":30}`), patch)
		require.NoError(t, err)

		assert.JSONEq(t, `{"name":"Jane","age":30}`, string(got))
	})

	t.Run("json text type", func(t *testing.T) {
		t.Parallel()

		patch := mustParsePatch(t, `{"name":"Jane"}`)
		got, err := Apply(JSON(`{"name":"John","age":30}`), patch)
		require.NoError(t, err)

		assert.JSONEq(t, `{"name":"Jane","age":30}`, string(got))
	})
}

type revision int

func TestScalarDocumentsPreserveNamedType(t *testing.T) {
	t.Parallel()

	t.Run("apply", func(t *testing.T) {
		t.Parallel()

		patch := mustNewPatch(t, revision(2))
		got, err := Apply(revision(1), patch)
		require.NoError(t, err)

		assert.Equal(t, revision(2), got)
	})

	t.Run("diff", func(t *testing.T) {
		t.Parallel()

		patch, err := Diff(revision(1), revision(2))
		require.NoError(t, err)

		data, err := patch.MarshalJSON()
		require.NoError(t, err)
		assert.JSONEq(t, `2`, string(data))
	})
}
