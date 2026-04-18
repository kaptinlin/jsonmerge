# Architecture Specs

## Package Layout

The library stays as a single package with a small file split:

- `merge.go` contains the public operations, sentinel errors, RFC merge logic, generation logic, and conversion helpers.
- `types.go` contains the generic document constraint, result wrapper, and option types.
- `merge_test.go` covers compliance, conversion, mutation, concurrency, and benchmarks.

> **Why**: The full merge pipeline is easier to audit when the implementation remains visible in one package.
>
> **Rejected**: Internal subpackages, code generation, or reflection adapters hidden behind extra layers.

## Merge Execution Pipeline

`Merge` follows this flow:

1. Convert `target` and `patch` into JSON-compatible values.
2. Clone a map target only when needed to preserve immutable default behavior for object merges.
3. Apply the recursive RFC 7386 algorithm.
4. Convert the merged value back into the caller's document type.

> **Why**: A single conversion boundary keeps the merge algorithm representation-agnostic.
>
> **Rejected**: Separate merge implementations for structs, maps, bytes, and strings.

## Generation Pipeline

`Generate` converts the source and target into JSON-compatible values, computes the minimal merge patch, and converts that patch back into the caller's document type.
Equal object documents produce an empty object patch rather than `nil`.

> **Why**: Generation should be the inverse companion to `Merge`, not a different contract.
>
> **Rejected**: Generation rules that require post-processing before the patch can be merged back.

## Dependency Rules

Production dependencies are limited to:

- `github.com/go-json-experiment/json` for marshal and unmarshal operations
- `github.com/kaptinlin/deepclone` for immutable default map merges

Tests may use `github.com/stretchr/testify`.

> **Why**: Each dependency has one clear responsibility tied to conversion or safety.
>
> **Rejected**: Utility dependencies for helpers that the standard library already covers.

## Performance Architecture

The implementation may optimize common JSON types directly, but correctness stays ahead of micro-optimizations:

- Keep the RFC merge path straightforward and allocation-aware.
- Optimize common primitives and containers before falling back to generic comparison.
- Benchmark hot-path changes before keeping them.

> **Why**: The package is small enough that performance work should stay legible and measurable.
>
> **Rejected**: Speculative tuning that obscures the RFC algorithm or changes behavior.

## Forbidden

- Do not split the merge or generation pipeline across layers that hide the RFC behavior.
- Do not add dependencies outside conversion, cloning, or test support without a clear contract need.
- Do not change hot-path code without rerunning compliance tests and benchmarks.

## Acceptance Criteria

- A contributor can trace the merge and generation flow from this spec to the implementation.
- Every dependency has a documented purpose.
- Performance guidance constrains both optimization work and semantic safety.

**Origin:** Split from the historical `CLAUDE.md` during the SPECS migration.
