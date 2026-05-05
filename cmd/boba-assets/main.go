// boba-assets sets up a web/ directory with everything needed to host
// a BubbleTea program compiled to WebAssembly.
//
// It copies:
//   - wasm_exec.js from GOROOT (Go WASM runtime support)
//   - boba/*.js from the boba module (terminal wrapper)
//   - ghostty-web/ghostty-web.js and ghostty-vt.wasm (terminal emulator)
//   - index.html (an embedded starter template, unless one already exists)
//
// Usage:
//
//	go run github.com/btwiuse/boba/cmd/boba-assets [--force] <output-dir>
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

//go:embed template/index.html
var templateFS embed.FS

const bobaModule = "github.com/btwiuse/boba"

func main() {
	force := pflag.BoolP("force", "f", false, "overwrite an existing index.html")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [--force] <output-dir>\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Populates <output-dir> with wasm_exec.js, boba/, ghostty-web/,")
		fmt.Fprintln(os.Stderr, "and a starter index.html for hosting a BubbleTea WASM program.")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if pflag.NArg() != 1 {
		pflag.Usage()
		os.Exit(2)
	}
	if err := run(pflag.Arg(0), *force); err != nil {
		fmt.Fprintf(os.Stderr, "boba-assets: %v\n", err)
		os.Exit(1)
	}
}

func run(outDir string, force bool) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	bobaDir, err := findModuleDir(bobaModule)
	if err != nil {
		return fmt.Errorf("locate %s (run 'go mod download' first?): %w", bobaModule, err)
	}

	goroot, err := goEnv("GOROOT")
	if err != nil {
		return fmt.Errorf("locate GOROOT: %w", err)
	}

	// wasm_exec.js (Go runtime)
	wasmExec := filepath.Join(goroot, "lib", "wasm", "wasm_exec.js")
	if _, err := os.Stat(wasmExec); err != nil {
		return fmt.Errorf("wasm_exec.js not found at %s: %w", wasmExec, err)
	}
	if err := copyFile(wasmExec, filepath.Join(outDir, "wasm_exec.js")); err != nil {
		return fmt.Errorf("copy wasm_exec.js: %w", err)
	}
	fmt.Printf("  wasm_exec.js          → %s\n", filepath.Join(outDir, "wasm_exec.js"))

	// boba/*.js (terminal wrapper)
	bobaSrc := filepath.Join(bobaDir, "serve", "static", "boba")
	bobaDst := filepath.Join(outDir, "boba")
	if err := os.MkdirAll(bobaDst, 0o755); err != nil {
		return err
	}
	n, err := copyJSFiles(bobaSrc, bobaDst)
	if err != nil {
		return fmt.Errorf("copy boba assets: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("no .js files found in %s — module may be incomplete", bobaSrc)
	}
	fmt.Printf("  boba/ (%d files)      → %s\n", n, bobaDst)

	// ghostty-web (terminal emulator)
	ghSrc := filepath.Join(bobaDir, "serve", "static", "ghostty-web")
	ghDst := filepath.Join(outDir, "ghostty-web")
	if err := os.MkdirAll(ghDst, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"ghostty-web.js", "ghostty-vt.wasm"} {
		src := filepath.Join(ghSrc, name)
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("ghostty-web asset missing at %s: %w", src, err)
		}
		if err := copyFile(src, filepath.Join(ghDst, name)); err != nil {
			return fmt.Errorf("copy %s: %w", name, err)
		}
	}
	fmt.Printf("  ghostty-web/          → %s\n", ghDst)

	// index.html (only if missing or --force)
	htmlDst := filepath.Join(outDir, "index.html")
	if _, err := os.Stat(htmlDst); err == nil && !force {
		fmt.Printf("  index.html exists, skipped (use --force to overwrite)\n")
	} else {
		html, err := templateFS.ReadFile("template/index.html")
		if err != nil {
			return err
		}
		if err := os.WriteFile(htmlDst, html, 0o644); err != nil {
			return fmt.Errorf("write index.html: %w", err)
		}
		fmt.Printf("  index.html            → %s\n", htmlDst)
	}

	return nil
}

func findModuleDir(path string) (string, error) {
	out, err := exec.Command("go", "list", "-m", "-json", path).Output()
	if err != nil {
		return "", err
	}
	var info struct {
		Dir string
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return "", err
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
	return strings.TrimSpace(string(out)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	// Normalize to 0o644: assets come from the Go module cache, which can
	// hold files with 0o755 or other perms we don't want to propagate into
	// the user's web directory.
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func copyJSFiles(srcDir, dstDir string) (int, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".js") {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(dstDir, e.Name())
		if err := copyFile(src, dst); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
