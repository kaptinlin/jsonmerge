# Domain Specs

## Merge Patch Semantics

The package follows RFC 7386 at the JSON value level:

- Object patches merge recursively.
- `null` deletes object members.
- Non-object patches replace the entire target.
- Arrays replace the entire target array.
- A `nil` map target behaves like JSON null; an object patch turns it into an object.

> **Why**: RFC 7386 is the domain contract. Go-specific ergonomics must not invent different merge rules.
>
> **Rejected**: Element-wise array merging, special delete markers, and type-specific merge exceptions.

## Supported Document Forms

The package accepts these document forms:

- `map[string]any` as an object document form
- `[]byte` containing valid JSON text
- `JSON` containing valid JSON text carried as a string
- plain `string` as a JSON string scalar
- other JSON-marshalable Go values, converted through JSON before apply or diff

Invalid JSON text in `[]byte` or `JSON` is rejected with `ErrInvalidJSON`.
Plain `string` is never parsed as JSON text; a JSON-looking string is still a JSON string scalar.

> **Why**: The call site must reveal whether a value is text, scalar data, or a patch. Guessing makes malformed JSON text look like valid user data.
>
> **Rejected**: Parse-or-raw string heuristics.

## Patch Rules

`Patch` values are immutable after construction:

- `Parse` constructs a patch from encoded JSON text.
- `NewPatch` constructs a patch from a Go value using the same canonical JSON value model as documents.
- `Patch.MarshalJSON` returns the encoded patch for storage or transport.
- The zero `Patch` value is the JSON `null` patch.

> **Why**: A patch should be explicit, validated, and safe to reuse across calls.
>
> **Rejected**: Passing arbitrary target-shaped values as patches at apply time.

## Typed Value Rules

Structs and other typed Go values follow their JSON encoding:

- Apply and diff operate on canonical JSON values.
- Sparse object patches are the normal way to update typed targets without struct zero-value ambiguity.
- Result projection must be lossless.
- Unknown object members fail projection unless the target type explicitly captures them through the JSON package.
- Missing members that would reappear during marshaling fail projection.

> **Why**: Returning a typed value after dropping JSON data is more dangerous than returning an error.
>
> **Rejected**: Silent unknown-member discard and best-effort typed projection.

## Diff Rules

`Diff` uses the same canonical JSON model as `Apply`:

- Non-object targets produce the target value as the patch.
- Object targets produce only changed or added members.
- Removed members are encoded as `nil`.
- Equal object documents produce an empty object patch.

> **Why**: A generated patch should round-trip through `Apply` in the canonical JSON model without extra normalization.
>
> **Rejected**: A separate raw-Go `deepEqual` model for diffing.

## Forbidden

- Do not merge arrays element by element.
- Do not accept malformed JSON text in `[]byte` or `JSON`.
- Do not parse plain `string` as JSON text.
- Do not promise lossless projection into a type that cannot express the merged JSON value.

## Acceptance Criteria

- Each supported representation has explicit acceptance and interpretation rules.
- The spec distinguishes JSON text from JSON string scalar values.
- Apply and diff semantics are documented for add, update, delete, replace, and object no-op cases.
