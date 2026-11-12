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

- 一覧のデフォルト表示は `<workspace> <ウィンドウ名>` です。workspace は2桁右揃えでアラインされます。`--format full` で `<workspace>  <ウィンドウ名> - <アプリケーション名>` になります（`list` / `switch` 共通オプション）。
- `wrep switch` ではアプリアイコンを表示します（fuzzel / rofi は rofi 拡張プロトコルの `\0icon\x1f<名前>` に対応）。デフォルトではウィンドウの class に対応する `.desktop` ファイル（`$XDG_DATA_HOME/applications` と `$XDG_DATA_DIRS/applications` を検索し、`StartupWMClass` または desktop id（ファイル名）が class と一致するもの。完全一致を優先し、なければ大文字小文字を区別せず一致するもの）の `Icon=` を使います。`Icon=` がテーマ名の場合は XDG アイコンテーマ検索（`default` テーマの `Inherits` チェーン + `hicolor` の `icons/<theme>/*/apps/<名前>.{png,svg,xpm}`、最後に `pixmaps/<名前>.<拡張子>`）で実ファイルを探し、見つかった絶対パスを fuzzel に渡します（fuzzel はパス指定のアイコンを直接読み込みます。fuzzel のアプリケーションモードと同じソース・同じ検索先です）。`.desktop` が見つからない場合は class 名そのものや、末尾ドット以降を小文字化した名前（例: `org.mozilla.Thunderbird` → `thunderbird`）でも同様に検索し、どれも見つからなければ class 由来の名前をそのまま渡して fuzzel のアイコンテーマ検索に任せます。テーマに無い名前はアイコンなしで表示されます。`--icon=false` で無効化。wofi / dmenu は非対応のため付きません。
  - `--icon-map <ファイル>` を指定すると、そのファイル（1行1エントリの `class=icon`、`#` はコメント）と組み込みマップで名前を変換します。解決順は「ユーザーマップ → 組み込みマップ（完全一致 → `*` 付きは前方一致）」です。組み込みマップの例: `footclient` → `utilities-terminal`、`org.pulseaudio.pavucontrol` → `multimedia-volume-control`、`steam_app_*` → `steam`。例: `footclient=myterm`
  - fuzzel のテーマ自体は `--icon-theme` オプション（デフォルト `default` → `Inherits=Adwaita` → … → `hicolor`）で切り替えられます。wrep から渡す場合は `internal/launcher/launcher.go` の fuzzel の argv に追記してください。
- 内部的には各行の末尾にタブ区切りで address が付き、fuzzel では `--with-nth=1 --accept-nth=2` により非表示・選択時のみ返却されます。
- デフォルトでは空タイトル・特殊ワークスペース（`monitor == -1`）のウィンドウを除外します。`--all` で含められます。
- `--launcher` には `fuzzel` 以外に `wofi` / `rofi` / `dmenu` も指定できます（デフォルト: `fuzzel`）。
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
internal/window/         # Window 型・フィルタ・行フォーマット・アイコン解決
internal/desktop/        # .desktop ファイル検索（class → Icon 取得）
internal/hyprland/       # hyprctl の呼び出し（clients / focuswindow）
internal/launcher/       # fuzzel / wofi / rofi / dmenu への一覧受け渡し
internal/cli/            # list / switch / focus コマンド
```
