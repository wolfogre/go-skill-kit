package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// analyzer pairs a modernize flag name with the minimum Go 1.x minor version
// that introduced the API or feature it suggests.
type analyzer struct {
	flag       string
	minVersion int
}

var analyzers = []analyzer{
	{"stringsbuilder", 10},  // strings.Builder
	{"plusbuild", 17},       // //go:build directive
	{"unsafefuncs", 17},     // unsafe.Add
	{"any", 18},             // any alias for interface{}
	{"stringscut", 18},      // strings.Cut
	{"atomic", 19},          // sync/atomic typed types (atomic.Bool etc.)
	{"fmtappendf", 19},      // fmt.Appendf
	{"stringscutprefix", 20}, // strings.CutPrefix / CutSuffix
	{"minmax", 21},          // min / max builtins
	{"slicescontains", 21},  // slices.Contains
	{"slicessort", 21},      // slices.Sort
	{"forvar", 22},          // redundant loop-variable re-declaration
	{"rangeint", 22},        // for i := range n
	{"reflecttypefor", 22},  // reflect.TypeFor[T]()
	{"mapsloop", 23},        // maps.Copy / Insert / Clone / Collect
	{"stditerators", 23},    // range over .All() iterators
	{"omitzero", 24},        // json omitzero tag
	{"stringsseq", 24},      // strings.SplitSeq / FieldsSeq
	{"testingcontext", 24},  // t.Context()
	{"waitgroup", 25},       // sync.WaitGroup.Go
	{"newexpr", 26},         // new(expr) builtin
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: gomodernize [modernize-flags] <packages>\n")
		os.Exit(1)
	}

	goMinor, err := readGoMinorVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gomodernize: %v\n", err)
		os.Exit(1)
	}

	var flags []string
	for _, a := range analyzers {
		if goMinor >= a.minVersion {
			flags = append(flags, "-"+a.flag)
		}
	}

	args := append(flags, os.Args[1:]...)

	var cmd *exec.Cmd
	if path, err := exec.LookPath("modernize"); err == nil {
		cmd = exec.Command(path, args...)
	} else {
		gorunArgs := append(
			[]string{"run", "golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest"},
			args...,
		)
		cmd = exec.Command("go", gorunArgs...)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "gomodernize: %v\n", err)
		os.Exit(1)
	}
}

func readGoMinorVersion() (int, error) {
	dir, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return parseGoMinorVersion(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return 0, fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func parseGoMinorVersion(gomod string) (int, error) {
	data, err := os.ReadFile(gomod)
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "go ")
		if !ok {
			continue
		}
		// version may be "1.21", "1.21.0", or "1.22rc1"
		parts := strings.SplitN(strings.TrimSpace(rest), ".", 3)
		if len(parts) >= 2 {
			minor := parts[1]
			// strip any non-digit suffix (e.g. "rc1")
			for i, c := range minor {
				if c < '0' || c > '9' {
					minor = minor[:i]
					break
				}
			}
			v, err := strconv.Atoi(minor)
			if err != nil {
				return 0, fmt.Errorf("invalid go version %q in go.mod", rest)
			}
			return v, nil
		}
	}
	return 0, fmt.Errorf("go directive not found in go.mod")
}
