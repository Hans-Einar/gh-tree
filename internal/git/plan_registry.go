package git

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
)

type planPhase uint8

const (
	planIssued planPhase = iota + 1
	planExecuting
	planConsumed
	planReleased
)

type registeredPlan struct {
	plan     ports.PreparedGitPlan
	phase    planPhase
	step     uint32
	receipt  api.Optional[ports.GitMutationReceipt]
	released bool
}
type planGroup struct {
	root      ports.PreparedGitPlan
	name      string
	entries   map[api.SourceVersion]*registeredPlan
	steps     [3]*registeredPlan
	deadline  time.Time
	bytes     uint64
	executing int
	released  bool
}

// This registry owns portable handles and completed-step metadata only. Native
// locks, scratch repositories, journals and file handles never enter it.
type planRegistry struct {
	mu              sync.Mutex
	issuer          ports.PlanIssuer
	authority       ports.ApprovalIssuer
	ttl             time.Duration
	maxBytes, bytes uint64
	groups          map[api.OperationID]*planGroup
}

func randomToken() (string, error) {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(token[:]), nil
}

func newPlanRegistry(authority ports.ApprovalIssuer, ttl time.Duration, maxBytes uint64) (*planRegistry, error) {
	// Never reuse the publicly observable SourceVersion issuer as this private
	// plan/receipt issuer lifetime. Only this registry can mint its receipts.
	secret, err := randomToken()
	if err != nil {
		return nil, err
	}
	issuer, err := ports.NewPlanIssuer(secret)
	if err != nil {
		return nil, err
	}
	return &planRegistry{issuer: issuer, authority: authority, ttl: ttl, maxBytes: maxBytes, groups: make(map[api.OperationID]*planGroup)}, nil
}

func (r *planRegistry) enabled() bool { return r.authority != (ports.ApprovalIssuer{}) }

// summaryEncoding is private to this issuer lifetime. Application obtains the
// issued digest via PlanSummaryDigest; it never recreates this encoding. Callers
// must bound native preflight inputs before constructing their immutable summary.
func summaryEncoding(summary api.MutationPlanSummary) ([32]byte, uint64) {
	encoded := fmt.Sprintf("%#v", summary.Data())
	return sha256.Sum256([]byte(encoded)), uint64(len(encoded))
}

func (r *planRegistry) issue(kind api.GitMutationKind, spec ports.PlanSpec) (ports.PreparedGitPlan, error) {
	switch kind {
	case api.CreateMutation:
		return r.issuer.IssueCreate(spec)
	case api.RetargetMutation:
		return r.issuer.IssueRetarget(spec)
	case api.StageMutation:
		return r.issuer.IssueStage(spec)
	case api.CommitMutation:
		return r.issuer.IssueCommit(spec)
	case api.RestoreMutation:
		return r.issuer.IssueRestore(spec)
	case api.StashMutation:
		return r.issuer.IssueStash(spec)
	case api.BranchMutation:
		return r.issuer.IssueBranch(spec)
	case api.PushMutation:
		return r.issuer.IssuePush(spec)
	}
	return nil, diagnostic(api.Invalid, "InvalidPlanKind", "The native plan kind is invalid.")
}

func (r *planRegistry) sweep(now time.Time) {
	for operation, group := range r.groups {
		if !now.Before(group.deadline) {
			group.released = true
			if group.executing == 0 {
				r.bytes -= group.bytes
				delete(r.groups, operation)
			}
		}
	}
}

