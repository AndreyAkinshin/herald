// Package main provides the entry point for herald.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/AndreyAkinshin/herald/internal/cli"
	"github.com/AndreyAkinshin/herald/internal/errors"
	"github.com/AndreyAkinshin/herald/internal/term"
)

var version = "dev"

func init() {
	if version != "dev" {
		return
	}

	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = strings.TrimPrefix(info.Main.Version, "v")
	}
}

func main() {
	cfg, err := cli.ParseArgs(version, os.Args[1:])
	if err != nil {
		handleError(err)
	}

	if cfg == nil {
		fmt.Printf("%s %s\n", term.BoldCyan("herald"), version)
		return
	}

	if err := cli.Run(cfg); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	fmt.Fprintf(os.Stderr, "%s %v\n", term.BoldRed("Error:"), err)

	if appErr, ok := err.(*errors.AppError); ok {
		os.Exit(appErr.ExitCode)
	}

	os.Exit(errors.ExitRuntime)
}
