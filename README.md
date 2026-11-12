# wrep

Hyprland (Wayland) で開いているウィンドウを、fuzzel 等の dmenu 互換ランチャで列挙・切替する単一バイナリの CLI です。

ウィンドウ一覧は `hyprctl clients -j` から取得し、選択したウィンドウは `hyprctl dispatch` の focus でフォーカスします。リアルタイム監視は行わず、起動のたびに一覧を取得します。

## ビルド

```
go build -o wrep .
```

静的リンクの単一バイナリ（約 4.5 MB）ができます。依存は hyprctl のみ（実行時）、ビルド時は Go のみです。

## 使い方

```
wrep list [--json] [--all]                  # 開いているウィンドウを一覧表示
wrep switch [--launcher fuzzel] [--prompt "Window> "] [--all]
                                            # ランチャで選択してフォーカス
wrep focus 0x55a1b9c2fb30                   # address を指定してフォーカス
```

- デフォルトでは空タイトル・特殊ワークスペース（`monitor == -1`）のウィンドウを除外します。`--all` で含められます。
- `--launcher` には `fuzzel` 以外に `wofi` / `rofi` / `dmenu` も指定できます（デフォルト: `fuzzel`）。fuzzel では `--with-nth` で address 列を非表示にします。
- フォーカスはまず旧構文（hyprlang時代）の `dispatch focuswindow address:...` を試し、失敗したら Hyprland 0.55+（Lua 設定）の `dispatch 'hl.dsp.focus({ window = "address:..." })'` にフォールバックします。address は `^0x[0-9a-fA-F]+$` のみ受け付けます（Lua へのインジェクション防止）。

## Hyprland での設定例

```
# hyprland.lua / hyprland.conf
hl.bind("SUPER", "W", function() hl.exec_cmd("wrep switch --launcher fuzzel") end)
```

## 開発

```
go test ./...     # テスト
go vet ./...      # 静的解析
```

構成:

```
main.go                  # エントリポイント
internal/window/         # Window 型・フィルタ・行フォーマット
internal/hyprland/       # hyprctl の呼び出し（clients / focuswindow）
internal/launcher/       # fuzzel / wofi / rofi / dmenu への一覧受け渡し
internal/cli/            # list / switch / focus コマンド
```
