// Package window models Hyprland clients and the switcher line format.
package window

import (
	"fmt"
	"strings"
)

// Workspace is the workspace object embedded in hyprctl clients output.
type Workspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Window is one entry of `hyprctl clients -j`.
type Window struct {
	Address        string    `json:"address"`
	Mapped         bool      `json:"mapped"`
	Hidden         bool      `json:"hidden"`
	Workspace      Workspace `json:"workspace"`
	Floating       bool      `json:"floating"`
	Monitor        int       `json:"monitor"`
	Class          string    `json:"class"`
	Title          string    `json:"title"`
	PID            int       `json:"pid"`
	Xwayland       bool      `json:"xwayland"`
	Pinned         bool      `json:"pinned"`
	Fullscreen     int       `json:"fullscreen"`
	FocusHistoryID int       `json:"focusHistoryID"`
}

// IsListable reports whether the window should appear in a switcher list
// by default: mapped, not hidden, with a title, not on a special workspace.
func (w Window) IsListable() bool {
	return w.Mapped && !w.Hidden && w.Title != "" && w.Monitor != -1
}

// FilterListable returns the windows to show. With includeAll, windows on
// special workspaces or with empty titles are kept (unmapped ones are not).
func FilterListable(windows []Window, includeAll bool) []Window {
	out := make([]Window, 0, len(windows))
	for _, w := range windows {
		if includeAll {
			if w.Mapped {
				out = append(out, w)
			}
		} else if w.IsListable() {
			out = append(out, w)
		}
	}
	return out
}

// Format selects which fields appear in a launcher line.
type Format string

const (
	FormatShort Format = "short" // "<ws> <title>"
	FormatFull  Format = "full"  // "<ws>  <title> - <class>"
)

// FormatLine renders one line per window for launchers:
//
//	<display>\t<address>[\0icon\x1f<icon-name>]
//
// The workspace name is right-aligned in 2 columns. The address is the
// second tab-separated field so launchers that return the whole selected
// line (wofi/dmenu) still yield it via AddressFromLine; fuzzel hides it with
// --with-nth=1 and returns only it with --accept-nth=2. A non-empty icon is
// appended as rofi-style metadata "\0icon\x1f<name>", understood by fuzzel
// and rofi; pass "" for launchers without icon support.
func FormatLine(w Window, format Format, icon string) string {
	ws := fmt.Sprintf("%2s", w.Workspace.Name)
	var text string
	switch format {
	case FormatFull:
		text = fmt.Sprintf("%s  %s - %s", ws, w.Title, w.Class)
	default:
		text = fmt.Sprintf("%s %s", ws, w.Title)
	}
	line := text + "\t" + w.Address
	if icon != "" {
		line += "\x00icon\x1f" + icon
	}
	return line
}

// AddressFromLine extracts the window address from a selected launcher line.
// Rofi-style metadata ("\0icon\x1f...") is stripped first.
func AddressFromLine(line string) string {
	line = strings.TrimRight(line, "\n")
	if i := strings.IndexByte(line, 0); i >= 0 {
		line = line[:i]
	}
	if i := strings.LastIndex(line, "\t"); i >= 0 {
		return strings.TrimSpace(line[i+1:])
	}
	return strings.TrimSpace(line)
}
