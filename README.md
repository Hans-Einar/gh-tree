# gh-tree

`gh-tree` is a keyboard-first GitHub CLI extension for navigating pull requests,
branches, commits and local Git worktrees from one TUI. It reuses `gh`
authentication and keeps destructive Git operations behind explicit safety checks.

## Install

```bash
gh auth login
gh extension install Hans-Einar/gh-tree
```

Upgrade later with:

```bash
gh extension upgrade tree
```

Run it inside a GitHub repository:

```bash
gh tree
```

or browse explicitly:

```bash
gh tree --repo Hans-Einar/ponsse
```

Worktree operations are enabled when the current directory belongs to the same
repository being viewed.

## The v2 cockpit

On a normal-width terminal the UI is arranged as:

```text
┌────────────────────────────────┬────────────────────────────────┐
│ PR / branch navigator          │ Local worktrees                │
│                                │                                │
│ Concept1/                      │ > ponsse       main            │
│ MVP1/                          │   ponsse-C1    feature/...     │
│ > #60 UIBox                    │   ponsse-MVP1  DETACHED        │
│                                │                                │
├────────────────────────────────┴────────────────────────────────┤
│ Selected PR / branch identity + active-worktree status          │
│ head: steering/Concept1/ui-box                                  │
│ base: main                                                       │
│ sha:  <full SHA, never ellipsized>                              │
│                                                                  │
│ path: C:\...\ponsse-C1                                         │
│ branch: steering/Concept1/ui-box                                │
│ working: CLEAN                                                   │
│ remote: origin/... · ahead 0 / behind 2                         │
│ PR: #60 DRAFT · HEAD matches PR                                 │
└──────────────────────────────────────────────────────────────────┘
```

`Tab` / `Shift+Tab` moves focus between panes. The footer changes to show the
commands that apply to the focused pane. Narrow terminals stack the panes.

## Git mental model

The TUI deliberately exposes the relationship between the Git objects:

```text
worktree
   │
   ▼
 HEAD ──normally──> local branch ──> commit
   └──detached─────────────────────> commit
```

A **commit** is an immutable history node. A **branch** is just a movable name
pointing to one commit, normally the newest commit on that line of development.
When you commit while `HEAD` is attached to a branch, that branch moves to the
new commit.

A **worktree** is a physical checkout directory. The same repository can have
several worktrees at once, each on a different branch or at a detached commit.
Git normally prevents the same local branch from being checked out in two
worktrees simultaneously.

A GitHub **pull request** compares a **head branch** against a **base branch**:

```text
feature branch ── PR head ──┐
                            ├──> proposed merge into main
main branch    ── PR base ──┘
```

Commits are made to the branch, not to the PR object. After those commits are
pushed, an open PR based on that branch updates automatically.

## Navigator

PRs are grouped by namespace derived from their head branch. Known technical
prefixes are stripped:

```text
steering/Concept1/ui-box          → Concept1/ui-box
codex/MVP1/machine-service/slc003 → MVP1/machine-service/slc003
review/emulator/timer2-slc015     → emulator/timer2-slc015
```

Branches show an associated open PR when one exists.

Navigator keys:

| Key | Action |
| --- | --- |
| `↑`/`↓`, `k`/`j` | move selection |
| `Enter` | open namespace / select item |
| `Backspace` | parent namespace |
| `p` / `b` | PR / branch mode |
| `h` | open commit history for selected PR/branch |
| `/` | filter |
| `r` | refresh GitHub + worktree state |
| `Tab` | focus next pane |

The branch/API listing remains bounded to keep refreshes predictable on large
repositories.

## Worktree manager

Existing worktrees come directly from `git worktree list --porcelain`; no
`config.json` entry is required. If only the primary checkout exists, focus the
worktree pane and press `c` to create another one.

Worktree-pane commands:

| Key | Action |
| --- | --- |
| `Enter` / `a` | make selected worktree active inside `gh-tree` |
| `c` | create a worktree from the selected PR/branch/commit |
| `x` | check out the selected PR/branch in the active secondary worktree |
| `f` | fetch/prune remote state |
| `p` | fast-forward-only pull of the active tracking branch |
| `m` | stage all current changes and commit with an entered message |
| `P` | push the active branch; first push can set upstream |
| `n` | create a new branch in the active worktree |
| `o` | create a draft GitHub PR from the active pushed branch |
| `h` | browse commits reachable from active `HEAD` |

