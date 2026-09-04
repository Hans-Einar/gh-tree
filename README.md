# gh-tree

`gh-tree` is a keyboard-first GitHub CLI extension for navigating the real Git
history, pull requests, branches and local worktrees, reviewing changes, and
launching known development targets from the active worktree.

## Install / upgrade

```bash
gh auth login
gh extension install Hans-Einar/gh-tree
# existing install:
gh extension upgrade tree
```

Run inside a GitHub repository:

```bash
gh tree
```

The normal cockpit keeps PR/branch navigation on the left, contextual repository
information on the right, and the active-worktree state below. The footer is
context-sensitive.

## Focus and pane navigation

`Tab` moves between the main panes. In the branch-context pane, `Shift+Tab`
cycles the local subpanes (Branch / Commits / Message) instead of forcing every
subpane into the main Tab loop. `Ctrl+Shift+Tab` is accepted as a reverse-main-
focus shortcut where the terminal reports it distinctly.

YaST/ncurses-style mnemonics provide direct jumps; the mnemonic letter is
highlighted in the pane title:

| Key | Destination |
| --- | --- |
| `Alt+N` | Navigator |
| `Alt+W` | Worktrees |
| `Alt+A` | Active worktree root |
| `Alt+L` | Launch list in the active-worktree pane |
| `Alt+B` | Branch context / metadata |
| `Alt+C` | Branch commit-list subpane |
| `Alt+M` | Commit-message subpane |

The active heading or subheading remains visibly highlighted while it owns
keyboard focus. `Alt+A`, followed by `Enter`, opens the active-worktree chooser.

## Floating dialogs

Input and confirmation dialogs use a centered floating modal layer. The existing
cockpit remains behind the modal in a dimmed/shaded state instead of being
pushed upward or resized. Dialogs prefer a wide, low layout and bounded list
content so the appearance is obvious even on large terminals.

This applies to worktree selection and the normal create/launch/confirmation
flows. While a modal is open, background panes are inert until the modal is
accepted or cancelled.

## Git graph

Press `g` from the navigator/details area to open the real commit DAG inside the
main TUI, or start directly with:

```bash
gh tree --graph
```

The graph is built from commit parent relationships and decorates commits with:

- `HEAD`;
- local branches (`refs/heads/*`);
- remote-tracking branches (`refs/remotes/*`);
- tags;
- open PR heads;
- local worktrees.

Branches are labels on commits, not separate history nodes. Remote-tracking refs
are explicitly treated as local observations: they may be stale until a fetch.
The graph is bounded and `L` loads more history.

A compact mental model:

```text
worktree -> HEAD -> local branch -> commit -> parent commit(s)
                  ^
                  | push/fetch relationship
                  v
             origin/branch

PR = GitHub comparison of a head branch against a base branch
```

## Branch context and commit browsing

In branch mode, `Enter` on a branch opens that branch's context in the right
pane. The context contains fixed sections for branch identity, a bounded commit
list, and the selected commit's message. Long messages scroll inside their own
viewport; they do not resize the whole TUI.

`Alt+B`, `Alt+C`, and `Alt+M` focus Branch context, Commits, and Message
respectively. The focused subheading is visibly highlighted until focus moves.

While the Commits subpane is active:

| Key | Action |
| --- | --- |
| `↑` / `↓` | select commit |
| `c` | create a new worktree from the selected historical commit |
| `x` | checkout the selected commit detached into the active secondary worktree |

Historical checkout reuses the normal worktree safety gate: the primary
worktree is protected and dirty worktrees are refused.

Branch-list PR annotations show direction relative to the branch:

```text
main                 < [PR #60]
feature/ui           [PR #60] >
integration          < [PR #72]  [PR #73] >
```

`< [PR #N]` means this branch is the PR base (the PR flows into this branch).
`[PR #N] >` means this branch is the PR head (the PR flows out of this branch).

## Worktrees

Existing worktrees are discovered from `git worktree list --porcelain`; no
manual target configuration is needed. The active worktree is an internal cwd
used by `gh-tree` operations (a child process cannot change the parent shell's
cwd).

Important keys while the worktree pane is focused:

| Key | Action |
| --- | --- |
| `Enter` / `a` | activate selected worktree |
| `c` | create worktree from selected PR/branch/commit |
| `x` | safely retarget selected secondary worktree |
| `f` | fetch/prune remote state |
| `p` | fast-forward-only pull |
| `m` | commit |
| `P` | push (never force-push) |
| `n` | new branch |
| `o` | create draft PR |
| `h` | commit history |
| `d` | worktree diff |
| `D` | staged diff |
| `z` / `Z` | stash / pop latest stash with confirmation |

`Alt+A` focuses the Active worktree root. Press `Enter` there to open a floating
chooser listing all discovered worktrees and activate one without moving to the
Worktrees pane first.

