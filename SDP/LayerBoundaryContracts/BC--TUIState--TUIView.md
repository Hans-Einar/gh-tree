# BC--TUIState--TUIView

State: FROZEN
Version: 1.0.0

Freeze: [BCFreeze--001](BCFreeze--001.md); effective after reviewed PR #56 merge.
Authority: #55/#21; accepted REFDES #52/PR54, merge4a42222f7bfedc1d80693effbb25a1a82fcff65e.
Applies to: presentation and interaction transport of SLC-01..12, version display in13.
Supersedes: none; no implementation authority before the seven-contract freeze.

## Responsibilities and dependency direction

State is the one deterministic ephemeral interaction authority. View owns pure
layout, cell/text/graph/console interpretation, style, labels, overlays, help and
animation frames from supplied values. Application supplies semantic facts through
its API; View never infers business relations from disconnected branch names/OIDs.
The restricted Composition host maps Tea inputs/effects, drives one disposable
presentation timer, owns output interpretation caches and applies measurements.

`tuistate -> application/api + domain + tuistate/viewmodel`;
`tuiview -> tuistate/viewmodel + approved pure rendering libraries`;
`tuistate/viewmodel -> domain + pure standard values`.
State never imports View; the host calls both. Viewmodel imports neither API nor
State parent nor ports/adapters. View has no Client, callbacks, timer/clock,
filesystem/network/process access or mutable backend objects. No global styles
are mutated, including during construction. Color/terminal profile is supplied
explicitly; rendering libraries must not discover it by implicit I/O.

[BoundaryTypes--001](BoundaryTypes--001.md) and
[accepted REFDES](../Design/CR-%2321/REFDES--001.md) govern common scope/copy rules.
Application/State BC owns the public API; this boundary contains presentation
values only, never a second product command protocol.

## Pure entry points

```go
// package tuistate
func Update(State, Input) (State, []Effect)
func Project(State) viewmodel.Snapshot

// package tuiview
func Layout(viewmodel.Snapshot, viewmodel.Viewport) viewmodel.Measurement
func Render(viewmodel.Snapshot, viewmodel.Measurement, Theme) string
func InterpretOutput(ConsoleParserState, viewmodel.OutputInput) (ConsoleParserState, viewmodel.ConsolePresentation)
```

All functions are deterministic for their explicit inputs. Result snapshots,
measurements and parser state own their mutable data; repeated Render changes no
controller or backend state. Render requires a measurement matching the snapshot's
presentation/viewport generation; stale measurement returns bounded unavailable
presentation rather than silently resizing a different session. It may recompute
the same pure layout when the contract explicitly supplies both original inputs.
No asynchronous effect runs from Layout/Render/InterpretOutput.

## Viewmodel shape and identity

Viewmodel owns its own optional values and local presentation IDs, avoiding an API
dependency. Stable semantic identity embedded in rows uses Domain values; position
and display label are not selection/operation identity.

| Record | Required fields and semantics |
|---|---|
| Snapshot | presentation generation, Mode, FocusPath, selected semantic ElementIDs, typed pane models, one optional Modal, applicable ActionBindings, status/notices, supplied animation frame and version display |
| Viewport | Known bool, Width/Height int, generation; only unknown initial dimensions use fallback. Known0/very small is a real bound. |
| FocusPath | closed pane/subpane/input path consistent with Mode/Modal. View receives it; title text never determines focus. |
| ElementID | closed namespace-key/RepositoryID/PRID/BranchID/Revision/WorktreeID/StashID/LaunchPointID+optional exact saved alias/SessionID value. No arbitrary backend pointer. |
| PaneModel | semantic rows/details and availability/completeness, selected identity, scroll/filter/cursor values; no inference of absence from partial lists |
| BranchRow / WorktreeRow / PRRow | exact IDs/revisions and already reconciled relation/status/unknown badges; View formats these facts only |
| GraphModel | accumulated commits/parent edges/exact refs and supplied semantic PR/worktree annotations, source generation, selection and scroll. Commit subject/message remain verbatim facts. |
| DiffModel | exact comparison/endpoints, file changes, bounded patch/binary/truncation metadata, selection and scroll; working-tree source label accurately means index-to-worktree |
| LaunchRow | LaunchPointID, optional exact saved alias, source/project/provider labels, default/availability and ordered selected-member display; no invocation builder |
| ConsoleModel | SessionID, lifecycle/capabilities, current input focus, supplied safe display summary, copied ConsolePresentation and current source/presentation generations |
| Modal | local ModalID, owner intent/operation presentation key, closed purpose/body variant, exact target details, allowed semantic choices and body scroll; no ConfirmationID usable by View to approve anything |
| ActionBinding | semantic ActionID, actual key/chord, current scope/applicability/enabled state and label; State is the sole action authority |
| StatusNotice | safe text, severity and exact subject identity where relevant; no raw exception/backend callback |

