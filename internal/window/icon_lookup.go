package window

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"wrep/internal/desktop"
)

// iconExts are the file extensions tried when resolving an icon name,
// in preference order (matches what fuzzel accepts).
var iconExts = []string{".png", ".svg", ".xpm"}

// themeChain returns theme followed by the themes it inherits from,
// read from each data dir's icons/<theme>/index.theme ([Icon Theme]
// Inherits=...). Cycles are cut. hicolor is always appended last as the
// fallback theme mandated by the XDG icon theme specification.
func themeChain(theme string) []string {
	seen := map[string]bool{}
	var out []string
	var walk func(string)
	walk = func(t string) {
		if t == "" || seen[t] {
			return
		}
		seen[t] = true
		out = append(out, t)
		for _, dir := range desktop.DataDirs() {
			for _, inherit := range parseInherits(filepath.Join(dir, "icons", t, "index.theme")) {
				walk(inherit)
			}
		}
	}
	walk(theme)
	if !seen["hicolor"] {
		out = append(out, "hicolor")
	}
	return out
}

// parseInherits reads the Inherits= line of an index.theme file.
func parseInherits(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	inTheme := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "[") {
			inTheme = line == "[Icon Theme]"
			continue
		}
		if !inTheme {
			continue
		}
		if key, value, ok := strings.Cut(line, "="); ok && strings.TrimSpace(key) == "Inherits" {
			for _, v := range strings.Split(value, ",") {
				if v = strings.TrimSpace(v); v != "" {
					out = append(out, v)
				}
			}
		}
	}
	return out
}

// LookupIconFile resolves an icon theme name to an absolute file path by
// searching the same locations fuzzel does: every XDG data dir's
// icons/<theme>/*/apps/<name>.<ext> for each theme in the "default" inherit
// chain, then pixmaps/<name>.<ext>. The first hit wins. ok is false when no
// file exists.
func LookupIconFile(name string) (string, bool) {
	if name == "" || strings.ContainsRune(name, '/') {
		return "", false
	}
	dirs := desktop.DataDirs()
	for _, theme := range themeChain("default") {
		for _, dir := range dirs {
			base := filepath.Join(dir, "icons", theme)
			for _, ext := range iconExts {
				matches, err := filepath.Glob(filepath.Join(base, "*", "apps", name+ext))
				if err == nil && len(matches) > 0 {
					return matches[0], true
				}
			}
		}
	}
	for _, dir := range dirs {
		for _, ext := range iconExts {
			p := filepath.Join(dir, "pixmaps", name+ext)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, true
			}
		}
	}
	return "", false
}

// ResolveIconAuto picks the icon string handed to the launcher for a window
// class when no --icon-map is configured. Unlike FallbackIconName (which
// only lowercases the class tail and hopes the launcher's theme lookup
// hits), it first consults the application's .desktop file — the same
// source fuzzel's application mode uses — and then resolves the name to an
// actual icon file path via LookupIconFile, passing an absolute path to the
// launcher (fuzzel loads path-valued icons directly). Only when nothing is
// found does it fall back to the class-derived name, letting the launcher's
// own theme lookup try.
func ResolveIconAuto(class string) string {
	if ent, ok := desktop.FindByClass(class); ok && ent.Icon != "" {
		if filepath.IsAbs(ent.Icon) {
			if st, err := os.Stat(ent.Icon); err == nil && !st.IsDir() {
				return ent.Icon
			}
			return ent.Icon // let the launcher report/ignore it
		}
		if path, ok := LookupIconFile(ent.Icon); ok {
			return path
		}
		return ent.Icon
	}
	if path, ok := LookupIconFile(class); ok {
		return path
	}
	if fb := FallbackIconName(class); fb != class {
		if path, ok := LookupIconFile(fb); ok {
			return path
		}
	}
	return FallbackIconName(class)
}
