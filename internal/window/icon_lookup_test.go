package window

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateXDG points XDG_DATA_HOME at a temp dir and XDG_DATA_DIRS at an
// empty temp dir so tests do not see the host's real icons/applications.
func isolateXDG(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	sys := t.TempDir()
	t.Setenv("XDG_DATA_HOME", home)
	t.Setenv("XDG_DATA_DIRS", sys)
	return home
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLookupIconFileHicolor(t *testing.T) {
	home := isolateXDG(t)
	want := filepath.Join(home, "icons", "hicolor", "48x48", "apps", "firefox.png")
	touch(t, want)
	got, ok := LookupIconFile("firefox")
	if !ok || got != want {
		t.Errorf("LookupIconFile(firefox) = %q, %v; want %q, true", got, ok, want)
	}
}

func TestLookupIconFileScalableSvg(t *testing.T) {
	home := isolateXDG(t)
	want := filepath.Join(home, "icons", "hicolor", "scalable", "apps", "emacs.svg")
	touch(t, want)
	got, ok := LookupIconFile("emacs")
	if !ok || got != want {
		t.Errorf("LookupIconFile(emacs) = %q, %v; want %q, true", got, ok, want)
	}
}

func TestLookupIconFilePixmapsCaseSensitive(t *testing.T) {
	home := isolateXDG(t)
	want := filepath.Join(home, "pixmaps", "Alacritty.svg")
	touch(t, want)
	if got, ok := LookupIconFile("Alacritty"); !ok || got != want {
		t.Errorf("LookupIconFile(Alacritty) = %q, %v; want %q, true", got, ok, want)
	}
	// Linux file systems are case-sensitive: lowercase must not match.
	if _, ok := LookupIconFile("alacritty"); ok {
		t.Error("LookupIconFile(alacritty) must not match pixmaps/Alacritty.svg")
	}
}

func TestLookupIconFileSystemDir(t *testing.T) {
	isolateXDG(t)
	sys := t.TempDir()
	t.Setenv("XDG_DATA_DIRS", sys)
	want := filepath.Join(sys, "icons", "hicolor", "24x24", "apps", "steam.png")
	touch(t, want)
	got, ok := LookupIconFile("steam")
	if !ok || got != want {
		t.Errorf("LookupIconFile(steam) = %q, %v; want %q, true", got, ok, want)
	}
}

func TestLookupIconFileNotFound(t *testing.T) {
	isolateXDG(t)
	if _, ok := LookupIconFile("no-such-icon"); ok {
		t.Error("want ok=false for unknown icon")
	}
	if _, ok := LookupIconFile(""); ok {
		t.Error("want ok=false for empty name")
	}
	if _, ok := LookupIconFile("/abs/path.png"); ok {
		t.Error("want ok=false for path-like name")
	}
}

func TestLookupIconFileThemeInherits(t *testing.T) {
	home := isolateXDG(t)
	// "default" theme inherits "custom", and the icon only exists there.
	touch(t, filepath.Join(home, "icons", "default", "index.theme"))
	os.WriteFile(filepath.Join(home, "icons", "default", "index.theme"),
		[]byte("[Icon Theme]\nInherits=custom\n"), 0o644)
	want := filepath.Join(home, "icons", "custom", "32x32", "apps", "thing.png")
	touch(t, want)
	got, ok := LookupIconFile("thing")
	if !ok || got != want {
		t.Errorf("LookupIconFile(thing) = %q, %v; want %q via Inherits chain", got, ok, want)
	}
}

func TestResolveIconAutoDesktopIconName(t *testing.T) {
	home := isolateXDG(t)
	appDir := filepath.Join(home, "applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// .desktop uses a themed icon name different from the class tail.
	os.WriteFile(filepath.Join(appDir, "org.mozilla.Thunderbird.desktop"),
		[]byte("[Desktop Entry]\nIcon=org.mozilla.Thunderbird\nStartupWMClass=thunderbird\n"), 0o644)
	want := filepath.Join(home, "icons", "hicolor", "22x22", "apps", "org.mozilla.Thunderbird.png")
	touch(t, want)
	got := ResolveIconAuto("thunderbird")
	if got != want {
		t.Errorf("ResolveIconAuto(thunderbird) = %q; want %q", got, want)
	}
}

func TestResolveIconAutoDesktopAbsoluteIcon(t *testing.T) {
	home := isolateXDG(t)
	appDir := filepath.Join(home, "applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	iconPath := filepath.Join(home, "icons", "com.t3tools.T3Code.desktop.png")
	touch(t, iconPath)
	os.WriteFile(filepath.Join(appDir, "com.t3tools.T3Code.desktop"),
		[]byte("[Desktop Entry]\nIcon="+iconPath+"\nStartupWMClass=com.t3tools.T3Code\n"), 0o644)
	if got := ResolveIconAuto("com.t3tools.T3Code"); got != iconPath {
		t.Errorf("ResolveIconAuto = %q; want %q", got, iconPath)
	}
}

func TestResolveIconAutoClassItself(t *testing.T) {
	home := isolateXDG(t)
	// No .desktop: class "Alacritty" matches pixmaps/Alacritty.svg as-is.
	want := filepath.Join(home, "pixmaps", "Alacritty.svg")
	touch(t, want)
	if got := ResolveIconAuto("Alacritty"); got != want {
		t.Errorf("ResolveIconAuto(Alacritty) = %q; want %q", got, want)
	}
}

func TestResolveIconAutoFallbackName(t *testing.T) {
	isolateXDG(t)
	// Nothing found anywhere: keep the historical class-derived name so the
	// launcher's own theme lookup can still try.
	if got := ResolveIconAuto("org.mozilla.Thunderbird"); got != "thunderbird" {
		t.Errorf("ResolveIconAuto fallback = %q; want %q", got, "thunderbird")
	}
	if got := ResolveIconAuto("firefox"); got != "firefox" {
		t.Errorf("ResolveIconAuto(firefox) = %q; want %q", got, "firefox")
	}
}
