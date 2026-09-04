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

The normal cockpit keeps PR/branch navigation on the left, Git-owned local
worktrees on the right, and selected/active-worktree status below. `Tab` and
`Shift+Tab` move focus; the footer is context-sensitive.

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

Dirty worktrees block retargeting. The primary worktree is protected. Exact PR
head SHA is verified before PR-backed checkout/deployment.

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

Keys:

| Key | Action |
| --- | --- |
| `F5` | run default launch point |
| `Ctrl+F5` | discover / choose launch point |
| `Shift+F5` | stop attached launch |
| `F6` | restart current launch |

The active worktree is always the launch cwd. Output is captured in a bounded
log buffer and running/exit/failure state is shown in the cockpit. Attached
launches are stopped when `gh-tree` exits.

### npm / pnpm / yarn

`package.json` scripts are discovered automatically. A name such as `dev:wan`
is **one exact script name**:

```bash
npm run dev:wan
```

The chooser may display it hierarchically as `npm / dev / wan`, but never splits
it into multiple commands. A pnpm or yarn lockfile selects that package manager
for the saved invocation.

### Make

Simple Make targets are discovered from `Makefile`, `makefile` or `GNUmakefile`.
Multiple selected targets form one ordered native invocation. The UI may show:

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
committable and contains launch intent, not machine-specific worktree paths:

```json
{
  "default": "dev-wan",
  "launch": {
    "dev-wan": {
      "provider": "npm",
      "script": "dev:wan"
    },
    "release": {
      "provider": "make",
      "targets": ["clean", "all", "install"]
    }
  }
}
```

Configuration is validated before execution and commands are always started as
an executable plus argument vector, not as shell-concatenated strings.

## Safety

`gh-tree` keeps conservative defaults:

- never silently discard dirty/untracked worktree data;
- never reset the primary worktree as a browsing side effect;
- never force-push;
- exact SHA/ref verification for PR-backed operations;
- fast-forward-only built-in pull;
- selective staging paths are repository-relative and path-escape checked;
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
