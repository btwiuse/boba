//go:build !js

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/btwiuse/boba/internal/cli"
)

func main() {
	if err := cli.Execute(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
