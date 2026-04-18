# JSON Merge Patch

RFC 7386 JSON Merge Patch implementation for Go with a small generic API. The package keeps one semantic model across maps, structs, JSON bytes, and JSON strings while returning the same document type it receives.

For usage examples, installation, and API-oriented guidance, see [README.md](README.md).

## Commands

```bash
task test            # Run all tests with race detection
task test-verbose    # Run tests with verbose output
task test-coverage   # Generate coverage.out and coverage.html
task bench           # Run benchmarks
task lint            # Run golangci-lint and go.mod/go.sum tidy checks
task markdownlint    # Lint Markdown files
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
├── merge.go         # Public API, sentinel errors, RFC 7386 merge/generate pipeline
├── types.go         # Document constraint, Result, Options, and Option helpers
├── merge_test.go    # RFC compliance, conversion, mutation, concurrency, benchmarks
├── example_test.go  # Executable examples checked by go test
├── examples/        # Runnable demos for map, struct, string, and byte documents
└── SPECS/           # Canonical package contracts and coding standards
```

## Agent Workflow

### Design Phase — Read SPECS First

Before changing code or docs, read the relevant `SPECS/` documents first. `README.md` is user-facing; `SPECS/` defines the package contract.

Workflow:

1. Identify the relevant spec files from the index below.
2. Verify the current code matches the spec before updating docs.
3. If code and spec intentionally changed, update the spec and code together instead of documenting stale behavior.
4. Keep `AGENTS.md` as a symlink to `CLAUDE.md`.

## SPECS Index

Specification documents in [`SPECS/`](SPECS/) — package contracts, data-form rules, and coding standards:

| Spec | Topic |
| --- | --- |
| [`SPECS/00-overview.md`](SPECS/00-overview.md) | Package scope, priorities, and non-goals |
| [`SPECS/10-domain-specs.md`](SPECS/10-domain-specs.md) | RFC 7386 semantics and accepted document forms |
| [`SPECS/20-api-specs.md`](SPECS/20-api-specs.md) | Public API, options, and error contracts |
| [`SPECS/40-architecture-specs.md`](SPECS/40-architecture-specs.md) | Package layout, execution pipeline, and dependencies |
| [`SPECS/50-coding-standards.md`](SPECS/50-coding-standards.md) | Contribution rules, tests, and lint requirements |

## Design Philosophy

- **KISS** — Keep one merge model and one package. Prefer direct code paths over representation-specific layers.
- **YAGNI** — Stop at RFC 7386: merge application, patch generation, and acceptance checks. No schema logic, transport helpers, or custom merge strategies.
- **SRP** — `merge.go` owns the merge pipeline, `types.go` owns public types and options, and `SPECS/` owns the written contract.
- **Simplicity as art** — `Merge`, `Generate`, and `Valid` are the public document operations; `WithMutate(true)` is the only performance knob.
- **Errors as teachers** — Wrap failures with stable sentinels and enough stage context to show whether conversion, marshal, or unmarshal failed.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive Disclosure**: The common path is `Merge`, `Generate`, and `Valid`; advanced callers can opt into in-place map updates with `WithMutate(true)`.

## Coding Rules

### Must Follow

- Go 1.26.2 — use modern standard library helpers where they simplify code.
- Follow [Google Go Best Practices](https://google.github.io/go-style/best-practices)
- Follow [Google Go Style Decisions](https://google.github.io/go-style/decisions)
- KISS/DRY/YAGNI — keep the package small, direct, and free of speculative APIs.
- Keep `README.md` user-facing and keep canonical contributor rules in `CLAUDE.md` and `SPECS/`.
- Preserve RFC 7386 semantics across all document forms: object merge, `null` delete, non-object replace, and whole-array replace.
- Keep `Merge`, `Generate`, and `Valid` representation-preserving: the caller gets back the same static type `T`.
- Keep default map merges immutable; only `WithMutate(true)` may update `map[string]any` targets in place.
- Wrap failures with `ErrMarshal`, `ErrUnmarshal`, or `ErrConversion`, and keep the sentinel in the error chain with `%w`.
- Benchmark hot-path changes and rerun compliance tests when touching merge, generation, conversion, or comparison logic.
- Keep `AGENTS.md` as a symlink to `CLAUDE.md`; do not duplicate the file.

### Forbidden

- No `panic` in production code — return errors instead.
- No premature abstraction — three similar lines are better than a helper used once.
- No feature creep — only implement what RFC 7386 support or representation-preserving ergonomics require.
- No element-wise array merges, custom delete markers, or schema-aware merge rules.
- No behavior drift between map, struct, byte-slice, and string document forms unless the JSON-level contract stays identical.
- No documentation masquerading as code — keep contract prose in `SPECS/`, not in unused runtime flags or tables.
- No working around dependency bugs — if a dependency blocks work, write `reports/<dependency-name>.md` instead of reimplementing it inline.

## Testing

- Use Go's `testing` package with `testify/assert` in package tests.
- Keep coverage for RFC 7386 Appendix A behavior, mutation rules, conversion failures, raw-string versus invalid-byte handling, concurrency, and benchmarks.
- Keep runnable examples in `example_test.go` aligned with `README.md`.
- Run `task test` and `task lint` for code changes.
- Run `task markdownlint` for Markdown changes.
- Run `task yamllint` for YAML changes such as `lefthook.yml`.

## Dependencies

- `github.com/go-json-experiment/json` — marshal and unmarshal at the conversion boundary.
- `github.com/kaptinlin/deepclone` — preserve immutable default behavior for `map[string]any` object merges.
- `github.com/stretchr/testify` — test assertions only.

## Performance

- `WithMutate(true)` is a performance option for `map[string]any` object merges, not a semantic mode.
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
