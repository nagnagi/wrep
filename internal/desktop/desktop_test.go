package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateXDG points XDG_DATA_HOME at a temp dir and XDG_DATA_DIRS at an
// empty temp dir so tests do not see the host's real applications.
func isolateXDG(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	sys := t.TempDir()
	t.Setenv("XDG_DATA_HOME", home)
	t.Setenv("XDG_DATA_DIRS", sys)
	return home
}

func writeDesktop(t *testing.T, dir, name, content string) {
	t.Helper()
	appDir := filepath.Join(dir, "applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindByClassStartupWMClass(t *testing.T) {
	home := isolateXDG(t)
	writeDesktop(t, home, "myapp.desktop", "[Desktop Entry]\nName=MyApp\nIcon=myapp-icon\nStartupWMClass=org.example.MyApp\n")
	ent, ok := FindByClass("org.example.MyApp")
	if !ok || ent.Icon != "myapp-icon" || ent.StartupWMClass != "org.example.MyApp" {
		t.Errorf("entry = %+v, ok = %v", ent, ok)
	}
}

func TestFindByClassDesktopID(t *testing.T) {
	home := isolateXDG(t)
	writeDesktop(t, home, "firefox.desktop", "[Desktop Entry]\nName=Firefox\nIcon=firefox\n")
	ent, ok := FindByClass("firefox")
	if !ok || ent.Icon != "firefox" {
		t.Errorf("entry = %+v, ok = %v", ent, ok)
	}
}

func TestFindByClassCaseInsensitiveFallback(t *testing.T) {
	home := isolateXDG(t)
	writeDesktop(t, home, "Alacritty.desktop", "[Desktop Entry]\nIcon=Alacritty\nStartupWMClass=Alacritty\n")
	ent, ok := FindByClass("alacritty")
	if !ok || ent.Icon != "Alacritty" {
		t.Errorf("entry = %+v, ok = %v", ent, ok)
	}
}

func TestFindByClassExactWinsOverFold(t *testing.T) {
	home := isolateXDG(t)
	writeDesktop(t, home, "a.desktop", "[Desktop Entry]\nIcon=fold-icon\nStartupWMClass=Foo\n")
	writeDesktop(t, home, "b.desktop", "[Desktop Entry]\nIcon=exact-icon\nStartupWMClass=foo\n")
	ent, ok := FindByClass("foo")
	if !ok || ent.Icon != "exact-icon" {
		t.Errorf("entry = %+v, ok = %v; want exact match to win", ent, ok)
	}
}

func TestFindByClassSkipsHidden(t *testing.T) {
	home := isolateXDG(t)
	writeDesktop(t, home, "hidden.desktop", "[Desktop Entry]\nIcon=x\nHidden=true\nStartupWMClass=ghost\n")
	if _, ok := FindByClass("ghost"); ok {
		t.Error("hidden entry must be skipped")
	}
}

func TestFindByClassNotFound(t *testing.T) {
	isolateXDG(t)
	if _, ok := FindByClass("no-such-app"); ok {
		t.Error("want ok=false for unknown class")
	}
	if _, ok := FindByClass(""); ok {
		t.Error("want ok=false for empty class")
	}
}

func TestFindByClassSystemDir(t *testing.T) {
	isolateXDG(t)
	sys := t.TempDir()
	t.Setenv("XDG_DATA_DIRS", sys)
	writeDesktop(t, sys, "sysapp.desktop", "[Desktop Entry]\nIcon=sysapp\nStartupWMClass=sys.app\n")
	ent, ok := FindByClass("sys.app")
	if !ok || ent.Icon != "sysapp" {
		t.Errorf("entry = %+v, ok = %v", ent, ok)
	}
}

func TestParseIgnoresLocalizedAndOtherSections(t *testing.T) {
	home := isolateXDG(t)
	writeDesktop(t, home, "loc.desktop", "[Desktop Entry]\nName=Loc\nIcon=real-icon\nIcon[ja]=icon-ja\nStartupWMClass=loc.app\n\n[Desktop Action new]\nIcon=wrong-icon\n")
	ent, ok := FindByClass("loc.app")
	if !ok || ent.Icon != "real-icon" {
		t.Errorf("entry = %+v, ok = %v; want Icon=real-icon", ent, ok)
	}
}

func TestFindByClassAbsoluteIconPath(t *testing.T) {
	home := isolateXDG(t)
	writeDesktop(t, home, "t3.desktop", "[Desktop Entry]\nIcon=/home/user/.local/share/icons/t3.png\nStartupWMClass=com.t3tools.T3Code\n")
	ent, ok := FindByClass("com.t3tools.T3Code")
	if !ok || ent.Icon != "/home/user/.local/share/icons/t3.png" {
		t.Errorf("entry = %+v, ok = %v", ent, ok)
	}
}
