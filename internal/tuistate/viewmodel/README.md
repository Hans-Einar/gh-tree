# State-owned immutable presentation values

Issue #60 / M2 implements the presentation leaf in FROZEN 1.0.0
BC--TUIState--TUIView, BC--TUIState--Application and BoundaryTypes--001.
Its only product import is the accepted `internal/domain` foundation. There is
no API/ports/parent-State/View dependency, operation/confirmation capability,
callback, backend, effect, reducer, layout algorithm, renderer or output parser.

State supplies reconciled semantic facts and actual applicable actions. The View
formats those facts. These constructors validate representation and correlation;
they never associate repositories, resolve a ref, construct an invocation,
select an action, allocate a session or infer business truth from names/OIDs.

## Publication and validity

Each record family has private storage, `NewX(XSpec) (X, error)`, `Valid`,
`Fields` and `Clone`. A Spec is an editable constructor/accessor copy. Construction,
access and cloning recursively copy every slice and byte payload, including
optional nested records and slices inside nested records. Mutable maps are not
part of the public model. Domain values are already immutable comparable values.
No getter exposes owned mutable backing. Failed constructors return invalid zero.

All identities, tags, rows, bodies, pane records, snapshots and measurements have
invalid zero values. Intentional exceptions are `None[T]`, `UnknownViewport`,
zero `Scroll`, empty `Rect` and empty absolute `ByteRange`. Known viewport
dimensions require a positive ViewportGeneration; zero width/height are genuine
bounds. Unknown startup dimensions have no dimensions and generation zero.

`Optional[T]` describes presence only. It is not a generic deep-copy utility.
Containing record families validate a present T and clone mutable nested content.
All option payloads admitted by this leaf are immutable values, records or copied
scalar facts. A present invalid Domain ID/record is rejected; absence is explicit.

Ordinary text facts (commit subject/message, paths, patch bytes, labels and
diagnostic display text) remain verbatim, including invalid UTF-8/control bytes.
State must supply presentation-safe disclosure content without credentials or
private protocol detail. View must sanitize ordinary text before rendering and
must generate its own styling escapes. This leaf does not silently rewrite facts.
`ConsoleLine` differs: it is the later pure interpreter's output and admits only
valid UTF-8 without C0/DEL/C1 controls; lines contain no ANSI styling or newline.

## Identity, sources and selection

`ElementID` is a closed comparable union constructed by NewNamespaceElement,
NewRepositoryElement, NewPRElement, NewBranchElement, NewRevisionElement,
NewWorktreeElement, NewStashElement, NewLaunchElement and NewSessionElement.
Domain repository scope and full 40/64-digit OID values remain accessible.
Launch identity retains an optional exact case-sensitive saved alias; two aliases
of one LaunchPointID are distinct. Namespace/key elements are local presentation
keys for navigator folders, file paths and configured destinations; callers give
an explicit namespace and, where relevant, repository scope. Keys are not backend
capabilities or mutable-reference substitutes. Labels/positions do not enter IDs.

`SourceStatus` supplies independent availability, completeness, freshness,
optional SourceGeneration and notices. SourceGeneration is a local correlation
counter within the explicitly supplied model/source identity, not API SourceVersion,
observation time, an ordering across sources or mutation authority. Complete/More/
Partial/Unknown/NotAvailable remain distinct. Ahead/behind are optional and allowed
only for a resolved upstream. No missing count becomes zero.

`ListState` retains selected identity, optional nonnegative cursor, vertical and
horizontal scroll, filter and filter-editing fact. A selected identity need not
be in a partial page: State owns disappearance reconciliation and fallback.
No constructor clamps scrolling or changes selection to fit data or a viewport.

## Concrete family inventory

| Family | Values and retained details |
|---|---|
| Navigator | NamespaceRow with semantic parent/depth/folder expansion; NavigatorModel carries namespace rows, typed BranchRows/PRRows, folder path and ListState. |
| Branch/PR | BranchEndpoint, Upstream, PRAnnotation, WorktreeAnnotation, BranchRow, PRRow; exact endpoints, fork/base scope, relationship evidence, current/primary/active flags, metadata/unknown badges, full title/body. |
| Worktree | WorktreeRow, FileChange, WorktreesModel, ActiveModel; optional observed Attached/Detached/Unborn Head, locator/availability/lock/prunable facts, staged/worktree change status, rename source and counts. |
| History | CommitRow, BranchModel, HistoryModel; exact selected source/target, complete parent IDs, verbatim subject/message, author/email/committer and supplied source timestamps, independent list/detail/message scroll. |
| Graph | GraphModel, GraphRef, GraphAnnotation; accumulated commits/parents, roots/boundary parents, exact branch/tag/HEAD/cached ref facts, supplied PR/worktree annotations, source generation/completeness, selection and scroll. No lane prefixes enter commit messages. |
| Diff | Closed Comparison, FileChange, Patch, DiffModel; exact endpoints, rename/binary/full copied bounded patch payload, original-size/truncation notices, file selection/patch scroll, supplied stage/unstage applicability. |
| Stash | StashRow, StashesModel, StashComparisonDetail; stable stash OID, exact commit-parent and actual tree OIDs, origin/worktree/managed observations, positional label only for display, message/detail. Duplicate stash-object observations remain visible. |
| Launch | LaunchMember, LaunchRow, LaunchModel; exact ID/alias, provider/project/source labels, default/availability and ordered selected members. No executable/argv construction or provider interpretation. |
| Console | ConsoleSummary, ConsoleCapabilities, ConsoleModel, ConsolesModel; SessionID, lifecycle/cleanup/exit/activity facts, safe supplied display summary, actual input-focus fact, copied interpreted output, selection/scroll/notices. |
| Detail/status | TextDetail, DetailField, TargetDetail, RowMeta, SourceStatus, Badge, StatusNotice; exact subjects, full text/identity fields and supplied severity. |

