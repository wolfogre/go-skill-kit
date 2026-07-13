---
name: gocoverage
description: Show which Go source lines are not covered by tests. Use this whenever the user wants to improve test coverage, asks what code is untested, or wants you to write tests — run this first to see exactly which lines need covering before writing any tests.
allowed-tools: Bash(go run github.com/wolfogre/go-skill-kit/cmd/gocoverage@latest *)
---

# gocoverage

Reads one or more Go coverage files and prints the uncovered source lines with their line numbers so you can read the actual code and write targeted tests.

## Step 1: Generate a coverage file

If the user hasn't already provided a `coverage.out`, run a targeted test command. Prefer narrow scope over `./...` to keep the output focused:

```bash
# a specific package
go test -coverprofile=coverage.out ./internal/pkg/foo/...

# a specific test function
go test -coverprofile=coverage.out -run TestFoo ./internal/pkg/foo/...

# a specific package with coverage across all packages it calls
go test -coverprofile=coverage.out -coverpkg=./... ./internal/pkg/foo/...
```

Run this from the module root (where `go.mod` lives).

## Step 2: Run gocoverage

```bash
go run github.com/wolfogre/go-skill-kit/cmd/gocoverage@latest coverage.out
```

Multiple files are accepted and merged:

```bash
go run github.com/wolfogre/go-skill-kit/cmd/gocoverage@latest a.out b.out
```

The tool locates the module by walking up from the coverage file's path, so it can be run from anywhere - point it at a coverage file inside the module:

```bash
go run github.com/wolfogre/go-skill-kit/cmd/gocoverage@latest /path/to/module/coverage.out
```

## Step 3: Read the output

```
# The following lines are not covered by tests.

## github.com/foo/bar/pkg/file.go: 8/10 (80.0%)

10:	if err != nil {
11:		return err
12:	}

17:	return nil
```

- Each `##` section is one source file, with `covered/total statements (percentage)`.
- Lines are `lineNum:source code` — no padding on the line number.
- A blank line between groups means the blocks are non-consecutive in the file.
- If the output is `all statements covered!`, there's nothing to do.

## Step 4: Write the tests

Use the uncovered lines as your guide. Read the surrounding context in the source file to understand the logic, then write tests that exercise those paths.
