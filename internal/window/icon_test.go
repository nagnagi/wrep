package window

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveIcon(t *testing.T) {
	cases := []struct {
		name    string
		class   string
		userMap map[string]string
		want    string
	}{
		{"user map wins", "footclient", map[string]string{"footclient": "myterm"}, "myterm"},
		{"user map prefix", "steam_app_3548580", map[string]string{"steam_app_*": "games"}, "games"},
		{"user map beats default", "footclient", map[string]string{"footclient": "myterm", "foot": "x"}, "myterm"},
		{"default exact", "footclient", nil, "utilities-terminal"},
		{"default prefix", "steam_app_3548580", nil, "steam"},
		{"default prefix other", "steam_app_999", nil, "steam"},
		{"dot fallback", "org.mozilla.Thunderbird", nil, "thunderbird"},
		{"lowercase fallback", "Steam", nil, "steam"},
		{"no dot", "firefox", nil, "firefox"},
		{"empty", "", nil, ""},
	}
	for _, c := range cases {
		if got := ResolveIcon(c.class, c.userMap); got != c.want {
			t.Errorf("%s: ResolveIcon(%q) = %q, want %q", c.name, c.class, got, c.want)
		}
	}
}

func TestLoadIconMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "icons.conf")
	content := `# comment
footclient=myterm

steam_app_*=steam
bad line without equals
`
	// first without the bad line
	if err := os.WriteFile(path, []byte("# comment\nfootclient=myterm\n\nsteam_app_*=steam\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadIconMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if m["footclient"] != "myterm" || m["steam_app_*"] != "steam" || len(m) != 2 {
		t.Errorf("map = %v", m)
	}

	// bad line must error with line number
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadIconMap(path); err == nil {
		t.Error("want error for malformed line")
	}

	// missing file must error
	if _, err := LoadIconMap(filepath.Join(dir, "nope.conf")); err == nil {
		t.Error("want error for missing file")
	}
}
