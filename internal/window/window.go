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

// FormatLine renders one line per window: workspace, class, title, address
// (tab separated). The address is the last field so any dmenu-compatible
// launcher can return the whole selected line and the address can be
// extracted from it.
func FormatLine(w Window) string {
	return fmt.Sprintf("%s\t%s\t%s\t%s", w.Workspace.Name, w.Class, w.Title, w.Address)
}

// AddressFromLine extracts the window address from a selected launcher line.
func AddressFromLine(line string) string {
	line = strings.TrimRight(line, "\n")
	if i := strings.LastIndex(line, "\t"); i >= 0 {
		return strings.TrimSpace(line[i+1:])
	}
	return strings.TrimSpace(line)
}
