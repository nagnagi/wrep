# wrep

wrep is a single-binary CLI that lists the windows open in Hyprland (Wayland) and lets you switch between them from a dmenu-compatible launcher such as fuzzel.

The window list is obtained from `hyprctl clients -j`, and the selected window is focused via `hyprctl dispatch`. There is no real-time monitoring; the list is fetched fresh on every invocation.

<!-- Logo: none yet -->

## Installation

### Requirements

- Go (build time only)
- Hyprland with `hyprctl` available at runtime
- One of fuzzel, rofi, wofi, or dmenu (for `wrep switch`)

### Build from source

```sh
go build -o wrep .
```

This produces a statically linked single binary (~4.5 MB).

### install.sh

```sh
./install.sh                 # installs to ~/.local/bin
./install.sh -d /usr/local/bin
PREFIX=/usr ./install.sh     # installs to /usr/bin
```

### Arch Linux (PKGBUILD)

A `PKGBUILD` is provided in this repository:

```sh
makepkg -si
```

## Features

- List Hyprland windows as `<workspace> <title>`, with the workspace column right-aligned to two digits; `--format full` appends the application class (`<workspace>  <title> - <class>`).
- `wrep switch` pipes the list into a dmenu-compatible launcher (fuzzel by default; wofi / rofi / dmenu also supported) and focuses the selected window.
- App icons in the launcher: for fuzzel and rofi, the icon is resolved from the window class via the matching `.desktop` file (`Icon=` key, searched in `$XDG_DATA_HOME/applications` and `$XDG_DATA_DIRS/applications` by `StartupWMClass` or desktop file id, exact match preferred, then case-insensitive). Theme icon names are resolved to real files through the XDG icon theme lookup (`Inherits` chain of the selected theme plus `hicolor`, then `pixmaps/`); absolute paths are passed to fuzzel, which loads them directly. Disable with `--icon=false`. `--icon-map <file>` instead maps `class=icon` entries (user file first, then built-in defaults such as `footclient` -> `utilities-terminal`, `steam_app_*` -> `steam`).
- Empty-title and special-workspace (`monitor == -1`) windows are hidden by default; include them with `--all`.
- `wrep list --json` outputs the raw JSON for scripting.
- Focus falls back from the legacy `dispatch focuswindow address:...` syntax to the Hyprland 0.55+ Lua form `dispatch 'hl.dsp.focus({ window = "address:..." })'`. Addresses are validated against `^0x[0-9a-fA-F]+$` to prevent Lua injection.

## Usage

```sh
wrep list [--json] [--all] [--format short|full]   # list open windows
wrep switch [--launcher fuzzel] [--prompt "Window> "] [--all]
               [--format short|full] [--icon] [--icon-map <file>]
                                                    # pick a window in a launcher and focus it
wrep focus 0x55a1b9c2fb30                           # focus a window by address
```

Internally, each line carries the window address appended after a tab; fuzzel is invoked with `--with-nth=1 --accept-nth=2` so the address is hidden but returned on selection.

Example Hyprland keybind:

```lua
-- hyprland.lua / hyprland.conf
hl.bind("SUPER", "W", function() hl.exec_cmd("wrep switch --launcher fuzzel") end)
```

## License

MIT License. See [LICENSE](LICENSE).

## Contributing

Issues and pull requests are welcome.

```sh
go test ./...   # run tests
go vet ./...    # static analysis
```

Layout:

```
main.go                  # entry point
internal/window/         # Window type, filtering, line formatting, icon resolution
internal/desktop/        # .desktop file lookup (class -> Icon)
internal/hyprland/       # hyprctl calls (clients / focuswindow)
internal/launcher/       # handing the list to fuzzel / wofi / rofi / dmenu
internal/cli/            # list / switch / focus commands
```
