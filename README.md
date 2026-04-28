# go-skill-kit

A collection of CLI tools and [Copilot skills](https://docs.github.com/en/copilot/customizing-copilot/using-claude-as-your-copilot-llm) for Go development workflows.

Each tool under `cmd/` is a standalone binary. The matching skill under `skills/` tells a Copilot agent when and how to use it automatically.

## Skills & tools

| Skill | Binary | Description |
|-------|--------|-------------|
| [gocoverage](skills/gocoverage/SKILL.md) | [cmd/gocoverage](cmd/gocoverage/main.go) | Print uncovered Go source lines from a coverage profile |
| [gomodernize](skills/gomodernize/SKILL.md) | [cmd/gomodernize](cmd/gomodernize/main.go) | Run modernize with analyzers auto-selected by go.mod Go version |
| [goimportsreviser](skills/goimportsreviser/SKILL.md) | [cmd/goimportsreviser](cmd/goimportsreviser/main.go) | Organize Go imports into std/general/company/project groups |

## Install all

```bash
go install github.com/wolfogre/go-skill-kit/cmd/gocoverage@latest
go install github.com/wolfogre/go-skill-kit/cmd/gomodernize@latest
go install github.com/wolfogre/go-skill-kit/cmd/goimportsreviser@latest
```

```bash
curl -sL https://github.com/wolfogre/go-skill-kit/archive/refs/heads/main.tar.gz | tar xz --strip-components=2 --include='*/skills/*'
```

## gocoverage

```bash
go install github.com/wolfogre/go-skill-kit/cmd/gocoverage@latest
```

```bash
curl -sL https://github.com/wolfogre/go-skill-kit/archive/refs/heads/main.tar.gz | tar xz --strip-components=2 --include='*/skills/gocoverage/*'
```

## gomodernize

```bash
go install github.com/wolfogre/go-skill-kit/cmd/gomodernize@latest
```

```bash
curl -sL https://github.com/wolfogre/go-skill-kit/archive/refs/heads/main.tar.gz | tar xz --strip-components=2 --include='*/skills/gomodernize/*'
```

## goimportsreviser

```bash
go install github.com/wolfogre/go-skill-kit/cmd/goimportsreviser@latest
```

```bash
curl -sL https://github.com/wolfogre/go-skill-kit/archive/refs/heads/main.tar.gz | tar xz --strip-components=2 --include='*/skills/goimportsreviser/*'
```
