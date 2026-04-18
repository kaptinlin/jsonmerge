# Overview

## Purpose

`jsonmerge` implements [RFC 7386](https://datatracker.ietf.org/doc/html/rfc7386) for Go and keeps one API across map, struct, byte-slice, and string documents.
The package returns the same document type it receives, so callers can apply merge patches without rewriting their storage format first.

> **Why**: Go code commonly moves the same JSON payload between typed structs, raw bytes, strings, and `map[string]any`. The library should keep RFC 7386 semantics stable across those representations.
>
> **Rejected**: Separate APIs per representation, transport-specific helpers, and custom patch dialects.

## Scope

The package defines three operations:

- `Merge` applies a merge patch.
- `Generate` derives a merge patch that transforms one document into another.
- `Valid` reports whether a value is accepted as a patch input.

Supported documents include `map[string]any`, JSON-marshalable Go values such as structs, `[]byte`, `string`, and scalar patch values.

> **Why**: RFC 7386 is intentionally small. The library should cover patch application, patch generation, and patch acceptance, then stop.
>
> **Rejected**: Schema validation, persistence adapters, HTTP helpers, or domain-specific merge policies.

## Design Priorities

1. RFC fidelity before convenience.
2. Safe-by-default behavior for mutable Go maps.
3. Minimal public surface area.
4. Performance changes justified by benchmarks.

> **Why**: Merge-patch code sits on configuration and API boundaries, where semantic surprises cost more than an extra branch or allocation.
>
> **Rejected**: Convenience behavior that makes one representation easier at the cost of cross-type consistency.

## Non-Goals

This package does not define:

- RFC 6902 JSON Patch operations
- Element-wise array merge semantics
- Schema-aware merge behavior
- Pluggable merge strategies

> **Why**: Each of those features changes the contract from "RFC 7386 for Go" to a broader transformation framework.
>
> **Rejected**: Configuration knobs for per-project merge rules.

## Forbidden

- Do not add new public entry points for behavior already covered by `Merge`, `Generate`, or `Valid`.
- Do not change semantics for one document representation without changing the JSON-level contract for all of them.
- Do not introduce configurable array merge behavior; arrays replace as a whole.

## Acceptance Criteria

- The library purpose, priorities, and non-goals are documented without relying on `README.md`.
- A contributor can tell which problems belong in this package and which belong elsewhere.
- The overview does not duplicate detailed API or implementation rules from later specs.

**Origin:** Split from the historical `CLAUDE.md` during the SPECS migration.
