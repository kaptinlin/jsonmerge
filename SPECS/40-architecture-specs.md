# Architecture Specs

## Package Layout

The library stays as a single package with a small file split:

- `merge.go` contains public operations, sentinel errors, JSON normalization, RFC apply logic, diff logic, projection, and comparison helpers.
- `types.go` contains public `Patch` and `JSON` types.
- `merge_test.go` covers RFC compliance, string semantics, JSON number preservation, deterministic patch encoding, sparse typed patches, projection failures, immutability, normalized diffing, diff law, and benchmarks.
- `conversion_test.go` covers representation preservation for JSON text and named scalar types.

> **Why**: The full merge pipeline is easier to audit when the implementation remains visible in one package.
>
> **Rejected**: Internal subpackages, code generation, or reflection adapters hidden behind extra layers.

## Normalized JSON Boundary

All public inputs enter a private normalized JSON value model before merge or diff logic runs.
The model is private: callers can construct, apply, diff, and marshal patches, but they cannot inspect or mutate the value tree directly.

- Encoded JSON text enters through a decoder that preserves number literals.
- Go values enter by marshaling to deterministic JSON and parsing through the same JSON text boundary.
- Patch encoding uses deterministic compact JSON.
- Equality compares normalized JSON values, not Go static types.

> **Why**: The package needs one semantic boundary without turning into a public JSON AST package.
>
> **Rejected**: Public value trees, parser rewrites, and per-representation merge paths.

## Apply Pipeline

`Apply` follows this flow:

1. Convert the target into a normalized JSON value.
2. Apply the normalized `Patch` using the recursive RFC 7386 algorithm.
3. Project the merged JSON value back into the requested Go type.
4. Verify projection did not change the normalized JSON value.

> **Why**: A single normalized JSON boundary keeps apply logic representation-agnostic and prevents silent projection loss.
>
> **Rejected**: Separate apply implementations for structs, maps, bytes, strings, and JSON text.

## Diff Pipeline

`Diff` converts source and target into normalized JSON values, computes a minimal merge patch for object targets, and returns a `Patch`.
Equal object documents produce an empty object patch.
Equal scalar and array roots return the root value because a merge patch only has object-level no-op syntax.

> **Why**: Diff should be the inverse companion to `Apply`, not a different equality model.
>
> **Rejected**: Raw Go type-sensitive comparison for JSON values, public diff options, or array element diffing.

## Dependency Rules

Production dependencies are limited to:

- `github.com/go-json-experiment/json` for marshal and unmarshal operations, including its v1 decoder compatibility layer for raw JSON number literals
- `github.com/kaptinlin/deepclone` for cloning normalized JSON values without local clone logic

Tests may use `github.com/google/go-cmp/cmp` and `github.com/stretchr/testify`.

> **Why**: The package needs one JSON boundary and one reusable cloning boundary so `Patch` values stay immutable without maintaining package-local clone code.
>
> **Rejected**: Ad-hoc utility helpers for behavior owned by a focused dependency.

## Performance Architecture

Correctness stays ahead of micro-optimizations:

- Keep the RFC apply path straightforward and allocation-aware.
- Keep `Patch` reusable and immutable.
- Benchmark hot-path changes before keeping them.
- Do not expose mutation as a public optimization without real caller and benchmark evidence.

> **Why**: The package is small enough that performance work should stay legible and measurable.
>
> **Rejected**: Speculative tuning that obscures the RFC algorithm or changes behavior.

## Forbidden

- Do not split apply or diff across layers that hide RFC behavior.
- Do not add dependencies outside conversion or test support without a clear contract need.
- Do not expose the normalized value tree publicly.
- Do not change hot-path code without rerunning compliance tests and benchmarks.

## Acceptance Criteria

- A contributor can trace apply and diff from this spec to the implementation.
- Every dependency has a documented purpose.
- Performance guidance constrains both optimization work and semantic safety.