// Production callers pass time.Now() with its monotonic component; UTC source
// observation timestamps are separate and must not replace the expiry clock.
func (r *planRegistry) issueRoot(summary api.MutationPlanSummary, role ports.PlanRole, now time.Time) (ports.PreparedGitPlan, error) {
	if !summary.Valid() {
		return nil, diagnostic(api.Invalid, "InvalidPlanSummary", "The preparation summary is invalid.")
	}
	if !r.enabled() {
		return nil, diagnostic(api.Unavailable, "MutationApprovalAuthorityRequired", "This adapter has no configured coordinator approval authority.")
	}
	d := summary.Data()
	if d.OriginVersion.Present() {
		return nil, diagnostic(api.Invalid, "UnexpectedPlanOrigin", "An initial preparation cannot carry continuation origin.")
	}
	digest, cost := summaryEncoding(summary)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sweep(now)
	if _, exists := r.groups[d.OperationID]; exists {
		return nil, diagnostic(api.Busy, "OperationPlanExists", "This operation already owns a plan reservation.")
	}
	if len(r.groups) >= 64 || cost > r.maxBytes-r.bytes {
		return nil, diagnostic(api.Busy, "PlanAdmissionFull", "The bounded plan or retained-summary budget is full.")
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	spec := ports.PlanSpec{Operation: d.OperationID, Version: d.PlanVersion, Token: token, Group: token, Step: 0, Role: role, Summary: summary, SummaryDigest: digest}
	plan, err := r.issue(d.Kind, spec)
	if err != nil {
		return nil, err
	}
	entry := &registeredPlan{plan: plan, phase: planIssued}
	group := &planGroup{root: plan, name: token, entries: map[api.SourceVersion]*registeredPlan{d.PlanVersion: entry}, deadline: now.Add(r.ttl), bytes: cost}
	group.steps[0] = entry
	r.groups[d.OperationID] = group
	r.bytes += cost
	return plan, nil
}

func (r *planRegistry) owned(plan ports.PreparedGitPlan) (api.OperationID, api.SourceVersion, error) {
	operation, err := ports.PlanOperation(plan)
	if err != nil {
		return api.OperationID{}, api.SourceVersion{}, diagnostic(api.Invalid, "InvalidPlan", "The opaque plan is invalid.")
	}
	kind, err := ports.PlanKind(plan)
	if err != nil {
		return api.OperationID{}, api.SourceVersion{}, err
	}
	version, err := ports.PlanVersion(plan)
	if err != nil {
		return api.OperationID{}, api.SourceVersion{}, err
	}
	if err = r.issuer.ValidateBinding(plan, operation, kind, version); err != nil {
		return api.OperationID{}, api.SourceVersion{}, diagnostic(api.Invalid, "ForeignPlan", "The plan was not issued by this adapter lifetime.")
	}
	return operation, version, nil
}

func (r *planRegistry) find(plan ports.PreparedGitPlan) (*planGroup, *registeredPlan, error) {
	operation, version, err := r.owned(plan)
	if err != nil {
		return nil, nil, err
	}
	group := r.groups[operation]
	if group == nil {
		return nil, nil, diagnostic(api.ConfirmationStale, "UnknownOrExpiredPlan", "The plan reservation is no longer live.")
	}
	entry := group.entries[version]
	if entry == nil || !ports.PlanEqual(entry.plan, plan) {
		return nil, nil, diagnostic(api.Invalid, "UnknownPlan", "The exact plan was not registered in this operation.")
	}
	return group, entry, nil
}

func sameFacetVersions(a, b []api.FacetVersion) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Data() != b[i].Data() {
			return false
		}
	}
	return true
}

