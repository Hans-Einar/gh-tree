package ports

import "github.com/Hans-Einar/gh-tree/internal/application/api"

type PreparedGitPlan interface {
	preparedGitPlan()
	Valid() bool
}
type CreatePlan struct{ identity planIdentity }

func (v CreatePlan) preparedGitPlan() {}
func (v CreatePlan) Valid() bool      { return v.identity.valid() && v.identity.kind == api.CreateMutation }
func (i PlanIssuer) IssueCreate(spec PlanSpec) (CreatePlan, error) {
	p, e := i.issue(api.CreateMutation, spec)
	if e != nil {
		return CreatePlan{}, e
	}
	return CreatePlan{p}, nil
}

type RetargetPlan struct{ identity planIdentity }

func (v RetargetPlan) preparedGitPlan() {}
func (v RetargetPlan) Valid() bool {
	return v.identity.valid() && v.identity.kind == api.RetargetMutation
}
func (i PlanIssuer) IssueRetarget(spec PlanSpec) (RetargetPlan, error) {
	p, e := i.issue(api.RetargetMutation, spec)
	if e != nil {
		return RetargetPlan{}, e
	}
	return RetargetPlan{p}, nil
}

type StagePlan struct{ identity planIdentity }

func (v StagePlan) preparedGitPlan() {}
func (v StagePlan) Valid() bool      { return v.identity.valid() && v.identity.kind == api.StageMutation }
func (i PlanIssuer) IssueStage(spec PlanSpec) (StagePlan, error) {
	p, e := i.issue(api.StageMutation, spec)
	if e != nil {
		return StagePlan{}, e
	}
	return StagePlan{p}, nil
}

type CommitPlan struct{ identity planIdentity }

func (v CommitPlan) preparedGitPlan() {}
func (v CommitPlan) Valid() bool      { return v.identity.valid() && v.identity.kind == api.CommitMutation }
func (i PlanIssuer) IssueCommit(spec PlanSpec) (CommitPlan, error) {
	p, e := i.issue(api.CommitMutation, spec)
	if e != nil {
		return CommitPlan{}, e
	}
	return CommitPlan{p}, nil
}

type RestorePlan struct{ identity planIdentity }

func (v RestorePlan) preparedGitPlan() {}
func (v RestorePlan) Valid() bool {
	return v.identity.valid() && v.identity.kind == api.RestoreMutation
}
func (i PlanIssuer) IssueRestore(spec PlanSpec) (RestorePlan, error) {
	p, e := i.issue(api.RestoreMutation, spec)
	if e != nil {
		return RestorePlan{}, e
	}
	return RestorePlan{p}, nil
}

type StashPlan struct{ identity planIdentity }

func (v StashPlan) preparedGitPlan() {}
func (v StashPlan) Valid() bool      { return v.identity.valid() && v.identity.kind == api.StashMutation }
func (i PlanIssuer) IssueStash(spec PlanSpec) (StashPlan, error) {
	p, e := i.issue(api.StashMutation, spec)
	if e != nil {
		return StashPlan{}, e
	}
	return StashPlan{p}, nil
}

type BranchPlan struct{ identity planIdentity }

func (v BranchPlan) preparedGitPlan() {}
func (v BranchPlan) Valid() bool      { return v.identity.valid() && v.identity.kind == api.BranchMutation }
func (i PlanIssuer) IssueBranch(spec PlanSpec) (BranchPlan, error) {
	p, e := i.issue(api.BranchMutation, spec)
	if e != nil {
		return BranchPlan{}, e
	}
	return BranchPlan{p}, nil
}

type PushPlan struct{ identity planIdentity }

func (v PushPlan) preparedGitPlan() {}
func (v PushPlan) Valid() bool      { return v.identity.valid() && v.identity.kind == api.PushMutation }
func (i PlanIssuer) IssuePush(spec PlanSpec) (PushPlan, error) {
	p, e := i.issue(api.PushMutation, spec)
	if e != nil {
		return PushPlan{}, e
	}
	return PushPlan{p}, nil
}
func identity(v PreparedGitPlan) (planIdentity, bool) {
	switch p := v.(type) {
	case CreatePlan:
		return p.identity, p.Valid()
	case RetargetPlan:
		return p.identity, p.Valid()
	case StagePlan:
		return p.identity, p.Valid()
	case CommitPlan:
		return p.identity, p.Valid()
	case RestorePlan:
		return p.identity, p.Valid()
	case StashPlan:
		return p.identity, p.Valid()
	case BranchPlan:
		return p.identity, p.Valid()
	case PushPlan:
		return p.identity, p.Valid()
	default:
		return planIdentity{}, false
	}
}
