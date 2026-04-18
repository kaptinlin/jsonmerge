# API Specs

## Merge Contract

`Merge[T Document](target, patch T, opts ...Option) (*Result[T], error)` applies a merge patch and returns the merged document wrapped in `Result[T]`.
The result preserves the caller's document type `T`.

By default, object merges avoid mutating map targets. A non-object patch replaces the target value entirely, even when the target started as an object.

> **Why**: Callers should be able to keep their existing document type and rely on RFC 7386 replacement behavior without a separate API.
>
> **Rejected**: In-place mutation by default and representation-specific merge entry points.

## Generate Contract

`Generate[T Document](source, target T) (T, error)` returns a patch that transforms `source` into `target` when reapplied with `Merge`.
The patch value is returned in the same type `T`.

> **Why**: Keeping generation and application in the same generic type makes round-trips predictable for callers.
>
> **Rejected**: Always returning `map[string]any` or another canonical patch type.

## Valid Contract

`Valid[T Document](patch T) bool` reports whether a value is accepted as a patch input by the package's conversion rules.
It is an acceptance check, not a schema or business-rule validator.

> **Why**: Callers often need a cheap gate before calling `Merge`, but RFC 7386 does not define domain validation.
>
> **Rejected**: Validation that inspects application-specific structure or semantics.

## Result and Options

`Result[T]` exposes the merged document as `Doc`.
`WithMutate(true)` permits in-place updates for map targets during object merges. It does not redefine RFC replacement semantics.

> **Why**: Mutation is a performance trade-off, not a semantic mode switch.
>
> **Rejected**: Option sets that alter merge rules or create separate mutable result types.

## Error Contract

Failures wrap one of these sentinel errors so callers can use `errors.Is`:

- `ErrMarshal`
- `ErrUnmarshal`
- `ErrConversion`

Context may be added around the sentinel, but the sentinel must remain in the error chain.
Invalid JSON bytes fail with an unmarshal error. Invalid JSON strings may still succeed because they are accepted as raw string scalar values.

> **Why**: Callers need stable error classes without losing local context.
>
> **Rejected**: Large error taxonomies and exact-string matching.

## Forbidden

- Do not return a different document type than the caller requested.
- Do not make `WithMutate` change RFC 7386 semantics.
- Do not treat `Valid` as a substitute for application-level validation.

## Acceptance Criteria

- The contract of `Merge`, `Generate`, `Valid`, `Result`, and `WithMutate` is documented in one place.
- Mutation, replacement, and error behavior are explicit.
- A caller can derive the correct `errors.Is` checks from this spec alone.

**Origin:** Split from the historical `CLAUDE.md` during the SPECS migration.
