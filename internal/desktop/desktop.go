// Package desktop searches XDG application directories for .desktop files
// and extracts the fields wrep needs (Icon, StartupWMClass).
package desktop

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Entry holds the fields of a .desktop file that wrep cares about.
type Entry struct {
	// Icon is the raw Icon= value: an absolute path, or an icon theme name.
	Icon string
	// StartupWMClass is the StartupWMClass= value, if present.
	StartupWMClass string
	// Hidden reports the Hidden=true flag.
	Hidden bool
}

// DataDirs returns $XDG_DATA_HOME followed by $XDG_DATA_DIRS, with the
// defaults from the XDG base directory specification.
func DataDirs() []string {
	dirs := []string{}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if v := os.Getenv("XDG_DATA_HOME"); v != "" {
			dirs = append(dirs, v)
		} else {
			dirs = append(dirs, filepath.Join(home, ".local", "share"))
		}
	}
	if v := os.Getenv("XDG_DATA_DIRS"); v != "" {
		dirs = append(dirs, strings.Split(v, ":")...)
	} else {
		dirs = append(dirs, "/usr/local/share", "/usr/share")
	}
	return dirs
}

// FindByClass returns the entry of the first .desktop file whose
// StartupWMClass matches class, or whose desktop-file id (file name without
// ".desktop") equals class. Exact matches win over case-insensitive ones.
// Files with Hidden=true are skipped.
func FindByClass(class string) (Entry, bool) {
	if class == "" {
		return Entry{}, false
	}
	var fold Entry
	foldOK := false
	for _, dir := range DataDirs() {
		appDir := filepath.Join(dir, "applications")
		entries, err := os.ReadDir(appDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".desktop") {
				continue
			}
			id := strings.TrimSuffix(e.Name(), ".desktop")
			ent, ok := parse(filepath.Join(appDir, e.Name()))
			if !ok || ent.Hidden {
				continue
			}
			if id == class {
				return ent, true
			}
			if ent.StartupWMClass == class {
				return ent, true
			}
			if !foldOK && (strings.EqualFold(ent.StartupWMClass, class) || strings.EqualFold(id, class)) {
				fold, foldOK = ent, true
			}
		}
	}
	return fold, foldOK
}

// parse reads the [Desktop Entry] section of a .desktop file.
func parse(path string) (Entry, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Entry{}, false
	}
	defer f.Close()
	var ent Entry
	inEntry := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inEntry = line == "[Desktop Entry]"
			continue
		}
		if !inEntry {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		// Skip localized keys like Icon[ja]=...
		if strings.Contains(key, "[") {
			continue
		}
		switch key {
		case "Icon":
			ent.Icon = strings.TrimSpace(value)
		case "StartupWMClass":
			ent.StartupWMClass = strings.TrimSpace(value)
		case "Hidden":
			ent.Hidden = strings.EqualFold(strings.TrimSpace(value), "true")
		}
	}
	return ent, sc.Err() == nil
}
