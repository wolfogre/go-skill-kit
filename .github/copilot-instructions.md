# Copilot Instructions

## Build & run

```bash
# Build all binaries
go build ./cmd/...

# Build a specific tool
cd cmd/gocoverage && go build .

# Run a specific tool directly
go run ./cmd/gocoverage coverage.out
```

No tests exist yet.

## Architecture

Each skill is a pair of two things that share the same name:

- `cmd/<name>/main.go` — a standalone Go binary (no shared packages, all logic lives in `package main`)
- `skills/<name>/SKILL.md` — a Copilot skill file that tells an AI agent when and how to invoke the binary

The `skills/` directory is meant to be pointed at by a Copilot agent configuration. The `description` field in each `SKILL.md` frontmatter is the triggering mechanism.

## Conventions

**Adding a new skill/tool**: create both `cmd/<name>/` and `skills/<name>/SKILL.md` together. The binary name and the skill name must match. Also add a new `builds` entry in `.goreleaser.yml`.

**SKILL.md structure**: every SKILL.md requires YAML frontmatter with `name` and `description`. The `description` should be slightly pushy — tell the agent not just what the skill does but when to proactively use it. The body is instructions to the AI, not a README.

**Binaries are gitignored**: each `cmd/<name>/` has its own `.gitignore` that ignores the compiled binary. Don't commit binaries.

**No shared packages**: tools are intentionally self-contained in `package main`. Don't introduce `internal/` or shared libraries unless there is a strong reason.