Application OperationID/ConfirmationID are not imported into viewmodel. State binds
them privately to the modal/effects and projects an independent ModalID and safe
identity text/Domain targets. Layout correlates only that local ModalID and
viewport/presentation generation. Exact40/64-digit revision and stable stash
identity remain inspectable in confirmation/detail views; abbreviated labels
cannot replace the bound target.

## Layout, measurement and bounded degradation

```go
// package viewmodel
type Rect struct { X, Y, Width, Height int }
type ConsoleRect struct {
    SessionID domain.SessionID
    ContentGeneration uint64
    Rect Rect
}
type Measurement struct {
    ViewportGeneration, PresentationGeneration uint64
    ModalID Optional[ModalID]
    ConfirmationPresentable bool
    Panes []PaneRect
    Consoles []ConsoleRect
}
```

PaneRect names one pane/subpane with its actual bounds. Every rendered cell stays
within known viewport bounds; no old-wrapper fallback invents100x30 for a known
small screen. Wide cockpit retains the accepted panes; compact layouts prioritize
the focused pane with explicit access to others. Hidden actionable stash/launch
lists are not an acceptable degradation. Modal layout reserves exact target and
choices, scrolling body text. At sizes where safe confirmation cannot be presented,
render a bounded resize/read-only notice and retain cancel/navigation; State
suppresses approval until a matching adequate measurement arrives.

Render and native terminal measurement use this one geometry algorithm. Host
forwards changed positive console-content dimensions through API with current
SessionID/content/viewport correlation; it never recreates layout constants or
sends a stale rectangle to a replacement session. No-zero-resize native calls.
Modal/focus/mode/viewport changes invalidate prior measurements before approval
or resize effects. Layout does not mutate interaction state to make itself fit.

## Text, console output, graph and styling

Clip/wrap by terminal cells/graphemes, handling combining/wide characters,
invalid UTF-8 and width0/1 safely. Sanitize ordinary metadata, patch and log text;
only renderer-owned styling escapes reach the host. No raw CSI/OSC/control
passthrough from external data. Preserve baseline CR/backspace line behavior
without claiming a complete VT emulator. Runtime retains raw transport bytes;
View's pure interpreter produces safe presentation with gaps/truncation visible.

OutputInput contains SessionID, source/presentation generation, absolute offset,
copied bytes, stream kind and explicit gaps/end. Parser state never guesses over
a missing byte range; report/reset the affected presentation state safely. Host
owns bounded per-session parser/cache lifetime. State may keep copied presentation
snapshots supplied by host, not a second native output store. State emits required
SessionOutput query effects; host executes them, interprets through the pure View
function, then returns correlated presentation input. No backend query originates
inside Render or the interpreter.

Both `--graph` and in-app graph share the same State/controller and pure graph
renderer over accumulated pages. View computes lanes/decorations from supplied
DAG facts without reading Git or embedding lane prefixes into semantic commit
messages. Page continuation retains graph identity/generation; stale pages cannot
splice another graph into the current one. Root/merge/multiple-root cases and
full details remain available.

Theme is immutable instance-local style/palette/profile data, supplied by the
host. View formats supplied lifecycle/activity facts; it never labels a process
working/idle merely from existence, provider or silence. Animation consumes the
supplied frame only. Help is projected from the actual applicable ActionBindings:
q/Tab/control keys cannot advertise navigation when terminal input owns them.
Reserved global mnemonics remain visible/applicable according to State policy.

## Lifecycle, errors and forbidden behavior

The host disposes its one timer and bounded output caches on teardown and joins
its event bridge through Application shutdown. Late ticks/presentation/layout
results with old identity/generation are ignored. Rendering faults do not run a
backend recovery or alter focus. The host can show a bounded rendering diagnostic
and report its primary error to Composition's aggregate shutdown path.

View is forbidden to call Application/Git/GitHub/Runtime/Storage/Discovery, create
effects, assign active worktree/selection/modal authority, infer semantic joins,
read a clock or terminal environment, mutate globals or request native resize.
State is forbidden to rebuild geometry or execute View callbacks. Host is forbidden
to become a product workflow layer while transporting events/measurements.

## Verification and history

V-VIEW-01..03, V-STATE-01..04, V-COMP-02/03 and all relevant E2E cases prove this
boundary. Include1x1/narrow/short/wide bounds, Unicode/control input, exact modal
identity and hidden-approval refusal, actual key help under terminal focus, geometry
agreement, stale modal/session measurement, graph page continuity and immutable
repeat rendering. No substring-only tests substitute for geometry/purity proof.

1.0.0 DRAFT under #55; whole-set independent review precedes freeze. Incompatible
changes require BC-CHANGE/refreeze and affected conformance verification.

Freeze history: 2026-09-06, whole set independently REVIEWED at
7685494e45c0ef44fbccf9b49a589a90a78026d0, then marked FROZEN 1.0.0 by Master.
BCFreeze--001 governs effective authority after final metadata review/CI/PR56 merge.
