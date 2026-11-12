package launcher

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNamesSorted(t *testing.T) {
	names := Names()
	want := []string{"dmenu", "fuzzel", "rofi", "wofi"}
	if len(names) != len(want) {
		t.Fatalf("Names() = %v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("Names() = %v, want %v", names, want)
		}
	}
}

func TestDefaultLauncher(t *testing.T) {
	if DefaultLauncher != "fuzzel" {
		t.Errorf("DefaultLauncher = %q", DefaultLauncher)
	}
}

func TestBuildArgv(t *testing.T) {
	cases := []struct {
		name string
		want []string
	}{
		{"fuzzel", []string{"fuzzel", "--dmenu", "--prompt", "W> ", "--with-nth", "1,2,3", "--width", "80"}},
		{"wofi", []string{"wofi", "--show", "dmenu", "--prompt", "W> "}},
		{"rofi", []string{"rofi", "-dmenu", "-i", "-p", "W> "}},
		{"dmenu", []string{"dmenu", "-i", "-p", "W> "}},
	}
	for _, c := range cases {
		got := Launchers[c.name].BuildArgv("W> ")
		if len(got) != len(c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: argv = %v, want %v", c.name, got, c.want)
				break
			}
		}
	}
}

func TestSelectUnknownLauncher(t *testing.T) {
	if _, err := Select("nope", []string{"x"}, "p"); err == nil {
		t.Error("want LauncherError for unknown launcher")
	}
}

func TestSelectEmptyLines(t *testing.T) {
	got, err := Select("fuzzel", nil, "p")
	if err != nil || got != "" {
		t.Errorf("Select(empty) = %q, %v", got, err)
	}
}

// writeFakeLauncher registers a fake launcher binary in PATH and returns a
// cleanup-free test launcher entry.
func writeFakeLauncher(t *testing.T, name, script string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake launcher")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	Launchers[name] = Launcher{Name: name, BuildArgv: func(prompt string) []string {
		return []string{name, "--prompt", prompt}
	}}
	t.Cleanup(func() { delete(Launchers, name) })
}

func TestSelectReturnsSelectedLine(t *testing.T) {
	// Fake launcher discards stdin and prints a fixed selected line.
	writeFakeLauncher(t, "fakeselect", "#!/bin/sh\ncat >/dev/null\nprintf '1\\tfoot\\t~ - foot\\t0xaaa\\n'\n")
	got, err := Select("fakeselect", []string{"1\tfoot\t~ - foot\t0xaaa"}, "W> ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "1\tfoot\t~ - foot\t0xaaa" {
		t.Errorf("Select = %q", got)
	}
}

func TestSelectCancelReturnsEmpty(t *testing.T) {
	writeFakeLauncher(t, "fakecancel", "#!/bin/sh\ncat >/dev/null\nexit 1\n")
	got, err := Select("fakecancel", []string{"a\tb\tc\t0x1"}, "W> ")
	if err != nil || got != "" {
		t.Errorf("Select(cancel) = %q, %v; want \"\", nil", got, err)
	}
}

func TestSelectMissingBinary(t *testing.T) {
	Launchers["missingbin"] = Launcher{Name: "missingbin", BuildArgv: func(prompt string) []string {
		return []string{"definitely-not-a-real-binary-xyz"}
	}}
	t.Cleanup(func() { delete(Launchers, "missingbin") })
	if _, err := Select("missingbin", []string{"a"}, "p"); err == nil {
		t.Error("want LauncherError for missing binary")
	}
}
