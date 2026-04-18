# Domain Specs

## Merge Patch Semantics

The package follows RFC 7386 exactly at the JSON value level:

- Object patches merge recursively.
- `null` deletes object members.
- Non-object patches replace the entire target.
- Arrays replace the entire target array.
- A `nil` map target behaves like an empty object when the patch is an object.

> **Why**: RFC 7386 is the domain contract. Go-specific ergonomics must not invent different merge rules.
>
> **Rejected**: Element-wise array merging, special delete markers, and type-specific merge exceptions.

## Supported Document Forms

The package accepts these document forms:

- `map[string]any` as the canonical in-memory object form
- `[]byte` containing valid JSON
- `string`, parsed as JSON when valid and otherwise treated as a raw string scalar
- Other JSON-marshalable Go values, converted through JSON before merge or generation

Invalid JSON bytes are rejected. Invalid JSON strings are still accepted as string scalar values.

> **Why**: The package should keep a single semantic model and let conversion bridge Go values into JSON values.
>
> **Rejected**: Divergent acceptance rules for equivalent JSON content based only on the caller's storage type.

## Typed Value Rules

Structs and other typed Go values follow their JSON encoding:

- Merge and generation operate on the marshaled JSON form.
- Zero values in a struct patch are real values once marshaled.
- Callers that need omission-versus-zero-value distinction should use pointer fields or `map[string]any` patches.

> **Why**: JSON Merge Patch works on JSON documents, not on Go field metadata.
>
> **Rejected**: Reflection-only field-presence tracking that disagrees with JSON marshaling.

## Patch Generation Rules

`Generate` uses the same JSON model:

- Non-object targets produce the target value as the patch.
- Object targets produce only changed or added members.
- Removed members are encoded as `nil`.
- Equal object documents produce an empty object patch.

> **Why**: A generated patch should round-trip through `Merge` and land on the target document without extra normalization.
>
> **Rejected**: Verbose patches that restate unchanged fields or no-op generation that returns `nil` for equal objects.

## Forbidden

- Do not merge arrays element by element.
- Do not accept invalid JSON bytes as documents or patches.
- Do not promise omission semantics for non-pointer struct zero values.

## Acceptance Criteria

- Each supported representation has explicit acceptance and interpretation rules.
- The spec distinguishes invalid JSON bytes from raw string scalar behavior.
- Merge and generation semantics are documented for add, update, delete, replace, and no-op cases.

**Origin:** Split from the historical `CLAUDE.md` during the SPECS migration.
