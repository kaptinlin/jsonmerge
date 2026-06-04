# Overview

## Purpose

`jsonmerge` implements RFC 7386 JSON Merge Patch for Go with explicit patch values and one JSON semantic model across supported document forms.
The package applies and diffs JSON values while returning the requested Go result type only when that projection is lossless.

> **Why**: Merge-patch code often sits on configuration and API boundaries. Silent guessing or silent data loss is more expensive than an explicit constructor or a returned error.
>
> **Rejected**: Implicit patch/document overloading, representation-specific merge rules, and lossy typed projection.

## Scope

The package defines four operations:

- `Parse` parses encoded JSON text into a `Patch`.
- `NewPatch` converts a Go value into a `Patch`.
- `Apply` applies a `Patch` to a document.
- `Diff` derives a `Patch` that transforms one JSON value into another.

Supported documents include `map[string]any`, `[]byte`, `JSON`, plain `string`, structs, typed Go values, scalar values, and `nil` when the result can be represented by the requested type.

> **Why**: RFC 7386 is intentionally small. The library should cover patch construction, patch application, patch generation, and representation projection, then stop.
>
> **Rejected**: Schema validation, persistence adapters, HTTP helpers, validation-only entry points, or domain-specific merge policies.

## Design Priorities

1. RFC fidelity before convenience.
2. Explicit data-form boundaries before clever coercion.
3. Lossless projection before type-preservation convenience.
4. Minimal public surface area.
5. Performance changes justified by benchmarks.

> **Why**: A small public API can stay stable only when each exported name earns its place.
>
> **Rejected**: Public wrappers, mutation options, inspection summaries, or strategy hooks without current evidence.

## Non-Goals

This package does not define:

- RFC 6902 JSON Patch operations
- Element-wise array merge semantics
- Schema-aware merge behavior
- Pluggable merge strategies
- File IO, CLI behavior, redaction, source precedence, or config overlay policy

> **Why**: Each feature changes the package from an RFC 7386 kernel into a broader transformation framework.
>
> **Rejected**: Configuration knobs for per-project merge rules.

## Forbidden

- Do not add public entry points that duplicate `Parse`, `NewPatch`, `Apply`, `Diff`, or `Patch.MarshalJSON`.
- Do not make plain `string` mean JSON text.
- Do not return a typed result after silently dropping JSON members.
- Do not introduce public mutation semantics.
- Do not introduce configurable array merge behavior; arrays replace as a whole.

## Acceptance Criteria

- The library purpose, priorities, and non-goals are documented without relying on `README.md`.
- A contributor can tell which problems belong in this package and which belong elsewhere.
- The overview does not duplicate detailed API or implementation rules from later specs.
