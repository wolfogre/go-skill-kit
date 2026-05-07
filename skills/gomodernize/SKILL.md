---
name: gomodernize
description: Check Go code for modernization opportunities and apply fixes. Use this whenever the user wants to modernize their Go code, update code to use newer standard library APIs, or before writing code review comments about outdated patterns — run gomodernize first to see what can be automatically improved.
---

# gomodernize

Reads the project's `go.mod` to determine the Go version, enables only the analyzers whose required APIs are available in that version, then runs `modernize`.

## Step 1: Check for issues (dry run)

Run from the module root:

```bash
go run github.com/wolfogre/go-skill-kit/cmd/gomodernize@latest ./...
```

Or target a specific package:

```bash
go run github.com/wolfogre/go-skill-kit/cmd/gomodernize@latest ./internal/pkg/foo/...
```

Output lists diagnostics with file, line, and suggested change. Exit code 0 means nothing to modernize.

## Step 2: Apply fixes

```bash
go run github.com/wolfogre/go-skill-kit/cmd/gomodernize@latest -fix ./...
```

Some fixes may leave unused imports or variables — these are trivial compile errors. Fix them before committing.

If modernize reports "conflicting fixes", run it again; each pass resolves one layer of conflicts.

## Step 3: Review the diff

Check `git diff` before committing. Pay attention to:
- Loop-to-function rewrites that discard comments inside the loop body.
- `omitzero` changes, which alter JSON marshaling behavior for zero-value struct fields.
- `testingcontext` changes, which replace `context.WithCancel` + `defer cancel()` with `t.Context()` — only safe if `cancel` isn't used elsewhere.

## Notes

- `gomodernize` selects analyzers automatically based on the `go` directive in `go.mod`. There is no need to pass analyzer flags manually.
- Extra flags (e.g. `-fix`, `-diff`, `-json`) are forwarded directly to `modernize`.
- Always uses `go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest` internally; no local installation required.
