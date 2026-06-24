<div align="center">

# ⚡ Bolt

**A blazingly fast terminal file manager written in Go**

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

</div>

---

Bolt is a keyboard-driven terminal file manager with a rich interactive TUI, syntax-highlighted file previews, fuzzy search, file tagging, diff viewer, disk usage analysis, duplicate detection, and more — all from the command line.

## Features

- **Full TUI** — dual-pane explorer with file preview, live filter, multi-select
- **Syntax highlighting** — 200+ languages via Chroma
- **File operations** — copy, cut, paste, rename, batch rename, delete, trash, duplicate, symlink, chmod
- **Undo** — undo last rename, mkdir, touch, trash, or duplicate
- **Diff viewer** — side-by-side colorized diff in the preview pane
- **File tags** — attach searchable tags to any file, persisted in config
- **Disk usage** — live size bars in directory preview
- **Duplicate finder** — scan directory for identical files by SHA-256
- **Recent files** — tracks visited paths, jump back with one keystroke
- **Fuzzy search** — recursive file + content search with trigram index
- **Archive support** — zip, unzip, tar create/extract
- **Trash** — soft delete with restore, instead of permanent rm
- **Bookmarks** — save and jump to named directories
- **File watcher** — monitor directories for changes
- **Checksum** — SHA-256/512/MD5 from the TUI or CLI
- **External editor** — open any file in `$EDITOR` / `$VISUAL`
- **Shell completion** — bash, zsh, fish, PowerShell
- **Plugin system** — extend with external commands via config

---

## Installation

### macOS / Linux — prebuilt (no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/The-True-Hooha/Bolt/master/install.sh | bash
```

Downloads latest release binary, installs to `~/.local/bin/bolt`, adds it to your shell profile.

### macOS / Linux — build from source

```bash
curl -fsSL https://raw.githubusercontent.com/The-True-Hooha/Bolt/master/install.sh | bash -s -- --from-source
```

Requires Go 1.21+. Clones repo, compiles, and installs.

### Windows — prebuilt (no Go required)

```powershell
irm https://raw.githubusercontent.com/The-True-Hooha/Bolt/master/install.ps1 | iex
```

Downloads latest `.exe`, installs to `%APPDATA%\bolt\bin`, adds it to your **User PATH** — no admin rights needed.

### Windows — build from source

```powershell
.\install.ps1 -FromSource
```

Requires Go 1.21+.

### Manual (any platform)

```bash
git clone https://github.com/The-True-Hooha/Bolt
cd Bolt
go build -o bolt .
bolt install          # copies binary + updates PATH automatically
```

### Custom install directory

```bash
bolt install /usr/local/bin        # Linux / macOS
bolt install C:\Tools\bolt         # Windows
```

### Uninstall

```bash
bolt uninstall        # removes binary and PATH entry
```

---

## Quick Start

```bash
bolt          # open interactive TUI in current directory
bolt ui .     # same, explicit
bolt ls       # list files (like ls/dir)
bolt preview README.md   # syntax-highlighted preview
```

---

## TUI — Interactive File Manager

Launch with `bolt` or `bolt ui [path]`.

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `↵` / `l` | Open file or enter directory |
| `⌫` / `h` | Go to parent directory |
| `g` / `Home` | Jump to top |
| `G` / `End` | Jump to bottom |
| `Ctrl+U` / `PgUp` | Page up |
| `Ctrl+D` / `PgDn` | Page down |
| `~` | Jump to home directory |
| `[` | Navigate back (history) |
| `]` | Navigate forward (history) |
| `P` | Go to typed path |
| `Ctrl+R` | Recent files panel |

### Selection & Clipboard

| Key | Action |
|-----|--------|
| `Space` | Toggle select file |
| `Ctrl+A` | Select / deselect all |
| `c` | Copy selected to clipboard |
| `x` | Cut selected to clipboard |
| `p` | Paste clipboard here |

### File Operations

| Key | Action |
|-----|--------|
| `r` | Rename current file |
| `R` | Batch rename selected (`{name}`, `{ext}`, `{n}`, `{N}`) |
| `m` | Create new directory |
| `n` | Create new file |
| `u` | Duplicate current file |
| `L` | Create symlink to current file |
| `Z` | Change permissions (octal input) |
| `D` | Permanently delete |
| `Ctrl+T` | Move to trash (recoverable) |
| `Ctrl+Z` | Undo last operation |
| `y` | Copy full path to status bar |

### Editor & View

| Key | Action |
|-----|--------|
| `e` | Open in `$EDITOR` / `$VISUAL` |
| `i` | Toggle extended info (perms, symlink target) |
| `d` | Diff — select 2 files then press `d` |
| `#` | SHA-256 checksum (shown in status bar) |
| `F` | Find duplicate files in current directory |

### Tags

| Key | Action |
|-----|--------|
| `t` | Add tag to current file |
| `T` | Remove last tag from current file |

Tags are stored in `~/.config/bolt/config.toml` and visible in the preview pane.

### Search & Filter

| Key | Action |
|-----|--------|
| `/` | Live fuzzy filter current directory |
| `Ctrl+F` | Recursive file + content search |
| `s` | Cycle sort mode (name → size → date → type) |
| `.` | Toggle hidden files |

### App

