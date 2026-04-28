package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/net/publicsuffix"
)

func main() {
	modRoot, modulePath, err := findModule()
	if err != nil {
		fmt.Fprintf(os.Stderr, "goimportsreviser: %v\n", err)
		os.Exit(1)
	}

	args := []string{
		"run", "github.com/incu6us/goimports-reviser/v3@latest",
		"-rm-unused",
		"-imports-order", "std,general,company,project",
		"-format",
	}

	if modulePath != "" {
		args = append(args, "-project-name", modulePath)
		if domain := companyDomain(modulePath); domain != "" {
			args = append(args, "-company-prefixes", domain)
		}
	}

	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	} else {
		args = append(args, "./...")
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = modRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "goimportsreviser: %v\n", err)
		os.Exit(1)
	}
}

// findModule walks up from cwd to find go.mod and returns the module root
// directory and module path.
func findModule() (root, modulePath string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			data, err := os.ReadFile(candidate)
			if err != nil {
				return "", "", err
			}
			f, err := modfile.ParseLax(candidate, data, nil)
			if err != nil {
				return "", "", err
			}
			if f.Module != nil {
				modulePath = f.Module.Mod.Path
			}
			return dir, modulePath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// companyDomain extracts the registered domain (eTLD+1) from a module path.
// For example:
//
//	"code.example.com/org/repo" → "example.com"
//	"github.com/user/repo"      → "github.com"
//	"myapp"                     → ""
func companyDomain(modulePath string) string {
	host, _, _ := strings.Cut(modulePath, "/")
	if !strings.Contains(host, ".") {
		return ""
	}
	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return host
	}
	return domain
}
