# go-skill-kit

A collection of CLI tools and [Copilot skills](https://docs.github.com/en/copilot/customizing-copilot/using-claude-as-your-copilot-llm) for Go development workflows.

Each tool under `cmd/` is a standalone binary. The matching skill under `skills/` contains a `SKILL.md` that tells an AI agent when and how to use it.

## Skills & tools

| Skill | Binary | Description |
|-------|--------|-------------|
| [gocoverage](skills/gocoverage/SKILL.md) | [cmd/gocoverage](cmd/gocoverage/main.go) | Print uncovered Go source lines from a coverage profile |
| [gomodernize](skills/gomodernize/SKILL.md) | [cmd/gomodernize](cmd/gomodernize/main.go) | Run modernize with analyzers auto-selected by go.mod Go version |

## Installation

```bash
go install github.com/wolfogre/go-skill-kit/cmd/gocoverage@latest
```

## Usage with Copilot

Point your Copilot agent at the `skills/` directory. Each `SKILL.md` has a frontmatter `description` that tells the agent when to trigger it automatically.

To download a skill locally (no git clone required):

```bash
curl -sL https://github.com/wolfogre/go-skill-kit/archive/refs/heads/main.tar.gz | tar xz --strip-components=2 --include='*/skills/gocoverage/*'
```

## Development

```bash
# build all binaries
go build ./cmd/...

# run a specific tool
go run ./cmd/gocoverage coverage.out
```
