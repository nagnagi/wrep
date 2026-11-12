// Package launcher feeds window lines to dmenu-compatible launchers.
package launcher

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// LauncherError is returned for unknown or missing launchers.
type LauncherError struct{ Msg string }

func (e *LauncherError) Error() string { return e.Msg }

// Launcher builds the argv used to present lines to the user.
type Launcher struct {
	Name      string
	BuildArgv func(prompt string) []string
}

// displayFields hides the trailing address field in launchers that support
// --with-nth (fuzzel).
const displayFields = "1,2,3"

// Launchers maps launcher names to their definitions.
var Launchers = map[string]Launcher{
	"fuzzel": {Name: "fuzzel", BuildArgv: func(prompt string) []string {
		return []string{"fuzzel", "--dmenu", "--prompt", prompt,
			"--with-nth", displayFields, "--width", "80"}
	}},
	"wofi": {Name: "wofi", BuildArgv: func(prompt string) []string {
		return []string{"wofi", "--show", "dmenu", "--prompt", prompt}
	}},
	"rofi": {Name: "rofi", BuildArgv: func(prompt string) []string {
		return []string{"rofi", "-dmenu", "-i", "-p", prompt}
	}},
	"dmenu": {Name: "dmenu", BuildArgv: func(prompt string) []string {
		return []string{"dmenu", "-i", "-p", prompt}
	}},
}

// DefaultLauncher is used when --launcher is not given.
const DefaultLauncher = "fuzzel"

// Names returns the sorted list of available launcher names.
func Names() []string {
	names := make([]string, 0, len(Launchers))
	for name := range Launchers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Select feeds lines to the launcher and returns the selected line.
// It returns "" when the list is empty or the user cancels (Esc).
// It is a variable so tests can stub it.
var Select = func(name string, lines []string, prompt string) (string, error) {
	l, ok := Launchers[name]
	if !ok {
		return "", &LauncherError{fmt.Sprintf("unknown launcher: %s (available: %s)",
			name, strings.Join(Names(), ", "))}
	}
	if len(lines) == 0 {
		return "", nil
	}
	argv := l.BuildArgv(prompt)
	if _, err := exec.LookPath(argv[0]); err != nil {
		return "", &LauncherError{fmt.Sprintf("launcher binary not found: %s", argv[0])}
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n") + "\n")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		// Launchers exit non-zero when the user cancels.
		return "", nil
	}
	return strings.TrimSpace(stdout.String()), nil
}
