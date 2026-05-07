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

type fileStat struct {
	covered int
	total   int
}

type blockKey struct {
	startLine, startCol, endLine, endCol int
}

type mergedBlock struct {
	numStmt  int
	maxCount int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gocoverage <coverage.out> [coverage2.out ...]")
		os.Exit(1)
	}

	// Merge profiles: a block is covered if Count>0 in any profile.
	var fileOrder []string
	fileBlockOrder := map[string][]blockKey{}
	fileBlockData := map[string]map[blockKey]*mergedBlock{}

	for _, arg := range os.Args[1:] {
		profiles, err := cover.ParseProfiles(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing coverage %s: %v\n", arg, err)
			os.Exit(1)
		}
		for _, p := range profiles {
			if _, ok := fileBlockData[p.FileName]; !ok {
				fileOrder = append(fileOrder, p.FileName)
				fileBlockData[p.FileName] = map[blockKey]*mergedBlock{}
			}
			bm := fileBlockData[p.FileName]
			for _, b := range p.Blocks {
				k := blockKey{b.StartLine, b.StartCol, b.EndLine, b.EndCol}
				if existing, ok := bm[k]; ok {
					if b.Count > existing.maxCount {
						existing.maxCount = b.Count
					}
				} else {
					fileBlockOrder[p.FileName] = append(fileBlockOrder[p.FileName], k)
					bm[k] = &mergedBlock{numStmt: b.NumStmt, maxCount: b.Count}
				}
			}
		}
	}

	// Build allBlocks and fileStats from merged data.
	var allBlocks []block
	fileStats := map[string]*fileStat{}
	for _, file := range fileOrder {
		st := &fileStat{}
		fileStats[file] = st
		for _, k := range fileBlockOrder[file] {
			mb := fileBlockData[file][k]
			st.total += mb.numStmt
			if mb.maxCount > 0 {
				st.covered += mb.numStmt
			} else {
				allBlocks = append(allBlocks, block{file, k.startLine, k.endLine})
			}
		}
	}

	modRoot, modName, goModCache, err := moduleInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading module info: %v\n", err)
		os.Exit(1)
	}

	// Group uncovered blocks by file, preserving order.
	uncoveredOrder := []string{}
	fileBlocks := map[string][]block{}
	for _, b := range allBlocks {
		if _, ok := fileBlocks[b.file]; !ok {
			uncoveredOrder = append(uncoveredOrder, b.file)
		}
		fileBlocks[b.file] = append(fileBlocks[b.file], b)
	}

	if len(uncoveredOrder) == 0 {
		fmt.Println("# all statements covered!")
		return
	}

	fmt.Println("# The following lines are not covered by tests.")
	fmt.Println()
	for _, file := range uncoveredOrder {
		srcPath := resolveSource(file, modName, modRoot, goModCache)
		st := fileStats[file]
		pct := 0.0
		if st.total > 0 {
			pct = float64(st.covered) / float64(st.total) * 100
		}
		lines, err := readLines(srcPath)
		if err != nil {
			fmt.Printf("## %s: %d/%d (%.1f%%)\n\n(source not found: %v)\n\n", file, st.covered, st.total, pct, err)
			continue
		}

		fmt.Printf("## %s: %d/%d (%.1f%%)\n\n", file, st.covered, st.total, pct)
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