// Operation-specific preparation must independently validate original intent,
// exact targets/paths and allowed own-step facet advancement before calling this.
// This method enforces the bounded group, actual stored completion and step order.
func (r *planRegistry) issueChild(root ports.PreparedGitPlan, summary api.MutationPlanSummary, prior api.Optional[ports.GitMutationReceipt], now time.Time) (ports.PreparedGitPlan, error) {
	if !summary.Valid() {
		return nil, diagnostic(api.Invalid, "InvalidPlanSummary", "The continuation summary is invalid.")
	}
	digest, cost := summaryEncoding(summary)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sweep(now)
	group, _, err := r.find(root)
	if err != nil {
		return nil, err
	}
	role, err := ports.PlanExecutionRole(root)
	if err != nil || role != ports.SequenceRoot || !ports.PlanEqual(group.root, root) || group.released {
		return nil, diagnostic(api.ConfirmationStale, "InvalidSequenceRoot", "The sequence root is not live.")
	}
	step := uint32(1)
	if receipt, p := prior.Value(); p {
		step = 2
		first := group.steps[1]
		if first == nil || first.phase != planConsumed {
			return nil, diagnostic(api.Invalid, "MissingCompletedPredecessor", "Continuation requires its completed first step.")
		}
		stored, known := first.receipt.Value()
		if !known || !receipt.Valid() || !receipt.Matches(first.plan) || !stored.Verified() || stored.CancellationRequested() || !receipt.Verified() || receipt.CancellationRequested() || !sameFacetVersions(stored.Versions(), receipt.Versions()) {
			return nil, diagnostic(api.ConfirmationStale, "InvalidPredecessorReceipt", "The predecessor does not match a verified uncanceled native completion.")
		}
	}
	if group.executing != 0 || group.steps[step] != nil {
		return nil, diagnostic(api.Busy, "SequenceStepAlreadyAdmitted", "The sequence step is executing or already issued.")
	}
	d := summary.Data()
	rootSummary, _ := ports.PlanSummary(root)
	if d.OperationID != rootSummary.Data().OperationID || d.Repository != rootSummary.Data().Repository || d.Worktree != rootSummary.Data().Worktree {
		return nil, diagnostic(api.Invalid, "ContinuationSubjectChanged", "A continuation cannot change its original operation or subject.")
	}
	if _, exists := group.entries[d.PlanVersion]; exists {
		return nil, diagnostic(api.Invalid, "PlanVersionReused", "Each issued step requires its own plan version.")
	}
	if err := continuationSummary(rootSummary.Data(), d, step); err != nil {
		return nil, err
	}
	if cost > r.maxBytes-r.bytes {
		return nil, diagnostic(api.Busy, "PlanAdmissionFull", "The retained-summary budget is full.")
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	// Group need only be stable within this private reservation. Recover its
	// issuer-assigned value from private registry metadata, never from caller text.
	groupName := group.name
	spec := ports.PlanSpec{Operation: d.OperationID, Version: d.PlanVersion, Token: token, Group: groupName, Step: step, Role: ports.Executable, Summary: summary, SummaryDigest: digest, Origin: api.Some(root)}
	plan, err := r.issue(d.Kind, spec)
	if err != nil {
		return nil, err
	}
	entry := &registeredPlan{plan: plan, phase: planIssued, step: step}
	group.entries[d.PlanVersion] = entry
	group.steps[step] = entry
	group.bytes += cost
	r.bytes += cost
	return plan, nil
}

func continuationSummary(root, child api.MutationPlanSummaryData, step uint32) error {
	bad := func() error {
		return diagnostic(api.Invalid, "ContinuationIntentChanged", "The child summary differs from the original approved sequence intent.")
	}
	switch root.Kind {
	case api.CommitMutation:
		if step == 2 && (child.Message != root.Message || child.CommitIndexPolicy != api.Some(api.ExistingIndex)) {
			return bad()
		}
	case api.RetargetMutation:
		if step == 2 && (child.Target != root.Target || child.RetargetMode != root.RetargetMode) {
			return bad()
		}
	case api.StashMutation:
		intent, p := root.StashIntent.Value()
		if !p {
			return bad()
		}
		pop, ok := intent.(api.PopStashIntent)
		if !ok {
			return bad()
		}
		next, p := child.StashIntent.Value()
		if !p {
			return bad()
		}
		if step == 1 {
			apply, ok := next.(api.ApplyStashIntent)
			if !ok || apply.Data().Worktree != pop.Data().Worktree || apply.Data().Stash != pop.Data().Stash || apply.Data().Occurrence != pop.Data().Occurrence {
				return bad()
			}
		} else {
			drop, ok := next.(api.DropStashIntent)
			if !ok || drop.Data().Stash != pop.Data().Stash || drop.Data().Occurrence != pop.Data().Occurrence {
				return bad()
			}
		}
	}
	return nil
}

func (r *planRegistry) begin(plan ports.PreparedGitPlan, approval ports.ExecutionApproval, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sweep(now)
	group, entry, err := r.find(plan)
	if err != nil {
		return err
	}
	role, err := ports.PlanExecutionRole(plan)
	if err != nil || role != ports.Executable {
		return diagnostic(api.Invalid, "SequenceRootNotExecutable", "A sequence root cannot execute as a native step.")
	}
	if group.released || entry.released || entry.phase != planIssued {
		return diagnostic(api.ConfirmationStale, "PlanNotIssued", "The plan is executing, consumed, released or expired.")
	}
	if !r.authority.Issued(approval) || !approval.ValidFor(plan) {
		return diagnostic(api.ConfirmationStale, "InvalidExecutionApproval", "The coordinator issuer, original plan or approved choice does not match.")
	}
	if group.executing != 0 {
		return diagnostic(api.Busy, "SequenceExecuting", "Another step of this operation is executing.")
	}
	entry.phase = planExecuting
	group.executing++
	return nil
}

// finish is called only after the native owner has reached its cleanup/result
// barrier. The registry cannot prove a native effect from an author-supplied bool.
func (r *planRegistry) finish(plan ports.PreparedGitPlan, versions []api.FacetVersion, verified, canceled bool) (ports.GitMutationReceipt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	group, entry, err := r.find(plan)
	if err != nil {
		return ports.GitMutationReceipt{}, err
	}
	if entry.phase != planExecuting {
		return ports.GitMutationReceipt{}, diagnostic(api.Invalid, "PlanNotExecuting", "Only the admitted native execution can complete this plan.")
	}
	defer func() {
		entry.phase = planConsumed
		group.executing--
		if group.released {
			operation, _, _ := r.owned(plan)
			r.bytes -= group.bytes
			delete(r.groups, operation)
		}
	}()
	if canceled {
		verified = false
	}
	token, err := randomToken()
	if err != nil {
		return ports.GitMutationReceipt{}, err
	}
	receipt, err := r.issuer.IssueReceipt(plan, token, versions, verified, canceled)
	if err != nil {
		return ports.GitMutationReceipt{}, err
	}
	entry.receipt = api.Some(receipt)
	return receipt, nil
}

func (r *planRegistry) release(plan ports.PreparedGitPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	operation, version, err := r.owned(plan)
	if err != nil {
		return err
	}
	group := r.groups[operation]
	// Repeated release of an owned, already-disposed handle is a no-op. This
	// grants no execution/continuation and requires no unbounded tombstone log.
	if group == nil {
		return nil
	}
	entry := group.entries[version]
	if entry == nil || !ports.PlanEqual(entry.plan, plan) {
		return diagnostic(api.Invalid, "UnknownPlan", "The exact plan is not part of this operation.")
	}
	entry.released = true
	if entry.phase == planIssued {
		entry.phase = planReleased
	}
	if ports.PlanEqual(group.root, plan) {
		group.released = true
		if group.executing == 0 {
			r.bytes -= group.bytes
			delete(r.groups, operation)
		}
	}
	return nil
}

func (a *Adapter) ReleasePlan(plan ports.PreparedGitPlan) error { return a.plans.release(plan) }
