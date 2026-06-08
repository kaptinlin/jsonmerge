# JSON Merge Patch

RFC 7386 JSON Merge Patch implementation for Go with explicit `Patch` values, one canonical JSON value model, and lossless result projection.

For installation and usage examples, see [README.md](README.md). For package contracts, read [`SPECS/`](SPECS/).

## Commands

```bash
task test            # Run all tests with race detection
task test-verbose    # Run tests with verbose output
task test-coverage   # Generate coverage.out and coverage.html
task bench           # Run benchmarks
task lint            # Run golangci-lint and go.mod/go.sum tidy checks
task yamllint        # Lint YAML files
task vuln            # Run govulncheck
task fmt             # Format Go code
task vet             # Run go vet
task deps            # Download and tidy dependencies
task deps:update     # Update dependencies and tidy modules
task clean           # Remove bin/ and Go caches
task verify          # Run deps, fmt, vet, lint, test, and vuln
```

## Architecture

```text
jsonmerge/
├── merge.go         # Public API, sentinels, canonicalization, RFC apply/diff, projection
├── types.go         # Patch and JSON public types
├── merge_test.go    # RFC compliance, string semantics, projection, immutability, benchmarks
├── conversion_test.go
├── example_test.go  # Executable examples checked by go test
├── examples/        # Runnable demos for map, struct, JSON text, and byte documents
└── SPECS/           # Canonical package contracts and coding standards
```

## Agent Workflow

### Design Phase — Read SPECS First

Before changing code or docs, read the relevant `SPECS/` documents first. `README.md` is user-facing; `SPECS/` defines the package contract.

Workflow:

1. Identify the relevant spec files from the index below.
2. Verify the current code matches the spec before updating docs.
3. If code and spec intentionally change, update the spec and code together.
4. Keep `AGENTS.md` as a symlink to `CLAUDE.md`.

## Agent Operating Rules

- Read the relevant SPECS before changing code or docs.
- Prefer the smallest direct change that satisfies the current contract.
- Keep edits surgical and avoid unrelated refactors.
- Verify behavior with tests that exercise public contracts, not spec text.
- Fail loudly with sentinel errors instead of silent coercion or data loss.
- Preserve user worktree changes you did not make.
- Do not add policy-only gate scripts that restate docs or SPECS.
- Do not add redundant tests that only mirror SPECS after behavior is covered.

## SPECS Index

| Spec | Topic |
| --- | --- |
| [`SPECS/00-overview.md`](SPECS/00-overview.md) | Package scope, priorities, and non-goals |
| [`SPECS/10-domain-specs.md`](SPECS/10-domain-specs.md) | RFC 7386 semantics and accepted document forms |
| [`SPECS/20-api-specs.md`](SPECS/20-api-specs.md) | Public API and error contracts |
| [`SPECS/40-architecture-specs.md`](SPECS/40-architecture-specs.md) | Package layout, execution pipeline, and dependencies |
| [`SPECS/50-coding-standards.md`](SPECS/50-coding-standards.md) | Contribution rules, tests, and lint requirements |

## Design Philosophy

- **KISS** — Keep one RFC 7386 kernel and one package. Prefer direct code paths over representation-specific layers.
- **YAGNI** — Stop at patch construction, apply, diff, and JSON marshaling. No schema logic, transport helpers, mutation options, or strategy frameworks.
- **SRP** — `merge.go` owns the merge pipeline; `types.go` owns public value types; `SPECS/` owns the written contract.
- **Simplicity as art** — `Patch`, `Parse`, `NewPatch`, `Apply`, and `Diff` are the public shape. Add nothing unless the current contract needs it.
- **Errors as teachers** — Wrap failures with stable user-semantic sentinels and enough stage context to locate the boundary.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Explicit patches**: callers construct `Patch` with `Parse` or `NewPatch`; `Apply` never accepts target-shaped patch values.
- **Visible data forms**: plain `string` is a JSON string scalar; `[]byte` and `JSON` are encoded JSON text.
- **Lossless projection**: `Apply[T]` returns `T` only when `T` can represent the merged JSON value without silently dropping data.
- **Pure default**: public apply paths do not mutate caller-owned maps.

## Coding Rules

### Must Follow

