---
name: goimportsreviser
description: Organize and format Go import statements. Use this whenever the user wants to fix import ordering, remove unused imports, or format imports — run goimportsreviser before finishing any code change that touches imports.
---

# goimportsreviser

Locates the project's `go.mod` (from the given path argument, or the current directory when none is given) to determine the module path and company domain, then runs `goimports-reviser` with the correct grouping:

1. **std** — Go standard library
2. **general** — third-party packages
3. **company** — packages sharing the same registered domain as the project (e.g. `example.com`)
4. **project** — packages within the current module

## Step 1: Run on the whole project

```bash
go run github.com/wolfogre/go-skill-kit/cmd/goimportsreviser@latest
```

Equivalent to running on `./...` from the module root. No arguments needed.

## Step 2: Run on specific files or packages

```bash
go run github.com/wolfogre/go-skill-kit/cmd/goimportsreviser@latest ./internal/pkg/foo/...
go run github.com/wolfogre/go-skill-kit/cmd/goimportsreviser@latest path/to/file.go
```

Any arguments are forwarded directly to `goimports-reviser`.

## Step 3: Preview without writing

```bash
go run github.com/wolfogre/go-skill-kit/cmd/goimportsreviser@latest -list-diff
```

Lists files whose imports differ without modifying them.

## Notes

- `-rm-unused` is always enabled: unused imports are removed automatically.
- `-format` is always enabled: code is also `gofmt`-formatted.
- The module is located by walking up from the first path argument, or the current directory when no argument is given. You can run it from outside the module by passing an absolute path, e.g. `goimportsreviser -list-diff /path/to/module`.
- The company domain is derived from the module path in `go.mod` using the registered (eTLD+1) domain. For example, `code.example.com/org/repo` → company prefix `example.com`.
- Requires Go in `PATH` (uses `go run github.com/incu6us/goimports-reviser/v3@latest` internally; no local installation required).
