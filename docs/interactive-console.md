# Interactive Console (v0.3.13)

`gh-tree` Console supports both launch-process tabs and real interactive shell tabs.

## Open a terminal

Press `Alt+T`. The new terminal starts in the current **Active worktree** and stays bound to that working directory even if another worktree is activated later.

The shell is selected automatically from the process ancestry where possible. This matters because a GitHub CLI extension is normally started as `shell -> gh -> gh-tree`.

- `cmd.exe` ancestry -> Command Prompt
- `powershell.exe` ancestry -> Windows PowerShell
- `pwsh.exe` ancestry -> PowerShell 7
- `bash`, `zsh`, `fish`, or `sh` ancestry -> the same Unix shell
- Windows fallback -> `COMSPEC`, then `cmd.exe`
- Unix fallback -> `SHELL`, then `sh`

Set `GH_TREE_SHELL` to an executable path/name to override automatic detection.

## Console tabs

Launches and interactive terminals share the Console tab bar. `Alt+1` through `Alt+9` select the first nine tabs. `Alt+T` can create additional shell tabs.

Interactive terminal sessions use a PTY on Unix and ConPTY on Windows. The PTY is resized with the Console pane.

While an interactive terminal has focus, normal keyboard input is forwarded to the terminal, including printable text, Enter, arrows, Tab completion and common control characters. `Ctrl+C` is sent to the PTY as an interrupt instead of exiting `gh-tree`.

Global `gh-tree` Alt mnemonics remain reserved, so `Alt+N`, `Alt+W`, `Alt+A`, `Alt+O`, and the console-tab shortcuts always provide a route back to the cockpit. Outside Console, `Ctrl+C` opens the normal centered exit-confirmation dialog.

`Shift+F5` remains the explicit stop shortcut for launch-process tabs. Close an interactive shell normally with its `exit` command.

All attached launch and terminal sessions are stopped when `gh-tree` exits.

## Rendering note

The session is backed by a real PTY/ConPTY, so shells and TTY-aware command-line programs receive terminal semantics. The current Console viewport stores bounded, line-oriented output rather than implementing a complete VT screen emulator; full-screen alternate-screen applications are therefore not yet the primary target.

The main TUI header shows the running `gh-tree` version at the upper-right so screenshots and bug reports identify the executable immediately.
