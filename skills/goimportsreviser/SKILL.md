---
name: goimportsreviser
description: Organize and format Go import statements. Use this whenever the user wants to fix import ordering, remove unused imports, or format imports — run goimportsreviser before finishing any code change that touches imports.
---

# goimportsreviser

Reads the project's `go.mod` to determine the module path and company domain, then runs `goimports-reviser` with the correct grouping:

1. **std** — Go standard library
2. **general** — third-party packages
3. **company** — packages sharing the same registered domain as the project (e.g. `example.com`)
4. **project** — packages within the current module

## Step 1: Run on the whole project

```bash
goimportsreviser
```

Equivalent to running on `./...` from the module root. No arguments needed.

## Step 2: Run on specific files or packages

```bash
goimportsreviser ./internal/pkg/foo/...
goimportsreviser path/to/file.go
```

Any arguments are forwarded directly to `goimports-reviser`.

## Step 3: Preview without writing

```bash
goimportsreviser -list-diff
```

Lists files whose imports differ without modifying them.

## Notes

- `-rm-unused` is always enabled: unused imports are removed automatically.
- `-format` is always enabled: code is also `gofmt`-formatted.
- The company domain is derived from the module path in `go.mod` using the registered (eTLD+1) domain. For example, `code.example.com/org/repo` → company prefix `example.com`.
- Requires Go in `PATH` (uses `go run github.com/incu6us/goimports-reviser/v3@latest` internally).
