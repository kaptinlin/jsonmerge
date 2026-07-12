# API Specs

## Patch Construction

`Parse(data []byte) (Patch, error)` parses encoded JSON text as a merge patch.
Malformed text, duplicate object names, and invalid UTF-8 return an error matching `ErrInvalidJSON`.

`NewPatch(value any) (Patch, error)` converts a Go value into a patch.
Values that cannot be represented as JSON return an error matching `ErrInvalidValue`.

`Patch.MarshalJSON() ([]byte, error)` encodes the patch as deterministic compact JSON.

> **Why**: Patch construction is the validation boundary. After construction, a `Patch` is safe to reuse.
>
> **Rejected**: `Valid` as a bool-only duplicate of patch construction.

## Apply Contract

`Apply[T any](target T, patch Patch) (T, error)` applies a merge patch and returns the requested Go type `T` when the merged JSON value can be represented without loss.

By default, `Apply` never mutates caller-owned maps. A non-object patch replaces the target JSON value entirely, even when the target started as an object.
Projection fails with `ErrCannotRepresent` when the requested type would narrow numbers, discard object members, turn JSON `null` into a non-nullable value, or otherwise marshal back to a different normalized JSON value.

> **Why**: Callers should get the type they asked for only when that type can honestly carry the result.
>
> **Rejected**: In-place mutation options, target-shaped patch arguments, and silent lossy projection.

## Diff Contract

`Diff(source, target any) (Patch, error)` returns a patch that transforms `source` into `target` in the normalized JSON model when RFC 7386 can represent the change.
It returns `ErrCannotRepresentPatch` when the target requires creating or replacing an object member with JSON `null`.
The returned patch is independent of the static type of either input.
`Diff` has no public options: arrays and scalar roots are replacement values, object roots are member patches, and equal object roots return `{}`.

> **Why**: A patch is its own concept. It should not be forced through the source or target type.
>
> **Rejected**: Returning generated patches as `map[string]any`, `[]byte`, `string`, or the source type; adding path, ignore, or optimization options to `Diff`.

## JSON Text Contract

`type JSON string` marks a string as encoded JSON text.
`[]byte` is also encoded JSON text.
Plain `string` is a JSON string scalar.
Encoded JSON number literals are preserved while carried as JSON text; typed Go numbers are limited by their Go type before conversion.

> **Why**: The data form must be visible at the call site.
>
> **Rejected**: Treating valid strings as JSON text and invalid strings as raw values.

## Error Contract

Failures wrap one of these sentinel errors so callers can use `errors.Is`:

- `ErrInvalidJSON`
- `ErrInvalidValue`
- `ErrCannotRepresent`
- `ErrCannotRepresentPatch`

Context may be added around the sentinel, but the sentinel must remain in the error chain.
Callers must not match exact error strings.

> **Why**: Stable error classes should describe user-visible problems, not the current marshal/unmarshal stage.
>
> **Rejected**: Implementation-stage sentinels and exact-string matching.

## Forbidden

- Do not add compatibility wrappers for removed APIs.
- Do not expose a mutable `Patch` value tree.
- Do not return `Result[T]` unless it carries current, proven semantics beyond `T`.
- Do not add public options until more than one real option is justified.
- Do not add path-based mutation entry points to this package.

## Acceptance Criteria

- The contract of `Patch`, `Parse`, `NewPatch`, `Apply`, `Diff`, `JSON`, and error sentinels is documented in one place.
- Mutation, replacement, string, and projection behavior are explicit.
- Deterministic patch encoding and JSON number preservation are explicit.
- A caller can derive the correct `errors.Is` checks from this spec alone.