Dirty worktrees block retargeting. The primary worktree is protected. Exact PR
head SHA is verified before PR-backed checkout/deployment.

## Dirty-worktree inspection and cleanup

When the active worktree is dirty, the lower pane lists the paths responsible
for that state and distinguishes staged, working-tree, untracked and conflicted
changes. Focus the Active worktree pane to operate on the selected path:

| Key | Action |
| --- | --- |
| `s` | stage selected path |
| `u` | unstage selected path |
| `d` | open the worktree diff |
| `z` | stash tracked + untracked changes |
| `r` | discard only the selected tracked working-tree change, after confirmation |

Cleanup is deliberately conservative. `r` refuses untracked files and conflict
entries; untracked files are never silently deleted and conflicts are never
auto-resolved.

## Diff / review

Press `d` on a selected PR or commit, or `d`/`D` in the worktree pane. Diff mode
shows changed files alongside a scrollable patch. Supported sources include:

- selected commit vs first parent;
- active worktree vs `HEAD`;
- staged changes vs `HEAD`;
- selected PR head vs its base using the fetched/verified private PR ref.

Text patches are bounded; binary files are represented as binary instead of
being dumped. In mutable worktree/staged views, `s` stages the selected file and
`u` unstages it. Pure PR/commit review remains read-only.

## Launch points — F5

`gh-tree` detects launch points using build-system providers. One launch point is
one native invocation/process; `gh-tree` is deliberately not a generic command
orchestrator.

The Active worktree pane shows launch points discovered for that worktree.
`Alt+L` jumps directly to the Launch list; use `↑`/`↓` to select and `Enter` to
run the selected native launch point.

Keys:

| Key | Action |
| --- | --- |
| `Alt+L` | focus discovered launch list |
| `F5` | run default launch point |
| `Ctrl+F5` | discover / choose launch point |
| `Shift+F5` | stop attached launch |
| `F6` | restart current launch |

Launch discovery scans the active worktree for actual provider manifests, not
lockfiles alone. This matters for repositories whose runnable project is below
the repository root. For example:

```text
ponsse-pr-60/
  package-lock.json       # not an npm project by itself
  Concept1/
    package.json          # detected npm project root
    package-lock.json
```

In that layout `gh-tree` discovers `Concept1` and runs its npm scripts with
`Concept1` as the process cwd. Common generated/dependency directories such as
`.git`, `node_modules`, `vendor`, `dist`, `build`, `target` and `.cache` are not
recursively scanned.

Output is captured in a bounded log buffer and running/exit/failure state is
shown in the cockpit. Attached launches are stopped when `gh-tree` exits.

### npm / pnpm / yarn

`package.json` scripts are discovered automatically, including package manifests
below the worktree root. A name such as `dev:wan` is **one exact script name**:

```bash
npm run dev:wan
```

The chooser may display it hierarchically as `npm / dev / wan`, but never splits
it into multiple commands. A pnpm or yarn lockfile beside that project's
`package.json` selects that package manager for the saved invocation.

### Make

Simple Make targets are discovered from `Makefile`, `makefile` or `GNUmakefile`,
including nested project roots. Multiple selected targets form one ordered native
invocation. The UI may show:

```text
clean : all : install
```

which executes exactly:

```bash
make clean all install
```

Target order is significant.

## `.gh-tree/run.json`

A discovered launch can be saved from the chooser. The repository-local file is
committable and contains launch intent, including a repository-relative project
cwd when the manifest is below the worktree root:

```json
{
  "default": "Concept1/dev:wan",
  "launch": {
    "Concept1/dev:wan": {
      "provider": "npm",
      "dir": "Concept1",
      "script": "dev:wan"
    },
    "release": {
      "provider": "make",
      "targets": ["clean", "all", "install"]
    }
  }
}
```

The `dir` value is validated as repository-relative before execution; it cannot
escape the active worktree. Commands are always started as an executable plus
argument vector, not as shell-concatenated strings.

## Safety

`gh-tree` keeps conservative defaults:

- never silently discard dirty/untracked worktree data;
- never reset the primary worktree as a browsing side effect;
- never force-push;
- exact SHA/ref verification for PR-backed operations;
- fast-forward-only built-in pull;
- selective staging paths are repository-relative and path-escape checked;
- tracked discard requires explicit confirmation and refuses untracked/conflicts;
- stash operations are explicit and conflicts are surfaced, not auto-resolved;
- remote-tracking refs are not described as current without a fetch;
- launch providers build argument vectors; discovered commands do not run until
  explicitly selected/F5-launched.

## Development

Go 1.25 or newer:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/gh-tree
```

CI covers Windows, Linux and macOS plus release cross-builds. Stable `v*` tags
are published with `cli/gh-extension-precompile@v2` for direct installation via
GitHub CLI.
