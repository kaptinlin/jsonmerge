package jsonmerge

import (
	"errors"
	"math"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRFCAppendixA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target string
		patch  string
		want   string
	}{
		{
			name:   "replace object member",
			target: `{"a":"b"}`,
			patch:  `{"a":"c"}`,
			want:   `{"a":"c"}`,
		},
		{
			name:   "add object member",
			target: `{"a":"b"}`,
			patch:  `{"b":"c"}`,
			want:   `{"a":"b","b":"c"}`,
		},
		{
			name:   "delete only member",
			target: `{"a":"b"}`,
			patch:  `{"a":null}`,
			want:   `{}`,
		},
		{
			name:   "delete one member",
			target: `{"a":"b","b":"c"}`,
			patch:  `{"a":null}`,
			want:   `{"b":"c"}`,
		},
		{
			name:   "replace array with scalar",
			target: `{"a":["b"]}`,
			patch:  `{"a":"c"}`,
			want:   `{"a":"c"}`,
		},
		{
			name:   "replace scalar with array",
			target: `{"a":"c"}`,
			patch:  `{"a":["b"]}`,
			want:   `{"a":["b"]}`,
		},
		{
			name:   "nested object merge",
			target: `{"a":{"b":"c"}}`,
			patch:  `{"a":{"b":"d","c":null}}`,
			want:   `{"a":{"b":"d"}}`,
		},
		{
			name:   "array replacement is whole value",
			target: `{"a":[{"b":"c"}]}`,
			patch:  `{"a":[1]}`,
			want:   `{"a":[1]}`,
		},
		{
			name:   "array patch replaces document",
			target: `["a","b"]`,
			patch:  `["c","d"]`,
			want:   `["c","d"]`,
		},
		{
			name:   "array patch replaces object",
			target: `{"a":"b"}`,
			patch:  `["c"]`,
			want:   `["c"]`,
		},
		{
			name:   "null patch replaces document",
			target: `{"a":"foo"}`,
			patch:  `null`,
			want:   `null`,
		},
		{
			name:   "string patch replaces document",
			target: `{"a":"foo"}`,
			patch:  `"bar"`,
			want:   `"bar"`,
		},
		{
			name:   "null target member is preserved",
			target: `{"e":null}`,
			patch:  `{"a":1}`,
			want:   `{"e":null,"a":1}`,
		},
		{
			name:   "object patch turns array target into object",
			target: `[1,2]`,
			patch:  `{"a":"b","c":null}`,
			want:   `{"a":"b"}`,
		},
		{
			name:   "deep null delete inside newly created object",
			target: `{}`,
			patch:  `{"a":{"bb":{"ccc":null}}}`,
			want:   `{"a":{"bb":{}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			patch := mustParsePatch(t, tt.patch)
			got, err := Apply(JSON(tt.target), patch)
			require.NoError(t, err)

			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestStringDocumentsAreScalars(t *testing.T) {
	t.Parallel()

	t.Run("string patch replaces string target", func(t *testing.T) {
		t.Parallel()

		patch := mustNewPatch(t, "published")
		got, err := Apply("draft", patch)
		require.NoError(t, err)

		assert.Equal(t, "published", got)
	})

	t.Run("malformed json text is not accepted by parse", func(t *testing.T) {
		t.Parallel()

		_, err := Parse([]byte(`{"name": invalid}`))
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrInvalidJSON))
	})

	t.Run("malformed json-looking string is a string value", func(t *testing.T) {
		t.Parallel()

		patch := mustNewPatch(t, `{"name": invalid}`)
		got, err := Apply("draft", patch)
		require.NoError(t, err)

		assert.Equal(t, `{"name": invalid}`, got)
	})

	t.Run("object result cannot project into string", func(t *testing.T) {
		t.Parallel()

		patch := mustNewPatch(t, map[string]any{"name": "Jane"})
		_, err := Apply("draft", patch)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrCannotRepresent))
	})
}

func TestJSONTextDocumentsAreExplicit(t *testing.T) {
	t.Parallel()

	patch := mustParsePatch(t, `{"name":"Jane"}`)
	got, err := Apply(JSON(`{"name":"John","age":30}`), patch)
	require.NoError(t, err)

	assert.JSONEq(t, `{"name":"Jane","age":30}`, string(got))
}

func TestInvalidJSONTextDocumentFails(t *testing.T) {
	t.Parallel()

	patch := mustParsePatch(t, `{"name":"Jane"}`)
	_, err := Apply(JSON(`{"name": invalid}`), patch)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidJSON))
}

func TestSparsePatchAppliesToTypedTarget(t *testing.T) {
	t.Parallel()

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
		Age   int    `json:"age"`
	}

	user := User{Name: "John", Email: "john@example.com", Age: 30}
	patch := mustNewPatch(t, map[string]any{"name": "Jane"})

	got, err := Apply(user, patch)
	require.NoError(t, err)

	assert.Equal(t, User{Name: "Jane", Email: "john@example.com", Age: 30}, got)
}

func TestProjectionMustBeLossless(t *testing.T) {
	t.Parallel()

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
		Age   int    `json:"age"`
	}

	user := User{Name: "John", Email: "john@example.com", Age: 30}

	t.Run("unknown member fails", func(t *testing.T) {
		t.Parallel()

		patch := mustNewPatch(t, map[string]any{"role": "admin"})
		_, err := Apply(user, patch)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrCannotRepresent))
	})

	t.Run("deleted non-omitempty field fails", func(t *testing.T) {
		t.Parallel()

		patch := mustNewPatch(t, map[string]any{"age": nil})
		_, err := Apply(user, patch)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrCannotRepresent))
	})

	t.Run("deleted omitempty field succeeds", func(t *testing.T) {
		t.Parallel()

		patch := mustNewPatch(t, map[string]any{"email": nil})
		got, err := Apply(user, patch)
		require.NoError(t, err)

		assert.Equal(t, User{Name: "John", Age: 30}, got)
	})
}

func TestDiffUsesCanonicalEquality(t *testing.T) {
	t.Parallel()

	type counter struct {
		N int `json:"n"`
	}

	tests := []struct {
		name   string
		source any
		target any
	}{
		{
			name:   "map and json text",
			source: map[string]any{"n": int(1)},
			target: JSON(`{"n":1}`),
		},
		{
			name:   "struct and bytes",
			source: counter{N: 1},
			target: []byte(`{"n":1}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			patch, err := Diff(tt.source, tt.target)
			require.NoError(t, err)

			data, err := patch.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, `{}`, string(data))
		})
	}
}

