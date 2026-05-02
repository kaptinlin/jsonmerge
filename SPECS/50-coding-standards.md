# Coding Standards

## Language and API Discipline

Target Go 1.26+ and keep the public surface small.
New exported API should only appear when RFC 7386 support or representation-preserving ergonomics clearly require it.

> **Why**: The library wins by being easy to audit and hard to misuse.
>
> **Rejected**: Surface-area growth for convenience helpers that duplicate existing behavior.

## RFC Compliance Rules

Contributors must preserve these semantics:

- Object patches merge recursively.
- `null` deletes object members.
- Non-object patches replace the target.
- Arrays replace whole values.
- Invalid JSON bytes fail fast during conversion.

> **Why**: These rules are the product contract, not implementation details.
>
> **Rejected**: Convenience behavior that makes one edge case nicer while drifting from RFC 7386.

## Error and Conversion Rules

Use the minimal sentinel set `ErrMarshal`, `ErrUnmarshal`, and `ErrConversion`.
Wrap underlying failures with `%w`, keep messages lowercase, and add only enough context to identify the failing stage.
Prefer JSON conversion boundaries over custom representation-specific semantics.

> **Why**: Callers need stable error classes and contributors need one mental model for typed and untyped documents.
>
> **Rejected**: Unused sentinels, exact-string error matching, and bespoke conversion paths for individual types.

## Implementation Rules

Default behavior must preserve map inputs unless the caller opts into `WithMutate(true)`.
Keep hot-path code direct where measurements show a benefit, but preserve the current semantic contract across common and fallback code paths.
Benchmark before keeping a performance change.

> **Why**: Mutation and optimization are trade-offs, not excuses to weaken correctness.
>
> **Rejected**: Abstractions that hide the RFC algorithm and micro-optimizations that are not benchmarked.

## Test and Lint Requirements

For code changes, run `task test` and `task lint`.
For markdown changes, keep `README.md`, `CLAUDE.md`, and `SPECS/**` aligned.
For YAML changes such as `lefthook.yml`, run `task yamllint`.
Tests must keep coverage for RFC Appendix A behavior, mutation rules, conversion failures, raw-string handling, and benchmarks.

> **Why**: This package's main failure modes are semantic drift and unchecked tooling gaps.
>
> **Rejected**: Spec-file layout tests and stale docs that silently drift from the canonical contracts.

## Forbidden

- Do not add helper abstractions that obscure RFC semantics in hot code.
- Do not define unused sentinel errors.
- Do not let `SPECS/**` drift from package behavior.
- Do not add `_test.go` files that only police `SPECS/` layout or `CLAUDE.md` links.
- Do not keep stale rules when the code has intentionally moved on; rewrite the spec to match the better current contract.

## Acceptance Criteria

- Code and docs describe the same behavior.
- Pre-commit and task-based linting cover code and YAML.
- Contributors know which behavioral changes require new tests and benchmarks.

**Origin:** This file preserves the git history of the original `CLAUDE.md`; the rest of the `SPECS/` set was split from the same source.
