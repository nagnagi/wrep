package hyprland

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const sampleClientsJSON = `[
  {
    "address": "0xaaa",
    "mapped": true,
    "hidden": false,
    "workspace": {"id": 1, "name": "1"},
    "monitor": 0,
    "class": "footclient",
    "title": "~ - foot",
    "pid": 2338,
    "fullscreen": 0
  },
  {
    "address": "0xbbb",
    "mapped": true,
    "hidden": false,
    "workspace": {"id": 2, "name": "2"},
    "monitor": 1,
    "class": "org.mozilla.Thunderbird",
    "title": "Inbox - Mozilla Thunderbird",
    "pid": 129134,
    "fullscreen": 1
  }
]`

// stubRunJSON replaces RunJSON for the duration of a test.
func stubRunJSON(t *testing.T, fn func(args ...string) ([]byte, error)) {
	t.Helper()
	orig := RunJSON
	RunJSON = fn
	t.Cleanup(func() { RunJSON = orig })
}

func stubRunDispatch(t *testing.T, fn func(dispatcher string, args ...string) error) {
	t.Helper()
	orig := RunDispatch
	RunDispatch = fn
	t.Cleanup(func() { RunDispatch = orig })
}

func TestClients(t *testing.T) {
	var gotArgs []string
	stubRunJSON(t, func(args ...string) ([]byte, error) {
		gotArgs = args
		return []byte(sampleClientsJSON), nil
	})
	windows, err := Clients()
	if err != nil {
		t.Fatal(err)
	}
	if len(gotArgs) != 1 || gotArgs[0] != "clients" {
		t.Errorf("args = %v, want [clients]", gotArgs)
	}
	if len(windows) != 2 {
		t.Fatalf("len(windows) = %d, want 2", len(windows))
	}
	if windows[0].Address != "0xaaa" || windows[1].Title != "Inbox - Mozilla Thunderbird" {
		t.Errorf("unexpected windows: %+v", windows)
	}
	if windows[1].Fullscreen != 1 {
		t.Errorf("Fullscreen = %d, want 1", windows[1].Fullscreen)
	}
}

func TestClientsEmpty(t *testing.T) {
	stubRunJSON(t, func(args ...string) ([]byte, error) { return []byte(""), nil })
	windows, err := Clients()
	if err != nil || windows != nil {
		t.Errorf("Clients() = %v, %v; want nil, nil", windows, err)
	}
}

func TestClientsError(t *testing.T) {
	stubRunJSON(t, func(args ...string) ([]byte, error) {
		return nil, &HyprlandError{Msg: "boom"}
	})
	if _, err := Clients(); err == nil || err.Error() != "boom" {
		t.Errorf("err = %v, want boom", err)
	}
}

func TestActiveWindow(t *testing.T) {
	stubRunJSON(t, func(args ...string) ([]byte, error) {
		return []byte(`{"address":"0xaaa","title":"~ - foot"}`), nil
	})
	w, err := ActiveWindow()
	if err != nil {
		t.Fatal(err)
	}
	if w == nil || w.Address != "0xaaa" {
		t.Errorf("ActiveWindow() = %+v", w)
	}

	stubRunJSON(t, func(args ...string) ([]byte, error) { return []byte(""), nil })
	if w, err := ActiveWindow(); w != nil || err != nil {
		t.Errorf("empty: ActiveWindow() = %v, %v", w, err)
	}
}

func TestFocusWindowLegacySyntax(t *testing.T) {
	var calls [][]string
	stubRunDispatch(t, func(dispatcher string, args ...string) error {
		calls = append(calls, append([]string{dispatcher}, args...))
		return nil
	})
	if err := FocusWindow("0xaaa"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0][0] != "focuswindow" || calls[0][1] != "address:0xaaa" {
		t.Errorf("calls = %v", calls)
	}
}

func TestFocusWindowLuaFallback(t *testing.T) {
	var calls [][]string
	stubRunDispatch(t, func(dispatcher string, args ...string) error {
		calls = append(calls, append([]string{dispatcher}, args...))
		if dispatcher == "focuswindow" {
			return errors.New("lua syntax error") // legacy form refused
		}
		return nil
	})
	if err := FocusWindow("0xAAA123"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("calls = %v", calls)
	}
	want := `hl.dsp.focus({ window = "address:0xAAA123" })`
	if calls[1][0] != want {
		t.Errorf("lua call = %q, want %q", calls[1][0], want)
	}
}

func TestFocusWindowRejectsInvalidAddress(t *testing.T) {
	stubRunDispatch(t, func(dispatcher string, args ...string) error {
		t.Error("must not dispatch for invalid address")
		return nil
	})
	for _, bad := range []string{"0xZZ", "address:0x1; os.execute('x')", "", "0x1\n"} {
		if err := FocusWindow(bad); err == nil {
			t.Errorf("FocusWindow(%q) = nil, want error", bad)
		}
	}
}

func TestFocusWindowError(t *testing.T) {
	stubRunDispatch(t, func(dispatcher string, args ...string) error {
		return errors.New("no such window")
	})
	if err := FocusWindow("0xdead"); err == nil {
		t.Error("want error")
	}
}

// TestRunDispatchArgv runs the real RunDispatch against a fake hyprctl binary
// that records its argv, to catch argument-construction bugs.
func TestRunDispatchArgv(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "argv.log")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + log + "\n"
	if err := os.WriteFile(filepath.Join(dir, "hyprctl"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := RunDispatch("focuswindow", "address:0x1"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if want := "dispatch\nfocuswindow\naddress:0x1\n"; string(data) != want {
		t.Errorf("argv = %q, want %q", string(data), want)
	}
}