func TestDiffPatchRoundTripsAcrossRepresentations(t *testing.T) {
	t.Parallel()

	source := map[string]any{"name": "John", "age": int64(30)}
	target := JSON(`{"name":"Jane","age":30}`)

	patch, err := Diff(source, target)
	require.NoError(t, err)

	got, err := Apply(source, patch)
	require.NoError(t, err)

	assert.JSONEq(t, string(target), mustMarshalJSON(t, got))
}

func TestApplyDoesNotMutateCallerOwnedMaps(t *testing.T) {
	t.Parallel()

	target := map[string]any{
		"profile": map[string]any{
			"theme": "dark",
		},
	}
	patchInput := map[string]any{
		"profile": map[string]any{
			"theme": "light",
			"flags": []any{
				map[string]any{"name": "new"},
			},
		},
	}
	patch := mustNewPatch(t, patchInput)

	got, err := Apply(target, patch)
	require.NoError(t, err)

	got["profile"].(map[string]any)["theme"] = "changed"
	got["profile"].(map[string]any)["flags"].([]any)[0].(map[string]any)["name"] = "changed"

	wantTarget := map[string]any{
		"profile": map[string]any{
			"theme": "dark",
		},
	}
	if diff := cmp.Diff(wantTarget, target); diff != "" {
		t.Fatalf("Apply() mutated target (-want +got):\n%s", diff)
	}

	wantPatchInput := map[string]any{
		"profile": map[string]any{
			"theme": "light",
			"flags": []any{
				map[string]any{"name": "new"},
			},
		},
	}
	if diff := cmp.Diff(wantPatchInput, patchInput); diff != "" {
		t.Fatalf("Apply() mutated patch input (-want +got):\n%s", diff)
	}
}

func TestApplyDoesNotAliasPatchValues(t *testing.T) {
	t.Parallel()

	patch := mustNewPatch(t, map[string]any{
		"profile": map[string]any{
			"flags": []any{
				map[string]any{"name": "new"},
			},
		},
	})

	got, err := Apply(map[string]any{}, patch)
	require.NoError(t, err)

	got["profile"].(map[string]any)["flags"].([]any)[0].(map[string]any)["name"] = "changed"

	next, err := Apply(map[string]any{}, patch)
	require.NoError(t, err)

	want := map[string]any{
		"profile": map[string]any{
			"flags": []any{
				map[string]any{"name": "new"},
			},
		},
	}
	if diff := cmp.Diff(want, next); diff != "" {
		t.Fatalf("Apply() reused mutated patch values (-want +got):\n%s", diff)
	}
}

func TestInvalidGoValueFails(t *testing.T) {
	t.Parallel()

	_, err := NewPatch(map[string]any{"limit": math.NaN()})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidValue))
}

func TestMapProjectionRejectsNonObjectResults(t *testing.T) {
	t.Parallel()

	patch := mustNewPatch(t, "scalar")
	_, err := Apply(map[string]any{"name": "John"}, patch)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrCannotRepresent))
}

func BenchmarkApplyMap(b *testing.B) {
	target := map[string]any{
		"name": "John",
		"profile": map[string]any{
			"active": true,
			"limits": map[string]any{
				"requests": 100,
			},
		},
	}
	patch := mustNewPatch(b, map[string]any{
		"profile": map[string]any{
			"limits": map[string]any{
				"requests": 200,
			},
		},
	})

	b.ResetTimer()
	for b.Loop() {
		if _, err := Apply(target, patch); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDiffMap(b *testing.B) {
	source := map[string]any{
		"name": "John",
		"profile": map[string]any{
			"active": true,
			"limits": map[string]any{
				"requests": 100,
			},
		},
	}
	target := map[string]any{
		"name": "Jane",
		"profile": map[string]any{
			"active": true,
			"limits": map[string]any{
				"requests": 200,
			},
		},
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := Diff(source, target); err != nil {
			b.Fatal(err)
		}
	}
}

func mustParsePatch(tb testing.TB, text string) Patch {
	tb.Helper()

	patch, err := Parse([]byte(text))
	if err != nil {
		tb.Fatalf("Parse(%q) failed: %v", text, err)
	}
	return patch
}

func mustNewPatch(tb testing.TB, value any) Patch {
	tb.Helper()

	patch, err := NewPatch(value)
	if err != nil {
		tb.Fatalf("NewPatch(%T) failed: %v", value, err)
	}
	return patch
}

func mustMarshalJSON(tb testing.TB, value any) string {
	tb.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		tb.Fatalf("json.Marshal(%T) failed: %v", value, err)
	}
	return string(data)
}
