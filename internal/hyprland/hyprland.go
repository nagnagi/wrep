// Package hyprland wraps hyprctl for querying and focusing windows.
package hyprland

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"wrep/internal/window"
)

// HyprlandError is returned when a hyprctl call fails.
type HyprlandError struct{ Msg string }

func (e *HyprlandError) Error() string { return e.Msg }

// RunJSON and RunDispatch are variables so tests can stub hyprctl.

// RunJSON runs `hyprctl <args> -j` and returns its stdout.
var RunJSON = func(args ...string) ([]byte, error) {
	argv := append(append([]string{}, args...), "-j")
	cmd := exec.Command("hyprctl", argv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, &HyprlandError{fmt.Sprintf("hyprctl %s failed: %s",
			strings.Join(args, " "), strings.TrimSpace(stderr.String()))}
	}
	return out, nil
}

// RunDispatch runs `hyprctl dispatch <dispatcher> [args...]`.
var RunDispatch = func(dispatcher string, args ...string) error {
	argv := append([]string{"hyprctl", "dispatch", dispatcher}, args...)
	cmd := exec.Command(argv[0], argv[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return &HyprlandError{fmt.Sprintf("hyprctl dispatch %s failed: %s",
			strings.Join(argv[2:], " "), strings.TrimSpace(stderr.String()))}
	}
	return nil
}

// Clients returns all clients known to Hyprland.
func Clients() ([]window.Window, error) {
	out, err := RunJSON("clients")
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}
	var windows []window.Window
	if err := json.Unmarshal(out, &windows); err != nil {
		return nil, &HyprlandError{fmt.Sprintf("cannot parse hyprctl clients output: %v", err)}
	}
	return windows, nil
}

// ActiveWindow returns the currently focused window, or nil if none.
func ActiveWindow() (*window.Window, error) {
	out, err := RunJSON("activewindow")
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}
	var w window.Window
	if err := json.Unmarshal(out, &w); err != nil {
		return nil, &HyprlandError{fmt.Sprintf("cannot parse hyprctl activewindow output: %v", err)}
	}
	if w.Address == "" {
		return nil, nil
	}
	return &w, nil
}

// addressRe guards against Lua injection: the address is interpolated into
// a Lua string on Hyprland 0.55+, so only 0x-prefixed hex is accepted.
var addressRe = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)

// FocusWindow focuses the window with the given address (e.g. 0x55a1b9c2fb30).
// It first tries the legacy hyprlang form, then falls back to the Lua form
// used by Hyprland 0.55+ (Lua config), where dispatch args are Lua code.
func FocusWindow(address string) error {
	if !addressRe.MatchString(address) {
		return &HyprlandError{fmt.Sprintf("invalid window address: %q", address)}
	}
	if err := RunDispatch("focuswindow", "address:"+address); err == nil {
		return nil
	}
	lua := fmt.Sprintf(`hl.dsp.focus({ window = %q })`, "address:"+address)
	return RunDispatch(lua)
}