- Use the Go version declared in `go.mod`; use modern standard library helpers where they simplify code.
- Follow [Google Go Best Practices](https://google.github.io/go-style/best-practices).
- Follow [Google Go Style Decisions](https://google.github.io/go-style/decisions).
- KISS/DRY/YAGNI — keep the package small, direct, and free of speculative APIs.
- Keep `README.md` user-facing and keep canonical contributor rules in `CLAUDE.md` and `SPECS/`.
- Preserve RFC 7386 semantics across all document forms: object merge, `null` delete, non-object replace, and whole-array replace.
- Keep `Patch` immutable and reusable.
- Keep plain `string` as a JSON string scalar; use `[]byte` or `JSON` for encoded JSON text.
- Return `ErrCannotRepresent` when result projection would silently lose JSON data.
- Wrap failures with `ErrInvalidJSON`, `ErrInvalidValue`, or `ErrCannotRepresent`, and keep the sentinel in the error chain with `%w`.
- Benchmark hot-path changes and rerun compliance tests when touching apply, diff, canonicalization, projection, or comparison logic.
- Keep `AGENTS.md` as a symlink to `CLAUDE.md`; do not duplicate the file.

### Forbidden

- No `panic` in production code — return errors instead.
- No compatibility wrappers for removed APIs.
- No premature abstraction — three similar lines are better than a helper used once.
- No feature creep — only implement what RFC 7386 support or representation-preserving ergonomics require.
- No element-wise array merges, custom delete markers, or schema-aware merge rules.
- No parse-or-raw string heuristics.
- No public mutation option unless a future benchmark-backed API is intentionally designed.
- No behavior drift between map, struct, byte-slice, `JSON`, and string forms unless the JSON-level contract changes for all of them.
- No documentation masquerading as code — keep contract prose in `SPECS/`, not in unused runtime flags or tables.
- No policy-only gate scripts that merely restate README, SPECS, or AGENTS rules.
- No spec mirror tests that assert documentation text instead of user-visible behavior.
- No working around dependency bugs — if a dependency blocks work, write `reports/<dependency-name>.md` instead of reimplementing it inline.

## Testing

- Use Go's `testing` package with `testify/assert` and `testify/require` in package tests.
- Use `go-cmp` for structural map/slice comparisons.
- Keep coverage for RFC Appendix A behavior, string scalar versus JSON text, invalid JSON text, invalid Go values, lossless projection, map immutability, canonical diffing, examples, and benchmarks.
- Keep runnable examples in `example_test.go` aligned with `README.md`.
- Run `task test` and `task lint` for code changes.
- Run `task yamllint` for YAML changes such as `lefthook.yml`.

## Dependencies

- `github.com/go-json-experiment/json` — marshal and unmarshal at the conversion boundary.
- `github.com/kaptinlin/deepclone` — clone canonical JSON values while preserving patch immutability.
- `github.com/google/go-cmp` — structural comparisons in tests.
- `github.com/stretchr/testify` — test assertions only.

## Performance

- Public `Apply` is pure by default and must preserve caller-owned values.
- Optimize common JSON primitives and containers only when benchmarks justify it.
- Run `task bench` after changing hot-path logic.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. Do not work around it by reimplementing the dependency's functionality.
2. Do not skip the dependency and write a local replacement.
3. Create a report file in `reports/<dependency-name>.md`.
4. Include the dependency version, trigger scenario, expected behavior, actual behavior, relevant errors, and any non-code workaround suggestion.
5. Continue with tasks that do not depend on the broken behavior.

## Agent Skills

Specialized skills are available through `.claude/skills/` and `.agents/skills/`:

| Skill | When to Use |
| --- | --- |
| [`agent-md-writing`](.agents/skills/agent-md-writing/) | Refresh `CLAUDE.md` and the `AGENTS.md` symlink after workflow or contract changes |
| [`readme-writing`](.agents/skills/readme-writing/) | Refresh installation, quick start, and usage docs in `README.md` |
| [`golangci-linting`](.agents/skills/golangci-linting/) | Configure golangci-lint or fix lint failures |
| [`modernizing`](.agents/skills/modernizing/) | Adopt newer Go language and standard library features safely |
| [`committing`](.agents/skills/committing/) | Create conventional commits for this library |
| [`code-simplifying`](.agents/skills/code-simplifying/) | Simplify recent code without changing behavior |
| [`go-best-practices`](.agents/skills/go-best-practices/) | Apply Go style, testing, interface, and error-handling guidance |
| [`dependency-selecting`](.agents/skills/dependency-selecting/) | Evaluate or replace third-party Go dependencies |
| [`taskfile-configuring`](.agents/skills/taskfile-configuring/) | Update task-based workflows in `Taskfile.yml` |
| [`golang-design-guide`](.agents/skills/golang-design-guide/) | Design or review public Go library APIs |
