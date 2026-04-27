package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// block represents a coverage block parsed from coverage.out.
// Format: <file>:<startLine>.<startCol>,<endLine>.<endCol> <stmts> <count>
type block struct {
	file      string
	startLine int
	endLine   int
	count     int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gocoverage <coverage.out>")
		os.Exit(1)
	}

	blocks, err := parseCoverage(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing coverage: %v\n", err)
		os.Exit(1)
	}

	modRoot, modName, goModCache, err := moduleInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading module info: %v\n", err)
		os.Exit(1)
	}

	// Group uncovered blocks by file, preserving order.
	fileOrder := []string{}
	fileBlocks := map[string][]block{}
	for _, b := range blocks {
		if b.count > 0 {
			continue
		}
		if _, ok := fileBlocks[b.file]; !ok {
			fileOrder = append(fileOrder, b.file)
		}
		fileBlocks[b.file] = append(fileBlocks[b.file], b)
	}

	if len(fileOrder) == 0 {
		fmt.Println("all statements covered!")
		return
	}

	for _, file := range fileOrder {
		srcPath := resolveSource(file, modName, modRoot, goModCache)
		lines, err := readLines(srcPath)
		if err != nil {
			fmt.Printf("## %s\n(source not found: %v)\n\n", file, err)
			continue
		}

		fmt.Printf("## %s\n", file)
		printed := map[int]bool{}
		for _, b := range fileBlocks[file] {
			for ln := b.startLine; ln <= b.endLine && ln <= len(lines); ln++ {
				if printed[ln] {
					continue
				}
				printed[ln] = true
				fmt.Printf("%4d: %s\n", ln, lines[ln-1])
			}
		}
		fmt.Println()
	}
}

func parseCoverage(path string) ([]block, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var blocks []block
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		b, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid line %q: %w", line, err)
		}
		blocks = append(blocks, b)
	}
	return blocks, scanner.Err()
}

// parseLine parses a line like:
// github.com/foo/bar/pkg.go:10.5,20.10 3 0
func parseLine(line string) (block, error) {
	// Split off the count and stmts at the end.
	lastSpace := strings.LastIndex(line, " ")
	if lastSpace < 0 {
		return block{}, fmt.Errorf("missing count")
	}
	count, err := strconv.Atoi(line[lastSpace+1:])
	if err != nil {
		return block{}, err
	}
	rest := line[:lastSpace]

	secondSpace := strings.LastIndex(rest, " ")
	if secondSpace < 0 {
		return block{}, fmt.Errorf("missing stmts")
	}
	rest = rest[:secondSpace]

	// rest is now "<file>:<startLine>.<startCol>,<endLine>.<endCol>"
	colon := strings.LastIndex(rest, ":")
	if colon < 0 {
		return block{}, fmt.Errorf("missing colon")
	}
	file := rest[:colon]
	coords := rest[colon+1:]

	comma := strings.Index(coords, ",")
	if comma < 0 {
		return block{}, fmt.Errorf("missing comma in coords")
	}
	startLine, err := parseLineNum(coords[:comma])
	if err != nil {
		return block{}, err
	}
	endLine, err := parseLineNum(coords[comma+1:])
	if err != nil {
		return block{}, err
	}

	return block{file: file, startLine: startLine, endLine: endLine, count: count}, nil
}

// parseLineNum parses "line.col" and returns the line number.
func parseLineNum(s string) (int, error) {
	dot := strings.Index(s, ".")
	if dot < 0 {
		return strconv.Atoi(s)
	}
	return strconv.Atoi(s[:dot])
}

// moduleInfo returns the module root directory, module name, and GOMODCACHE.
func moduleInfo() (root, name, goModCache string, err error) {
	out, err := exec.Command("go", "env", "GOMOD", "GOMODCACHE").Output()
	if err != nil {
		return "", "", "", err
	}
	lines := strings.SplitN(strings.TrimRight(string(out), "\n"), "\n", 2)
	gomod := strings.TrimSpace(lines[0])
	if len(lines) == 2 {
		goModCache = strings.TrimSpace(lines[1])
	}

	if gomod == "" || gomod == os.DevNull {
		return "", "", goModCache, nil
	}
	root = filepath.Dir(gomod)

	f, err := os.Open(gomod)
	if err != nil {
		return "", "", "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			break
		}
	}
	return root, name, goModCache, scanner.Err()
}

// resolveSource finds the source file on disk for a coverage path.
func resolveSource(coverFile, modName, modRoot, goModCache string) string {
	if modName != "" && strings.HasPrefix(coverFile, modName+"/") {
		rel := strings.TrimPrefix(coverFile, modName+"/")
		return filepath.Join(modRoot, rel)
	}
	if modName != "" && coverFile == modName {
		return filepath.Join(modRoot, ".")
	}

	// External module: path is <module>@<version>/file or <module>/file.
	// Try GOMODCACHE: strip to find module@version prefix.
	if goModCache != "" {
		// Find the longest prefix that looks like a module path (contains @).
		// Coverage tool stores: <module>/<file>, and module cache is <module>@<version>.
		// We can't know the version from coverage alone, so glob for it.
		parts := strings.Split(coverFile, "/")
		for i := len(parts) - 1; i > 0; i-- {
			modPath := strings.Join(parts[:i], "/")
			relPath := strings.Join(parts[i:], "/")
			pattern := filepath.Join(goModCache, modPath+"@*", relPath)
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				return matches[len(matches)-1]
			}
		}
	}

	// Fallback: try as a relative path.
	return coverFile
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
