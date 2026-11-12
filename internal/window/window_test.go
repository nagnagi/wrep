package window

import (
	"encoding/json"
	"testing"
)

func sampleClient() map[string]any {
	return map[string]any{
		"address":        "0xaaa",
		"mapped":         true,
		"hidden":         false,
		"workspace":      map[string]any{"id": 3, "name": "3"},
		"monitor":        0,
		"class":          "firefox",
		"title":          "Example - Mozilla Firefox",
		"pid":            1234,
		"xwayland":       false,
		"floating":       false,
		"pinned":         false,
		"fullscreen":     0,
		"focusHistoryID": 7,
		"stableId":       "18000029", // unknown fields must not break parsing
	}
}

func mustWindow(t *testing.T, data map[string]any) Window {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var w Window
	if err := json.Unmarshal(raw, &w); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestUnmarshalClient(t *testing.T) {
	w := mustWindow(t, sampleClient())
	if w.Address != "0xaaa" {
		t.Errorf("Address = %q", w.Address)
	}
	if w.Class != "firefox" || w.Title != "Example - Mozilla Firefox" {
		t.Errorf("Class/Title = %q/%q", w.Class, w.Title)
	}
	if w.Workspace.ID != 3 || w.Workspace.Name != "3" {
		t.Errorf("Workspace = %+v", w.Workspace)
	}
	if w.Monitor != 0 || w.PID != 1234 || w.FocusHistoryID != 7 {
		t.Errorf("Monitor/PID/FocusHistoryID = %d/%d/%d", w.Monitor, w.PID, w.FocusHistoryID)
	}
	if w.Fullscreen != 0 {
		t.Errorf("Fullscreen = %d", w.Fullscreen)
	}
}

func TestIsListable(t *testing.T) {
	base := mustWindow(t, sampleClient())
	if !base.IsListable() {
		t.Error("normal window should be listable")
	}
	cases := []struct {
		name   string
		mutate func(map[string]any)
		want   bool
	}{
		{"empty title", func(m map[string]any) { m["title"] = "" }, false},
		{"special workspace", func(m map[string]any) { m["monitor"] = -1 }, false},
		{"hidden", func(m map[string]any) { m["hidden"] = true }, false},
		{"unmapped", func(m map[string]any) { m["mapped"] = false }, false},
	}
	for _, c := range cases {
		data := sampleClient()
		c.mutate(data)
		if got := mustWindow(t, data).IsListable(); got != c.want {
			t.Errorf("%s: IsListable() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestFilterListable(t *testing.T) {
	mk := func(addr string, mutate func(map[string]any)) Window {
		data := sampleClient()
		data["address"] = addr
		if mutate != nil {
			mutate(data)
		}
		return mustWindow(t, data)
	}
	windows := []Window{
		mk("0xaaa", nil),
		mk("0xempty", func(m map[string]any) { m["title"] = "" }),
		mk("0xspecial", func(m map[string]any) { m["monitor"] = -1 }),
		mk("0xunmapped", func(m map[string]any) { m["mapped"] = false }),
	}
	got := FilterListable(windows, false)
	if len(got) != 1 || got[0].Address != "0xaaa" {
		t.Errorf("default filter = %v", addresses(got))
	}
	gotAll := FilterListable(windows, true)
	if len(gotAll) != 3 {
		t.Errorf("--all filter = %v, want 3 windows", addresses(gotAll))
	}
	for _, w := range gotAll {
		if w.Address == "0xunmapped" {
			t.Error("unmapped window must be excluded even with --all")
		}
	}
}

func addresses(ws []Window) []string {
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i] = w.Address
	}
	return out
}

func TestFormatLine(t *testing.T) {
	w := mustWindow(t, sampleClient())
	want := "3\tfirefox\tExample - Mozilla Firefox\t0xaaa"
	if got := FormatLine(w); got != want {
		t.Errorf("FormatLine = %q, want %q", got, want)
	}
}

func TestAddressFromLine(t *testing.T) {
	cases := []struct{ in, want string }{
		{"3\tfirefox\tSome title with\ttabs\t0xabc\n", "0xabc"},
		{"0xabc", "0xabc"},
		{"1\tfoot\t~ - foot\t0xaaa", "0xaaa"},
	}
	for _, c := range cases {
		if got := AddressFromLine(c.in); got != c.want {
			t.Errorf("AddressFromLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
