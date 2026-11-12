package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"wrep/internal/hyprland"
	"wrep/internal/launcher"
	"wrep/internal/window"
)

const clientsJSON = `[
  {"address":"0xaaa","mapped":true,"hidden":false,"workspace":{"id":1,"name":"1"},
   "monitor":0,"class":"footclient","title":"~ - foot","pid":1,"fullscreen":0},
  {"address":"0xbbb","mapped":true,"hidden":false,"workspace":{"id":2,"name":"2"},
   "monitor":-1,"class":"x","title":"","pid":2,"fullscreen":0}
]`

func stubHyprland(t *testing.T) *[]string {
	t.Helper()
	origJSON, origDispatch := hyprland.RunJSON, hyprland.RunDispatch
	var dispatches []string
	hyprland.RunJSON = func(args ...string) ([]byte, error) {
		return []byte(clientsJSON), nil
	}
	hyprland.RunDispatch = func(dispatcher string, args ...string) error {
		dispatches = append(dispatches, dispatcher+" "+strings.Join(args, " "))
		return nil
	}
	t.Cleanup(func() {
		hyprland.RunJSON, hyprland.RunDispatch = origJSON, origDispatch
	})
	return &dispatches
}

func run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	rc := Run(&stdout, &stderr, args)
	return rc, stdout.String(), stderr.String()
}

func TestList(t *testing.T) {
	stubHyprland(t)
	rc, out, _ := run(t, "list")
	if rc != 0 {
		t.Fatalf("rc = %d", rc)
	}
	want := " 1 ~ - foot\t0xaaa\n"
	if out != want {
		t.Errorf("list output = %q, want %q", out, want)
	}
}

func TestListFormatFull(t *testing.T) {
	stubHyprland(t)
	rc, out, _ := run(t, "list", "--format", "full")
	if rc != 0 {
		t.Fatalf("rc = %d", rc)
	}
	want := " 1  ~ - foot - footclient\t0xaaa\n"
	if out != want {
		t.Errorf("list --format full = %q, want %q", out, want)
	}
}

func TestListInvalidFormat(t *testing.T) {
	stubHyprland(t)
	rc, _, stderr := run(t, "list", "--format", "bogus")
	if rc != 2 || !strings.Contains(stderr, "invalid format") {
		t.Errorf("rc = %d, stderr = %q", rc, stderr)
	}
}

func TestListJSON(t *testing.T) {
	stubHyprland(t)
	rc, out, _ := run(t, "list", "--json")
	if rc != 0 {
		t.Fatalf("rc = %d", rc)
	}
	var windows []window.Window
	if err := json.Unmarshal([]byte(out), &windows); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(windows) != 1 || windows[0].Address != "0xaaa" {
		t.Errorf("windows = %+v", windows)
	}
	if windows[0].Class != "footclient" {
		t.Errorf("Class = %q", windows[0].Class)
	}
}

func TestListAll(t *testing.T) {
	stubHyprland(t)
	rc, out, _ := run(t, "list", "--all")
	if rc != 0 {
		t.Fatalf("rc = %d", rc)
	}
	if got := strings.Count(out, "\n"); got != 2 {
		t.Errorf("--all line count = %d, want 2", got)
	}
}

func TestSwitchFocusesSelected(t *testing.T) {
	dispatches := stubHyprland(t)
	origSelect := launcher.Select
	launcher.Select = func(name string, lines []string, prompt string) (string, error) {
		if len(lines) != 1 {
			t.Errorf("lines = %v", lines)
		}
		return lines[0], nil
	}
	t.Cleanup(func() { launcher.Select = origSelect })

	rc, _, _ := run(t, "switch", "--launcher", "fuzzel")
	if rc != 0 {
		t.Fatalf("rc = %d", rc)
	}
	if len(*dispatches) != 1 || (*dispatches)[0] != "focuswindow address:0xaaa" {
		t.Errorf("dispatches = %v", *dispatches)
	}
}

func TestSwitchIconMetadata(t *testing.T) {
	stubHyprland(t)
	origSelect := launcher.Select
	var gotLines []string
	launcher.Select = func(name string, lines []string, prompt string) (string, error) {
		gotLines = lines
		return "", nil
	}
	t.Cleanup(func() { launcher.Select = origSelect })

	// fuzzel supports icons: metadata must be appended by default
	run(t, "switch", "--launcher", "fuzzel")
	if len(gotLines) != 1 || !strings.Contains(gotLines[0], "\x00icon\x1ffootclient") {
		t.Errorf("fuzzel lines = %q, want icon metadata", gotLines)
	}
	// --icon=false disables it
	run(t, "switch", "--launcher", "fuffel", "--icon=false")
	if strings.Contains(gotLines[0], "\x00icon") {
		t.Errorf("--icon=false lines = %q, want no icon metadata", gotLines)
	}
	// wofi does not support icons: no metadata even by default
	run(t, "switch", "--launcher", "wofi")
	if strings.Contains(gotLines[0], "\x00icon") {
		t.Errorf("wofi lines = %q, want no icon metadata", gotLines)
	}
	// --format full includes class
	run(t, "switch", "--launcher", "fuffel", "--format", "full", "--icon=false")
	if !strings.HasPrefix(gotLines[0], " 1  ~ - foot - footclient\t") {
		t.Errorf("full format line = %q", gotLines[0])
	}
}

func TestSwitchCancel(t *testing.T) {
	dispatches := stubHyprland(t)
	origSelect := launcher.Select
	launcher.Select = func(name string, lines []string, prompt string) (string, error) {
		return "", nil
	}
	t.Cleanup(func() { launcher.Select = origSelect })

	rc, _, _ := run(t, "switch")
	if rc != 0 || len(*dispatches) != 0 {
		t.Errorf("rc = %d, dispatches = %v", rc, *dispatches)
	}
}

func TestFocus(t *testing.T) {
	dispatches := stubHyprland(t)
	rc, _, _ := run(t, "focus", "0xabc")
	if rc != 0 {
		t.Fatalf("rc = %d", rc)
	}
	if (*dispatches)[0] != "focuswindow address:0xabc" {
		t.Errorf("dispatches = %v", *dispatches)
	}
}

func TestHyprlandErrorExits1(t *testing.T) {
	orig := hyprland.RunJSON
	hyprland.RunJSON = func(args ...string) ([]byte, error) {
		return nil, &hyprland.HyprlandError{Msg: "boom"}
	}
	t.Cleanup(func() { hyprland.RunJSON = orig })

	rc, _, stderr := run(t, "list")
	if rc != 1 || !strings.Contains(stderr, "boom") {
		t.Errorf("rc = %d, stderr = %q", rc, stderr)
	}
}

func TestUnknownCommand(t *testing.T) {
	rc, _, _ := run(t, "bogus")
	if rc != 2 {
		t.Errorf("rc = %d, want 2", rc)
	}
}

func TestNoArgs(t *testing.T) {
	rc, _, _ := run(t)
	if rc != 2 {
		t.Errorf("rc = %d, want 2", rc)
	}
}
