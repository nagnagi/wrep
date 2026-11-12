// Package cli implements the wrep command line interface.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"wrep/internal/hyprland"
	"wrep/internal/launcher"
	"wrep/internal/window"
)

const version = "0.1.0"

// Main runs the CLI with os.Args[1:].
func Main(args []string) int {
	return Run(os.Stdout, os.Stderr, args)
}

// Run executes the CLI, writing normal output to stdout and errors to stderr.
func Run(stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "list":
		return cmdList(stdout, stderr, args[1:])
	case "switch":
		return cmdSwitch(stdout, stderr, args[1:])
	case "focus":
		return cmdFocus(stderr, args[1:])
	case "-h", "--help", "help":
		usage(stdout)
		return 0
	case "--version":
		fmt.Fprintf(stdout, "wrep %s\n", version)
		return 0
	default:
		fmt.Fprintf(stderr, "wrep: unknown command: %s\n", args[0])
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `usage: wrep <command> [options]

List and switch Hyprland windows via a dmenu-compatible launcher.

commands:
  list     list open windows
  switch   pick a window in a launcher and focus it
  focus    focus a window by address

run "wrep <command> -h" for command options.
`)
}

func cmdList(stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	asJSON := fs.Bool("json", false, "output raw JSON")
	includeAll := fs.Bool("all", false, "include special-workspace and empty-title windows")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	windows, err := hyprland.Clients()
	if err != nil {
		return fail(stderr, err)
	}
	windows = window.FilterListable(windows, *includeAll)
	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(windows); err != nil {
			return fail(stderr, err)
		}
		return 0
	}
	for _, w := range windows {
		fmt.Fprintln(stdout, window.FormatLine(w))
	}
	return 0
}

func cmdSwitch(stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("switch", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	launcherName := fs.String("launcher", launcher.DefaultLauncher, "launcher to use: "+strings.Join(launcher.Names(), ", "))
	prompt := fs.String("prompt", "Window> ", "launcher prompt")
	includeAll := fs.Bool("all", false, "include special-workspace and empty-title windows")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	windows, err := hyprland.Clients()
	if err != nil {
		return fail(stderr, err)
	}
	lines := make([]string, 0, len(windows))
	for _, w := range window.FilterListable(windows, *includeAll) {
		lines = append(lines, window.FormatLine(w))
	}
	selected, err := launcher.Select(*launcherName, lines, *prompt)
	if err != nil {
		return fail(stderr, err)
	}
	if selected == "" {
		return 0
	}
	if err := hyprland.FocusWindow(window.AddressFromLine(selected)); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func cmdFocus(stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("focus", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: wrep focus <address>")
		return 2
	}
	if err := hyprland.FocusWindow(fs.Arg(0)); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "wrep: %v\n", err)
	return 1
}
