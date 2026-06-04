# Architecture Specs

## Package Layout

The library stays as a single package with a small file split:

- `merge.go` contains public operations, sentinel errors, canonicalization, RFC apply logic, diff logic, projection, and comparison helpers.
- `types.go` contains public `Patch` and `JSON` types.
- `merge_test.go` covers RFC compliance, string semantics, sparse typed patches, projection failures, immutability, canonical diffing, and benchmarks.
- `conversion_test.go` covers representation preservation for JSON text and named scalar types.

> **Why**: The full merge pipeline is easier to audit when the implementation remains visible in one package.
>
> **Rejected**: Internal subpackages, code generation, or reflection adapters hidden behind extra layers.

## Apply Pipeline

`Apply` follows this flow:

1. Convert the target into a canonical JSON value.
2. Apply the already-canonical `Patch` using the recursive RFC 7386 algorithm.
3. Project the merged JSON value back into the requested Go type.
4. Verify projection did not change the canonical JSON value.

> **Why**: A single canonical boundary keeps apply logic representation-agnostic and prevents silent projection loss.
>
> **Rejected**: Separate apply implementations for structs, maps, bytes, strings, and JSON text.

## Diff Pipeline

`Diff` converts source and target into canonical JSON values, computes a minimal merge patch for object targets, and returns a `Patch`.
Equal object documents produce an empty object patch.

> **Why**: Diff should be the inverse companion to `Apply`, not a different equality model.
>
> **Rejected**: Raw Go type-sensitive comparison for JSON values.

## Dependency Rules

Production dependencies are limited to:

- `github.com/go-json-experiment/json` for marshal and unmarshal operations

Tests may use `github.com/google/go-cmp/cmp` and `github.com/stretchr/testify`.

> **Why**: The package needs one JSON boundary and no cloning dependency now that canonicalization produces caller-independent values.
>
> **Rejected**: Utility dependencies for helpers that the standard library already covers.

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
- Do not change hot-path code without rerunning compliance tests and benchmarks.

## Acceptance Criteria

- A contributor can trace apply and diff from this spec to the implementation.
- Every dependency has a documented purpose.
- Performance guidance constrains both optimization work and semantic safety.
