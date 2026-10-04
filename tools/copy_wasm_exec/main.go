package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: copy_wasm_exec OUTPUT")
		os.Exit(2)
	}

	for _, rel := range []string{
		filepath.Join("lib", "wasm", "wasm_exec.js"),
		filepath.Join("misc", "wasm", "wasm_exec.js"),
	} {
		src := filepath.Join(runtime.GOROOT(), rel)
		data, err := os.ReadFile(src)
		if err == nil {
			if err := os.WriteFile(os.Args[1], data, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", os.Args[1], err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintln(os.Stderr, "wasm_exec.js was not found. Set WASM_EXEC=/path/to/wasm_exec.js and rerun make web.")
	os.Exit(1)
}
