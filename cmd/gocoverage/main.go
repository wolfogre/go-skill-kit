package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/cover"
)

type block struct {
	file      string
	startLine int
	endLine   int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gocoverage <coverage.out> [coverage2.out ...]")
		os.Exit(1)
	}

	var allBlocks []block
	for _, arg := range os.Args[1:] {
		profiles, err := cover.ParseProfiles(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing coverage %s: %v\n", arg, err)
			os.Exit(1)
		}
		for _, p := range profiles {
			for _, b := range p.Blocks {
				if b.Count == 0 {
					allBlocks = append(allBlocks, block{p.FileName, b.StartLine, b.EndLine})
				}
			}
		}
	}

	modRoot, modName, goModCache, err := moduleInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading module info: %v\n", err)
		os.Exit(1)
	}

	// Group uncovered blocks by file, preserving order.
	var fileOrder []string
	fileBlocks := map[string][]block{}
	for _, b := range allBlocks {
		if _, ok := fileBlocks[b.file]; !ok {
			fileOrder = append(fileOrder, b.file)
		}
		fileBlocks[b.file] = append(fileBlocks[b.file], b)
	}

	if len(fileOrder) == 0 {
		fmt.Println("# all statements covered!")
		return
	}

	fmt.Println("# The following lines are not covered by tests.")
	fmt.Println()
	for _, file := range fileOrder {
		srcPath := resolveSource(file, modName, modRoot, goModCache)
		lines, err := readLines(srcPath)
		if err != nil {
			fmt.Printf("## %s\n\n(source not found: %v)\n\n", file, err)
			continue
		}

		fmt.Printf("## %s\n\n", file)
		lastPrinted := -1
		for _, b := range fileBlocks[file] {
			for ln := b.startLine; ln <= b.endLine && ln <= len(lines); ln++ {
				if ln == lastPrinted {
					continue
				}
				if lastPrinted >= 0 && ln > lastPrinted+1 {
					fmt.Println()
				}
				lastPrinted = ln
				fmt.Printf("%d:%s\n", ln, lines[ln-1])
			}
		}
		fmt.Println()
	}
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

	data, err := os.ReadFile(gomod)
	if err != nil {
		return "", "", "", err
	}
	f, err := modfile.ParseLax(gomod, data, nil)
	if err != nil {
		return "", "", "", err
	}
	if f.Module != nil {
		name = f.Module.Mod.Path
	}
	return root, name, goModCache, nil
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

	// External module: glob GOMODCACHE for <module>@<version>/file.
	if goModCache != "" {
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

	return coverFile
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