| Key | Action |
|-----|--------|
| `:` | Command mode (run any `bolt` subcommand) |
| `?` | Help screen |
| `q` / `Ctrl+C` | Quit |

### Batch Rename Patterns

When multiple files are selected, press `R` and enter a pattern:

| Token | Replaced with |
|-------|---------------|
| `{name}` | Original filename without extension |
| `{ext}` | Extension including dot (`.jpg`) |
| `{n}` | Index (1, 2, 3…) |
| `{N}` | Zero-padded index (001, 002, 003…) |

**Example:** select 3 photos, pattern `photo_{N}{ext}` → `photo_001.jpg`, `photo_002.jpg`, `photo_003.jpg`

---

## CLI Commands

All commands are also available standalone:

### File Operations

```bash
bolt ls [path]                    # list directory
bolt ls --tag work                # filter by tag
bolt cp <src> <dst>               # copy (use -r for dirs)
bolt mv <src> <dst>               # move / rename
bolt rm <path>                    # delete permanently
bolt mkdir <name>                 # create directory
bolt touch <file>                 # create/update file
bolt pwd                          # print working directory
bolt duplicate <file>             # copy with _copy suffix
bolt symlink <target> <link>      # create symbolic link
bolt chmod <octal> <file>         # change permissions
bolt info <path>                  # detailed file info
```

### Trash

```bash
bolt trash <file>                 # move to trash
bolt trash-list                   # list trashed files
bolt trash-restore <name> <dest>  # restore from trash
bolt trash-empty [-f]             # permanently delete trash
```

### Search & Find

```bash
bolt search <pattern>             # fuzzy file search
bolt search <pattern> --content   # include content search
bolt search <pattern> --regex     # regex mode
bolt find <pattern> [path]        # find files by name
bolt grep <pattern> [path]        # grep file contents
bolt index build [path]           # build search index
bolt index status                 # index info
bolt index clean                  # remove index
```

### Preview & Diff

```bash
bolt preview <file>               # syntax-highlighted preview
bolt preview -t dracula <file>    # with theme
bolt preview -n 50 <file>         # first 50 lines
bolt preview -p <file>            # plain (no highlighting)
bolt diff <file-a> <file-b>       # colorized unified diff
bolt diff --no-color <a> <b>      # plain diff output
```

### Disk & Analysis

```bash
bolt du [path]                    # disk usage breakdown
bolt du -n 10 [path]              # top 10 entries
bolt dupes [path]                 # find duplicate files
bolt checksum <file>              # SHA-256 (default)
bolt checksum -a md5 <file>       # MD5
bolt checksum -a sha512 <file>    # SHA-512
bolt recent                       # recently visited paths
```

### Tags

```bash
bolt tag add <path> <tag>         # tag a file
bolt tag remove <path> <tag>      # remove a tag
bolt tag list <path>              # list file's tags
bolt ls --tag <tag>               # list files with tag
```

### Archives

```bash
bolt zip <out.zip> <files...>     # create zip
bolt unzip <file.zip> [dest]      # extract zip
bolt tar create <out.tar.gz> <files...>   # create tar.gz
bolt tar extract <file.tar.gz> [dest]     # extract tar.gz
bolt tar list <file.tar.gz>               # list contents
```

### Bookmarks

```bash
bolt bm add <name> [path]         # save bookmark
bolt bm go <name>                 # cd to bookmark
bolt bm list                      # list bookmarks
bolt bm rm <name>                 # delete bookmark
```

### Watch

```bash
bolt watch [path]                 # watch for file changes
```

### Shell Completion

```bash
bolt completion bash   > ~/.bash_completion.d/bolt
bolt completion zsh    > ~/.zsh/completions/_bolt
bolt completion fish   > ~/.config/fish/completions/bolt.fish
bolt completion powershell >> $PROFILE
```

---

## Configuration

Config lives at `~/.config/bolt/config.toml`. Created automatically on first run.

```toml
[core]
cache_dir  = "~/.cache/bolt"
config_dir = "~/.config/bolt"
data_dir   = "~/.local/share/bolt"

[ls]
default_sort = "name"   # name | size | date | type
show_hidden  = false
color        = true

[tags.files]
# auto-managed by bolt tag / TUI tag operations

[plugins]
# [plugins.my-tool]
# enabled     = true
# command     = "/usr/local/bin/my-tool"
# description = "my custom plugin"
```

### Plugins

Any external binary can be registered as a Bolt command:

```toml
[plugins.fzf-open]
enabled     = true
command     = "/usr/local/bin/fzf-open"
description = "open file via fzf"
```

Then run with `bolt fzf-open` or from the TUI command bar (`:fzf-open`).

---

## Data Locations

| Data | Location |
|------|----------|
| Config | `~/.config/bolt/config.toml` |
| Search index | `~/.cache/bolt/` |
| Trash | `~/.local/share/bolt/trash/` |
| Recent files | `~/.local/share/bolt/recent.json` |
| Bookmarks | stored in config.toml |

---

## Themes

File preview supports all [Chroma](https://github.com/alecthomas/chroma) themes:

```bash
bolt preview -t monokai        <file>
bolt preview -t dracula        <file>
bolt preview -t github         <file>
bolt preview -t solarized-dark <file>
bolt preview -t nord           <file>
bolt preview -t one-dark       <file>
```

---

## License

MIT © 2026 David Ogar
