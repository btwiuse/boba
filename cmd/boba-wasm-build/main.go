// boba-wasm-build compiles a Go package to WebAssembly, injecting
// js/wasm stubs into BubbleTea v2 and its dependencies at build time.
//
// BubbleTea v2 lacks js/wasm build tags for signal handling and TTY
// initialization (see charmbracelet/bubbletea#1410). The atotto/clipboard
// library also lacks js stubs. This tool works around both by copying
// each module to a temp directory, adding stub files, and building
// through temporary go.mod replace directives.
//
// Usage:
//
//	go run github.com/btwiuse/boba/cmd/boba-wasm-build \
//	    -o web/app.wasm ./cmd/myapp/
//
// All flags and arguments are forwarded to `go build` unchanged.
// GOOS=js and GOARCH=wasm are set automatically.
package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed all:_stubs
var stubsFS embed.FS

// modulePatch describes a module that needs stub files injected at build time.
// stubDir is the subdirectory inside _stubs/ containing the files.
type modulePatch struct {
	modulePath string
	stubDir    string
}

var modulePatches = []modulePatch{
	{modulePath: "charm.land/bubbletea/v2", stubDir: "bubbletea"},
	{modulePath: "github.com/atotto/clipboard", stubDir: "clipboard"},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: boba-wasm-build [go build flags] <packages>")
		fmt.Fprintln(os.Stderr, "  e.g. boba-wasm-build -o web/app.wasm ./cmd/myapp/")
		os.Exit(2)
	}
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "boba-wasm-build: %v\n", err)
		os.Exit(1)
	}
}

func run(buildArgs []string) error {
	// Ensure dependency modules are downloaded so `go list -m` can locate
	// them. Fresh CI environments typically need this even if setup-go ran.
	download := exec.Command("go", "mod", "download")
	download.Stderr = os.Stderr
	if err := download.Run(); err != nil {
		return fmt.Errorf("go mod download: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "boba-wasm-build-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Locate the invoking project's go.mod
	origModPath, err := goEnv("GOMOD")
	if err != nil {
		return fmt.Errorf("locate go.mod: %w", err)
	}
	if origModPath == "" || origModPath == os.DevNull {
		return errors.New("no go.mod found — run this from a Go module")
	}

	// For each module that needs patching: locate in module cache, copy
	// to temp, inject stub files, and record replace directives.
	tmpMod := filepath.Join(tmpDir, "go.mod")
	if err := copyFile(origModPath, tmpMod); err != nil {
		return fmt.Errorf("copy go.mod: %w", err)
	}
	origSum := filepath.Join(filepath.Dir(origModPath), "go.sum")
	if _, err := os.Stat(origSum); err == nil {
		if err := copyFile(origSum, filepath.Join(tmpDir, "go.sum")); err != nil {
			return fmt.Errorf("copy go.sum: %w", err)
		}
	}

	for _, p := range modulePatches {
		// Locate module in cache
		srcDir, err := findModuleDir(p.modulePath)
		if err != nil {
			return fmt.Errorf("locate %s: %w", p.modulePath, err)
		}

		// Copy to temp (module cache is read-only)
		dstDir := filepath.Join(tmpDir, p.stubDir)
		if err := copyTree(srcDir, dstDir); err != nil {
			return fmt.Errorf("copy %s: %w", p.modulePath, err)
		}

		// Write embedded stubs into the copy
		if err := writeStubs(stubsFS, "_stubs/"+p.stubDir, dstDir); err != nil {
			return fmt.Errorf("write stubs for %s: %w", p.modulePath, err)
		}

		// Bump the go version in the patched module's go.mod so modern
		// language features (e.g. the `any` type alias) are accepted.
		dstMod := filepath.Join(dstDir, "go.mod")
		bumpCmd := exec.Command("go", "mod", "edit", "-modfile="+dstMod, "-go=1.18")
		bumpCmd.Stderr = os.Stderr
		if err := bumpCmd.Run(); err != nil {
			return fmt.Errorf("bump go version for %s: %w", p.modulePath, err)
		}

		// Add replace directive to the temp go.mod
		editCmd := exec.Command("go", "mod", "edit",
			"-modfile="+tmpMod,
			"-replace="+p.modulePath+"="+dstDir)
		editCmd.Stderr = os.Stderr
		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("add replace for %s: %w", p.modulePath, err)
		}
	}

	// Build with the temp modfile
	args := append([]string{"build", "-modfile=" + tmpMod}, buildArgs...)
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// writeStubs writes all files from a subdirectory of stubsFS into dstDir.
func writeStubs(stubsFS embed.FS, stubDir, dstDir string) error {
	return fs.WalkDir(stubsFS, stubDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		data, err := stubsFS.ReadFile(path)
		if err != nil {
			return err
		}
		name := filepath.Base(path)
		return os.WriteFile(filepath.Join(dstDir, name), data, 0o644)
	})
}

func findModuleDir(path string) (string, error) {
	var info struct {
		Dir string
	}

	// Fast path: module is already in the current module graph.
	if out, err := exec.Command("go", "list", "-m", "-json", path).Output(); err == nil {
		if json.Unmarshal(out, &info) == nil && info.Dir != "" {
			return info.Dir, nil
		}
	}

	// Slow path: module is not in go.mod yet (e.g. the workspace doesn't
	// depend on it directly). Fetch the latest version into the module
	// cache without modifying go.mod.
	out, err := exec.Command("go", "mod", "download", "-json", path+"@latest").Output()
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", path, err)
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return "", fmt.Errorf("parse download output for %s: %w", path, err)
	}
	if info.Dir == "" {
		return "", fmt.Errorf("module %s not in module cache", path)
	}
	return info.Dir, nil
}

func goEnv(name string) (string, error) {
	out, err := exec.Command("go", "env", name).Output()
	if err != nil {
		return "", err
	}
	s := string(out)
	if n := len(s); n > 0 && s[n-1] == '\n' {
		s = s[:n-1]
	}
	return s, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// copyTree copies src to dst, ensuring directories are writable so we can
// add new files (the module cache is copied with read-only perms).
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, relPath)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := copyFile(path, target); err != nil {
			return err
		}
		return os.Chmod(target, 0o644)
	})
}
