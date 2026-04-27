---
name: gocoverage
description: Show which Go source lines are not covered by tests. Use this whenever the user wants to improve test coverage, asks what code is untested, or wants you to write tests — run this first to see exactly which lines need covering before writing any tests.
---

# gocoverage

Reads one or more Go coverage files and prints the uncovered source lines with their line numbers so you can read the actual code and write targeted tests.

## Step 1: Generate a coverage file

If the user hasn't already provided a `coverage.out`, run:

```bash
go test -coverprofile=coverage.out ./...
```

Run this from the module root (where `go.mod` lives).

## Step 2: Run gocoverage

```bash
gocoverage coverage.out
```

Multiple files are accepted and merged:

```bash
gocoverage a.out b.out
```

Also must be run from the module root so the tool can resolve source paths.

## Step 3: Read the output

```
# The following lines are not covered by tests.

## github.com/foo/bar/pkg/file.go

10:	if err != nil {
11:		return err
12:	}

17:	return nil
```

- Each `##` section is one source file.
- Lines are `lineNum:source code` — no padding on the line number.
- A blank line between groups means the blocks are non-consecutive in the file.
- If the output is `all statements covered!`, there's nothing to do.

## Step 4: Write the tests

Use the uncovered lines as your guide. Read the surrounding context in the source file to understand the logic, then write tests that exercise those paths.
