# gh-tree

`gh-tree` is a keyboard-first GitHub CLI extension for browsing open pull
requests and branches as a namespace tree, remembering a folder per repository,
and safely deploying an exact PR head into a persistent local test worktree.

## Install

Install [GitHub CLI](https://cli.github.com/) and authenticate once:

```bash
gh auth login
gh extension install Hans-Einar/gh-tree
```

No Go compiler or repository clone is required. Upgrade later with:

```bash
gh extension upgrade tree
```

## Use

Run inside a local GitHub repository:

```bash
gh tree
```

Or browse a repository explicitly (worktree deployment remains disabled unless
the current directory belongs to that same repository):

```bash
gh tree --repo Hans-Einar/ponsse
```

Conceptual view:

```text
gh tree  Hans-Einar/ponsse  [PRs]  branches
/MVP1/

> machine-service/
  simulator/

                              #60  UIBox clean-cut rewrite
                              head: steering/MVP1/ui-box
                              base: main
                              sha:  3c83ea2d4e3ba071b5a6648129c6ebb136db8912
                              DRAFT

Worktrees (2)
  ponsse          → main @ a17b99de [current]
  ponsse-MVP1     → local/mvp1-test @ 3c83ea2d

[Enter] open  [Backspace] parent  [p] PRs  [b] branches
[w] deploy  [/] search  [r] refresh  [q] quit
```

### Keyboard shortcuts

| Key | Action |
| --- | --- |
| `↑`/`↓`, `k`/`j` | Move selection |
| `Enter` | Open a namespace or select an item |
| `Backspace` | Move to the parent namespace |
| `p` / `b` | Show pull requests / branches |
| `/` | Edit the filter; `Enter` accepts, `Ctrl+U` clears |
| `r` | Refresh GitHub and local worktree state |
| `w` | Deploy the selected PR to a configured test worktree |
| `q`, `Ctrl+C` | Quit (`q` types normally while search is active) |

The branch view intentionally loads at most 100 branches per refresh. This keeps
the v1 interaction bounded and API-friendly in large repositories.

## Namespace behavior

The PR head branch is authoritative. Leading technical prefixes are removed and
the remaining `/`-separated segments become folders:

```text
steering/Concept1/ui-box          → Concept1/ui-box
codex/MVP1/machine-service/slc003 → MVP1/machine-service/slc003
review/emulator/timer2-slc015     → emulator/timer2-slc015
```

Unstructured branch names appear under `misc/`. Unknown prefixes are preserved,
so the behavior is generic rather than tied to a particular project. The last
visited folder is stored per `owner/repo`; a deleted folder falls back to its
nearest existing ancestor and then to `/`.

## Configuration and state

Browsing works with no configuration. `gh-tree` uses the standard per-user
configuration directory returned by the operating system:

| Platform | Configuration | Navigation state |
| --- | --- | --- |
| Linux | `$XDG_CONFIG_HOME/gh-tree/config.json` or `~/.config/gh-tree/config.json` | `$XDG_STATE_HOME/gh-tree/state.json` or `~/.local/state/gh-tree/state.json` |
| macOS | `~/Library/Application Support/gh-tree/config.json` | `~/Library/Application Support/gh-tree/state.json` |
| Windows | `%AppData%\gh-tree\config.json` | `%AppData%\gh-tree\state.json` |

Override these paths with `--config PATH` and `--state PATH`. Navigation state is
managed automatically; do not commit either file to a repository.

Example `config.json`:

```json
{
  "stripPrefixes": [
    "steering",
    "codex",
    "worker",
    "review",
    "agent",
    "fix",
    "feature"
  ],
  "repos": {
    "Hans-Einar/ponsse": {
      "worktrees": {
        "Concept1": {
          "path": "C:\\Users\\hanse\\GIT\\ponsse-Concept1",
          "branch": "local/concept1-test"
        },
        "MVP1": {
          "path": "C:\\Users\\hanse\\GIT\\ponsse-MVP1",
          "branch": "local/mvp1-test"
        }
      }
    }
  }
}
```

Use absolute worktree paths. Omitting `stripPrefixes` enables the defaults shown
above; an explicit empty array disables prefix stripping.

## Safe worktree deployment

Create each persistent test worktree once, preferably on its configured local
test branch:

```bash
git worktree add -b local/concept1-test ../ponsse-Concept1 main
```

Select a PR, press `w`, choose a configured target, and confirm the displayed PR
number, full head SHA, path, and local branch. `gh-tree` then:

1. verifies that the target is a registered, non-primary, non-current worktree
   belonging to this repository;
2. refuses a dirty target before changing anything;
3. fetches `refs/pull/<number>/head` from `origin` into a private local ref;
4. verifies that the fetched commit exactly matches the SHA shown in the TUI;
5. rechecks cleanliness, then moves only the configured local test branch;
6. verifies the resulting branch, exact SHA, and clean status.

A clean detached target is migrated when its current commit is already reachable
from a local or remote branch. An unreferenced detached commit is refused with an
instruction to preserve it on a branch first. Deployment is also rejected if the
configured branch is checked out in another worktree. The main/current worktree
is never reset, dirty changes are never discarded, and the local test branch is
never pushed.

## Development

Go 1.25 or newer is required for source builds:

```bash
go test ./...
go vet ./...
go build ./cmd/gh-tree
```

The integration tests create temporary bare repositories and real linked
worktrees; they do not touch the repository in which the tests are run.

## Release

CI tests Linux, macOS, and Windows and cross-compiles the required release
targets. A `v*` tag starts `.github/workflows/release.yml`, which uses the
official `cli/gh-extension-precompile@v2` action to publish GitHub CLI-compatible
precompiled assets (including Windows amd64, Linux amd64/arm64, and macOS
amd64/arm64):

```bash
git tag v0.1.0
git push origin v0.1.0
```

After the release workflow succeeds, installation and upgrades use those
binaries and do not compile locally.