`PaneModel` is closed over Navigator/Worktrees/Active/Branch/Launch/Stashes/Console/
History/Graph/Diff bodies. Use NewNavigatorPane etc. and the matching typed accessor.
Each has a PaneHeader with availability/completeness, ContentGeneration, sources
and notices. PanePath and FocusPath use closed pane/part IDs, never titles.
Snapshot validates mode/pane/focus/modal compatibility, unique panes/selections,
projected selection consistency and actual terminal input ownership. A modal
owns current focus; latent underlying pane data remains available.

Comparison constructors cover CommitParent (explicit absent parent for a root),
CommitPair, IndexToWorktree, HeadToIndex (including unborn Head), PullRequest and
Stash views. PR comparison retains original scoped target/base and all supplied
resolved local base/head/merge-base endpoints. IndexToWorktree accurately names
the working-tree diff; only that variant may advertise stage, and only HeadToIndex
may advertise unstage. StashComparisonDetail covers base-to-worktree, base-to-index,
index-to-worktree, untracked and explicit zero-based parent index. It preserves the
observed two/three parent OIDs and actual tree endpoints; an absent untracked parent
has explicitly absent endpoints. Tree OIDs never become fabricated Revisions.
The values check supplied exact mappings, never obtain them.

Timestamp retains a supplied instant plus its explicitly supplied original offset
in seconds. Commit author/committer, PR update and stash creation times use these
optional values. Its private instant is UTC and exposes the original offset
separately, preserving source semantics without retaining a lazy local-zone object.
NewTimestamp reads no clock or source timezone; callers provide the observed offset.
These source properties are not UTC observation intervals or freshness signals.

## Modal and action presentation

ModalID and OwnerKey are opaque local presentation keys. State privately binds
them to intent/operation metadata. Neither is an Application ConfirmationID.
Modal has one purpose, one corresponding closed ModalBody, exact TargetDetail,
allowed semantic choices and body Scroll. All purposes retain enabled cancel.

Confirmation purposes cover deploy, retarget, tracked restore, stash create/apply/
pop/drop, push, quit and saved-alias replacement. Required target/worktree/path/
stash/alias identities are validated. Form purposes cover create worktree, new
branch, commit, PR and saved launch. Text/boolean FormField variants validate
closed FieldIDs, kind, rune cursor and multiline policy. Required fields retain
path/branch, stage-all commit intent, PR base/title/body/draft and saved alias/default
with executable override. Additional booleans expose retained checkout/detach and
maintainer-modification choices where relevant. Editing can contain incomplete
text; product validation remains Application work.

Chooser bodies carry typed worktrees, configured destinations or launch models.
Detail bodies carry exact stash diff or full text detail. Purpose/body and choice
combinations are validated; stale modal keys and incompatible focused fields are
rejected by Snapshot. Approval is still suppressed by State until an adequate
matching View measurement exists.

ActionBinding carries a closed semantic ActionID, copied actual key chord,
global/pane/modal/session scope, supplied applicability/enabled facts and label.
There is no execution payload or second command protocol. State owns bindings and
input capture; View derives help from these facts, including terminal key ownership.

## Generations, output and measurement

PresentationGeneration, ViewportGeneration, ContentGeneration and SourceGeneration
are distinct Go types. Positive generations are supplied, never incremented here.
Snapshot receives its animation frame and version text; no clock/environment lookup
or timer exists. Console interpreter generations describe its content independently
of the outer snapshot presentation generation, which changes with focus/modal/etc.

OutputInput carries SessionID, the three content/source/presentation correlations,
absolute byte offset, copied contiguous raw bytes, stream, ordered explicit preceding
gaps and end. Gaps are nonempty, nonoverlapping absolute ranges ending no later than
that payload's offset. A later payload after a hole is another input, so the parser
never needs to guess across missing bytes. Overflowing offsets are rejected.
ConsolePresentation carries the same correlation, retained absolute range, copied
safe lines and their source ranges, explicit gaps/truncation/end and notices.
ConsoleModel rejects presentation from another session or generation. Runtime
retains transport; host owns bounded parser/cache lifetime; View implements the
future CR/backspace interpreter. ActivityUnknown/OutputObserved are supplied facts,
never an invented working/idle classification from process existence.

Rect, PaneRect, ConsoleRect and Measurement validate nonnegative overflow-safe
actual bounds, unique geometry entries and known viewport containment. Empty
rectangles are permitted presentation facts; `Rect.Positive()` distinguishes them
from possible native resize dimensions. Measurement retains the input Viewport,
presentation generation, optional local modal key and presentability result.
`Matches(snapshot)` checks viewport/presentation/modal identity plus console
SessionID/ContentGeneration and pane membership. It performs no layout or resize.
Positive dimensions alone do not prove adequate confirmation geometry; the later
View Layout and State/host gates must prove that and retain bounded cancel/navigation.

## Verification and remaining gates

External-consumer tests construct all ten panes and all twenty modal purposes,
inspect full identities/details, reject invalid options/scopes/kinds/generations,
and test copied raw output, nested fields and actual distinct backing arrays.
Package-local tests deliberately forge contradictory private variants. The M1
checker validates pure imports/public type graphs and all twelve target selections.
Unit/race/full-suite results and frozen source CI are recorded in the bounded
M2-Viewmodel--001 report. These values are partial V-VIEW-03/M2 evidence only.
Layout/Render/InterpretOutput, State reducer behavior, actual geometry/text visual
QA, native terminals, integrated Slices, M2 completion and release remain later gates.
