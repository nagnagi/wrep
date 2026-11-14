package window

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DefaultIconMap maps window class names to icon theme names.
// Keys ending in "*" are prefix matches (e.g. "steam_app_*").
var DefaultIconMap = map[string]string{
	"foot":                       "utilities-terminal",
	"footclient":                 "utilities-terminal",
	"org.pulseaudio.pavucontrol": "multimedia-volume-control",
	"steam_app_*":                "steam",
	"com.anthropic.Claude":       "claude",
	"org.mozilla.Thunderbird":    "thunderbird",
	"google-chrome":              "google-chrome",
}

// ResolveIcon returns the icon theme name for a window class using the
// mapping tables. Lookup order: user map, then DefaultIconMap (each exact
// match first, then "*"-prefix), then FallbackIconName.
func ResolveIcon(class string, userMap map[string]string) string {
	if icon, ok := lookupIcon(userMap, class); ok {
		return icon
	}
	if icon, ok := lookupIcon(DefaultIconMap, class); ok {
		return icon
	}
	return FallbackIconName(class)
}

// FallbackIconName derives an icon theme name from the class alone, e.g.
// "org.mozilla.Thunderbird" -> "thunderbird", "Steam" -> "steam". This is
// what is passed to the launcher when no mapping is configured; whether the
// icon is actually drawn is up to the launcher's icon theme lookup.
func FallbackIconName(class string) string {
	c := class
	if i := strings.LastIndex(c, "."); i >= 0 && i < len(c)-1 {
		c = c[i+1:]
	}
	return strings.ToLower(c)
}

// lookupIcon finds class in m: exact key first, then keys ending in "*"
// as prefix matches.
func lookupIcon(m map[string]string, class string) (string, bool) {
	if icon, ok := m[class]; ok {
		return icon, true
	}
	for key, icon := range m {
		if strings.HasSuffix(key, "*") && strings.HasPrefix(class, strings.TrimSuffix(key, "*")) {
			return icon, true
		}
	}
	return "", false
}

// LoadIconMap reads a mapping file with one "class=icon" entry per line.
// Blank lines and lines starting with '#' are ignored.
func LoadIconMap(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := map[string]string{}
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("%s:%d: want 'class=icon'", path, lineNo)
		}
		m[key] = value
	}
	return m, sc.Err()
}
