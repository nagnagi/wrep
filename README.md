# wrep

<p align="center">
  A single-binary CLI that lists the windows open in Hyprland (Wayland) and switches between them from a dmenu-compatible launcher such as fuzzel
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go&logoColor=white" alt="Go 1.27+">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License">
</p>

<!-- Logo: none yet -->

```console
$ wrep list
 1 Firefox - GitHub
 2 kitty
10 nvim

$ wrep list --format full
 1  Firefox - GitHub - firefox
 2  kitty - kitty
10  nvim - neovim
```

## Installation

You need Hyprland with `hyprctl` at runtime, one of fuzzel, rofi, wofi, or dmenu (for `wrep switch`), and Go (build time only).

```sh
git clone https://github.com/nagnagi/wrep.git
cd wrep
./install.sh
```

`install.sh` builds wrep and installs it to `~/.local/bin`. To install somewhere else, set `PREFIX` or pass `-d`.

```sh
PREFIX="$HOME/opt" ./install.sh   # installs to ~/opt/bin
./install.sh -d /usr/local/bin
```

To build manually:

```sh
go build -o wrep .
```

This produces a single static binary (~4.5 MB).

### Arch Linux (PKGBUILD)

A `PKGBUILD` is provided in this repository:

```sh
makepkg -si
```

## Features

- The window list comes from `hyprctl clients -j`, fetched fresh on every invocation; the selected window is focused via `hyprctl dispatch`.
- `wrep list` shows open windows as `<workspace> <title>`, with the workspace right-aligned to two digits; `--format full` appends the application class (`<workspace>  <title> - <class>`).
- `wrep switch` pipes the list into a dmenu-compatible launcher (fuzzel by default; rofi, wofi, dmenu also supported) and focuses the selected window.
- App icons in the launcher are resolved from each window class via the matching `.desktop` file (`Icon=` key) and the XDG icon theme. Disable with `--icon=false`; use `--icon-map <file>` for explicit `class=icon` mappings (built-in defaults such as `footclient` -> `utilities-terminal` apply).
- Empty-title and special-workspace windows are hidden by default; include them with `--all`.
- `wrep list --json` outputs the raw JSON for scripting.
- Focus falls back from the legacy `dispatch focuswindow address:...` syntax to the Hyprland 0.55+ Lua form; addresses are validated against `^0x[0-9a-fA-F]+$` before use.

## Usage

```sh
wrep list [--json] [--all] [--format short|full]      # list open windows
wrep switch [--launcher fuzzel] [--prompt "Window> "]  # pick a window in a launcher and focus it
               [--all] [--format short|full] [--icon=false] [--icon-map <file>]
wrep focus 0x55a1b9c2fb30                              # focus a window by address
```

| Command | Action |
|---|---|
| `wrep list` | List open windows |
| `wrep switch` | Pick a window in a launcher and focus it |
| `wrep focus <address>` | Focus a window by address |

For the options of each command, run `wrep <command> -h`.

Example Hyprland keybind:

```lua
-- hyprland.conf / hyprland.lua
hl.bind("SUPER", "W", function() hl.exec_cmd("wrep switch --launcher fuzzel") end)
```

## License

[MIT License](LICENSE)

## Contributing

Issues and pull requests are welcome. Before you open a pull request, make sure that the tests pass.

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
