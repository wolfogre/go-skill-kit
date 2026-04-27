package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

// analyzer pairs a modernize flag name with the minimum Go 1.x minor version
// that introduced the API or feature it suggests.
type analyzer struct {
	flag       string
	minVersion int
}

var analyzers = []analyzer{
	{"stringsbuilder", 10},   // strings.Builder
	{"plusbuild", 17},        // //go:build directive
	{"unsafefuncs", 17},      // unsafe.Add
	{"any", 18},              // any alias for interface{}
	{"stringscut", 18},       // strings.Cut
	{"atomic", 19},           // sync/atomic typed types (atomic.Bool etc.)
	{"fmtappendf", 19},       // fmt.Appendf
	{"stringscutprefix", 20}, // strings.CutPrefix / CutSuffix
	{"minmax", 21},           // min / max builtins
	{"slicescontains", 21},   // slices.Contains
	{"slicessort", 21},       // slices.Sort
	{"forvar", 22},           // redundant loop-variable re-declaration
	{"rangeint", 22},         // for i := range n
	{"reflecttypefor", 22},   // reflect.TypeFor[T]()
	{"mapsloop", 23},         // maps.Copy / Insert / Clone / Collect
	{"stditerators", 23},     // range over .All() iterators
	{"omitzero", 24},         // json omitzero tag
	{"stringsseq", 24},       // strings.SplitSeq / FieldsSeq
	{"testingcontext", 24},   // t.Context()
	{"waitgroup", 25},        // sync.WaitGroup.Go
	{"newexpr", 26},          // new(expr) builtin
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

func parseGoMinorVersion(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	f, err := modfile.ParseLax(path, data, nil)
	if err != nil {
		return 0, err
	}
	if f.Go == nil {
		return 0, fmt.Errorf("go directive not found in go.mod")
	}
	// version is "1.21", "1.21.0", or "1.22rc1" — parse the minor part
	parts := strings.SplitN(f.Go.Version, ".", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("unexpected go version %q", f.Go.Version)
	}
	minor := parts[1]
	end := 0
	for end < len(minor) && minor[end] >= '0' && minor[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, fmt.Errorf("unexpected go version %q", f.Go.Version)
	}
	return strconv.Atoi(minor[:end])
}