The active worktree is an internal working directory for `gh-tree`. A child
process cannot change the parent CMD/PowerShell/Bash process's current directory,
so activating a worktree does not `cd` the shell that launched the extension.

### Creating worktrees

The create dialog suggests a sibling path beside the primary repository. It can
create from:

- selected PR head (the PR ref is fetched and its exact SHA is verified);
- selected local/remote branch;
- active commit;
- a historical commit selected in the commit browser.

Leave the branch field empty when intentionally creating a detached historical
checkout. Otherwise a normal local branch is preferred for continued work.

### Retargeting

`x` changes a secondary clean worktree to the PR or branch selected in the left
pane. `gh-tree` refuses dirty worktrees and protects the primary worktree. PR
checkout fetches `refs/pull/<number>/head`, verifies that the fetched SHA still
matches the one displayed in the TUI, then creates the local worktree branch.

## Active-worktree status

The lower panel shows:

- absolute path;
- primary/current flags;
- branch or `DETACHED HEAD`;
- full `HEAD` SHA;
- clean/dirty state with staged/modified/untracked/conflict counts;
- upstream tracking branch;
- ahead/behind counts from the fetched local remote-tracking state;
- associated open PR and whether local `HEAD` matches its head SHA.

Remote status is only as current as the last fetch/refresh; use `f` when the
network-visible state matters.

## Commit browser

Press `h` on a PR, branch or active worktree. The left pane becomes a bounded
commit list and the right pane shows the selected commit's full metadata and
message.

```text
┌────────────────────────────┬───────────────────────────────────┐
│ Commits                    │ Commit details                    │
│ > 4d8d72d  fix layout      │ commit: 4d8d72d...               │
│   a3b1940  refactor model  │ parents: ...                     │
│   93f024a  add tests       │ author/date                       │
│                            │                                   │
│                            │ full scrollable commit message    │
└────────────────────────────┴───────────────────────────────────┘
```

Use `Tab` to focus the details pane and arrows/PageUp/PageDown to scroll the
message. `L` loads another bounded page. From a historical commit, `c` creates a
worktree and `n` creates a branch starting at that commit.

## Safety model

Hard guarantees:

- dirty worktrees are never silently discarded;
- browsing never resets the primary worktree;
- PR identity is checked by exact head SHA before PR-backed checkout/deploy;
- branch names and paths are passed as process arguments, never interpolated
  into shell command strings;
- pull is fast-forward-only in the built-in command path;
- push never force-pushes;
- first upstream creation is shown before confirmation;
- detached HEAD is shown explicitly;
- Git's one-worktree-per-branch rule is allowed to fail clearly rather than
  being bypassed;
- command failures remain visible instead of reporting false success.

The original v1 configured deployment (`w`) remains available for users who
already have named test targets in config. With no legacy targets, `w` now sends
you to the interactive worktree pane instead of telling you to hand-edit JSON.

## Configuration and state

Browsing and worktree discovery need no configuration. Paths use the operating
system's user config/state locations:

| Platform | Configuration | State |
| --- | --- | --- |
| Linux | `$XDG_CONFIG_HOME/gh-tree/config.json` or `~/.config/gh-tree/config.json` | `$XDG_STATE_HOME/gh-tree/state.json` or `~/.local/state/gh-tree/state.json` |
| macOS | `~/Library/Application Support/gh-tree/config.json` | same application-support tree |
| Windows | `%AppData%\gh-tree\config.json` | `%AppData%\gh-tree\state.json` |

`stripPrefixes` can override the default namespace prefixes. Legacy v1
`repos.<owner/repo>.worktrees` targets remain supported, but are not needed for
ordinary v2 worktree discovery/creation.

## Development

Go 1.25 or newer:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/gh-tree
```

Integration tests create temporary bare repositories and real linked worktrees;
they do not modify the repository in which the tests run.

## Release

CI tests Windows, Linux and macOS and cross-builds the supported targets. A `v*`
tag triggers `.github/workflows/release.yml`, using
`cli/gh-extension-precompile@v2` to publish GitHub CLI-compatible binaries.

After a release:

```bash
gh extension upgrade tree
```
