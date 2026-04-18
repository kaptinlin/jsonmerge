# JSON Merge Patch

RFC 7386 JSON Merge Patch implementation for Go with type-safe generics, minimal API, and high performance.

## Commands

```bash
# Testing and verification
task test
task test-verbose
task test-coverage
task bench
task lint
task markdownlint
task yamllint
task verify

# Maintenance
task fmt
task vet
task deps
task clean
```

## SPECS Index

- `SPECS/00-overview.md` — package scope, priorities, and non-goals
- `SPECS/10-domain-specs.md` — RFC 7386 semantics and document-form rules
- `SPECS/20-api-specs.md` — public API, options, and error contracts
- `SPECS/40-architecture-specs.md` — package layout, execution pipeline, and dependencies
- `SPECS/50-coding-standards.md` — contribution rules, tests, and lint requirements

## Working Rules

- Work on `main`.
- Keep `README.md` user-facing and keep the canonical internal contract in `SPECS/`.
- Run `task test` and `task lint` for code changes.
- Run `task markdownlint` for markdown changes and `task yamllint` for YAML changes.
- Prefer package-local skills in `.agents/skills/` when they match the task.

## Agent Skills

Package-local skills in `.agents/skills/`:

- `agent-md-creating`
- `code-simplifying`
- `committing`
- `dependency-selecting`
- `go-best-practices`
- `linting`
- `modernizing`
- `ralphy-initializing`
- `ralphy-todo-creating`
- `readme-creating`
- `releasing`
- `testing`
