package api

type DirectoryPlatform uint8

const (
	DirectoryWindows DirectoryPlatform = 1
	DirectoryUnix    DirectoryPlatform = 2
)

func (v DirectoryPlatform) Valid() bool { return v >= DirectoryWindows && v <= DirectoryUnix }

type Completeness uint8

const (
	Complete                Completeness = 1
	More                    Completeness = 2
	Partial                 Completeness = 3
	Unknown                 Completeness = 4
	UnavailableCompleteness Completeness = 5
)

func (v Completeness) Valid() bool { return v >= Complete && v <= UnavailableCompleteness }

type ErrorCode uint8

const (
	Invalid           ErrorCode = 1
	Unavailable       ErrorCode = 2
	NotFound          ErrorCode = 3
	Busy              ErrorCode = 4
	Canceled          ErrorCode = 5
	Superseded        ErrorCode = 6
	StaleObservation  ErrorCode = 7
	ConfirmationStale ErrorCode = 8
	Conflict          ErrorCode = 9
	Permission        ErrorCode = 10
	Unsupported       ErrorCode = 11
	IOFailure         ErrorCode = 12
	ProcessFailure    ErrorCode = 13
	CleanupIncomplete ErrorCode = 14
	Indeterminate     ErrorCode = 15
)

func (v ErrorCode) Valid() bool { return v >= Invalid && v <= Indeterminate }

type EffectFacet uint8

const (
	ObjectAcquisition  EffectFacet = 1
	Recovery           EffectFacet = 2
	WorktreeBytes      EffectFacet = 3
	Index              EffectFacet = 4
	LocalRefsHead      EffectFacet = 5
	LocalConfiguration EffectFacet = 6
	RemoteRefsPR       EffectFacet = 7
	Storage            EffectFacet = 8
	RuntimeResources   EffectFacet = 9
)

func (v EffectFacet) Valid() bool { return v >= ObjectAcquisition && v <= RuntimeResources }

type EffectState uint8

const (
	NotStarted             EffectState = 1
	VerifiedNoTargetChange EffectState = 2
	AppliedVerified        EffectState = 3
	EffectPartial          EffectState = 4
	EffectIndeterminate    EffectState = 5
)

func (v EffectState) Valid() bool { return v >= NotStarted && v <= EffectIndeterminate }

type ResponsibleLayer uint8

const (
	LayerGit         ResponsibleLayer = 1
	LayerGitHub      ResponsibleLayer = 2
	LayerPersistence ResponsibleLayer = 3
	LayerRuntime     ResponsibleLayer = 4
	LayerApplication ResponsibleLayer = 5
	LayerDiscovery   ResponsibleLayer = 6
)

func (v ResponsibleLayer) Valid() bool { return v >= LayerGit && v <= LayerDiscovery }

type RecoveryKind uint8

const (
	RecoveryObjects          RecoveryKind = 1
	RecoveryJournal          RecoveryKind = 2
	RecoveryOriginal         RecoveryKind = 3
	RecoveryPayload          RecoveryKind = 4
	RecoveryManifest         RecoveryKind = 5
	RecoveryRawOriginal      RecoveryKind = 6
	RecoveryRetainedOriginal RecoveryKind = 7
	RecoveryRetainedPayload  RecoveryKind = 8
)

func (v RecoveryKind) Valid() bool { return v >= RecoveryObjects && v <= RecoveryRetainedPayload }

type Choice uint8

const (
	Proceed         Choice = 1
	Cancel          Choice = 2
	StashThenDeploy Choice = 3
)

func (v Choice) Valid() bool { return v >= Proceed && v <= StashThenDeploy }

type StorageFamily uint8

const (
	UserConfig  StorageFamily = 1
	Preferences StorageFamily = 2
	RunConfig   StorageFamily = 3
)

func (v StorageFamily) Valid() bool { return v >= UserConfig && v <= RunConfig }

type StorageLoadState uint8

const (
	LoadAbsent         StorageLoadState = 1
	ValidLegacy        StorageLoadState = 2
	ValidCurrent       StorageLoadState = 3
	Corrupt            StorageLoadState = 4
	UnsupportedVersion StorageLoadState = 5
	LoadUnavailable    StorageLoadState = 6
	UnsupportedProfile StorageLoadState = 7
)

func (v StorageLoadState) Valid() bool { return v >= LoadAbsent && v <= UnsupportedProfile }

type StorageCommitOutcome uint8

const (
	NotCommitted                 StorageCommitOutcome = 1
	Committed                    StorageCommitOutcome = 2
	CommittedDurabilityUncertain StorageCommitOutcome = 3
	StorageIndeterminate         StorageCommitOutcome = 4
)

func (v StorageCommitOutcome) Valid() bool { return v >= NotCommitted && v <= StorageIndeterminate }

type StorageDurability uint8

const (
	DurabilityNotApplicable       StorageDurability = 1
	SupportedCrashBarrierComplete StorageDurability = 2
	DurabilityUncertain           StorageDurability = 3
)

func (v StorageDurability) Valid() bool {
	return v >= DurabilityNotApplicable && v <= DurabilityUncertain
}

type StorageRecoveryKind uint8

const (
	Manifest         StorageRecoveryKind = 1
	RawOriginal      StorageRecoveryKind = 2
	RetainedOriginal StorageRecoveryKind = 3
	RetainedPayload  StorageRecoveryKind = 4
)

func (v StorageRecoveryKind) Valid() bool { return v >= Manifest && v <= RetainedPayload }

type TerminalMode uint8

const (
	Pipes    TerminalMode = 1
	Terminal TerminalMode = 2
)

func (v TerminalMode) Valid() bool { return v >= Pipes && v <= Terminal }

type SessionPhase uint8

const (
	Starting      SessionPhase = 1
	Running       SessionPhase = 2
	Stopping      SessionPhase = 3
	Cleaned       SessionPhase = 4
	CleanupFailed SessionPhase = 5
)

func (v SessionPhase) Valid() bool { return v >= Starting && v <= CleanupFailed }

type RuntimeCleanupStage uint8

const (
	Acquisition        RuntimeCleanupStage = 1
	ProcessContainment RuntimeCleanupStage = 2
	CwdAcquisition     RuntimeCleanupStage = 3
	UserProcessWait    RuntimeCleanupStage = 4
	Descendants        RuntimeCleanupStage = 5
	TerminalCleanup    RuntimeCleanupStage = 6
	InputCleanup       RuntimeCleanupStage = 7
	OutputCleanup      RuntimeCleanupStage = 8
	ControlCleanup     RuntimeCleanupStage = 9
	SupervisorOrBroker RuntimeCleanupStage = 10
	OuterContainment   RuntimeCleanupStage = 11
	HelperExtraction   RuntimeCleanupStage = 12
	EventTransfer      RuntimeCleanupStage = 13
)

func (v RuntimeCleanupStage) Valid() bool { return v >= Acquisition && v <= EventTransfer }

type RuntimeEventKind uint8

const (
	StateChanged    RuntimeEventKind = 1
	OutputAvailable RuntimeEventKind = 2
	RuntimeCleaned  RuntimeEventKind = 3
)

func (v RuntimeEventKind) Valid() bool { return v >= StateChanged && v <= RuntimeCleaned }

type OutputStream uint8

const (
	Stdout         OutputStream = 1
	Stderr         OutputStream = 2
	TerminalOutput OutputStream = 3
)

func (v OutputStream) Valid() bool { return v >= Stdout && v <= TerminalOutput }

type ExitCause uint8

const (
	NaturalExit     ExitCause = 1
	RequestedExit   ExitCause = 2
	FailedStartExit ExitCause = 3
)

func (v ExitCause) Valid() bool { return v >= NaturalExit && v <= FailedStartExit }

type CleanupState uint8

const (
	CleanupPending     CleanupState = 1
	CleanupComplete    CleanupState = 2
	CleanupFailedState CleanupState = 3
)

func (v CleanupState) Valid() bool { return v >= CleanupPending && v <= CleanupFailedState }

type RefBackend uint8

const (
	FilesRefs    RefBackend = 1
	ReftableRefs RefBackend = 2
	OtherRefs    RefBackend = 3
)

func (v RefBackend) Valid() bool { return v >= FilesRefs && v <= OtherRefs }

type FileKind uint8

const (
	RegularFile   FileKind = 1
	SymlinkFile   FileKind = 2
	DirectoryFile FileKind = 3
	OtherFile     FileKind = 4
)

func (v FileKind) Valid() bool { return v >= RegularFile && v <= OtherFile }

type IndexFlag uint8

const (
	AssumeUnchanged   IndexFlag = 1
	SkipWorktree      IndexFlag = 2
	IntentToAdd       IndexFlag = 3
	ExtendedIndexFlag IndexFlag = 4
)

func (v IndexFlag) Valid() bool { return v >= AssumeUnchanged && v <= ExtendedIndexFlag }

type ChangeKind uint8

const (
	Added       ChangeKind = 1
	Modified    ChangeKind = 2
	Deleted     ChangeKind = 3
	Renamed     ChangeKind = 4
	Copied      ChangeKind = 5
	TypeChanged ChangeKind = 6
	Untracked   ChangeKind = 7
	Unmerged    ChangeKind = 8
)

func (v ChangeKind) Valid() bool { return v >= Added && v <= Unmerged }

// ChangeCause identifies the comparison or native status cause of one change.
// Current index and filesystem facts alone cannot establish this distinction.
type ChangeCause uint8

const (
	IndexChangeCause     ChangeCause = 1
	WorktreeChangeCause  ChangeCause = 2
	UntrackedChangeCause ChangeCause = 3
	ConflictChangeCause  ChangeCause = 4
)

func (v ChangeCause) Valid() bool { return v >= IndexChangeCause && v <= ConflictChangeCause }

type FetchFreshnessKind uint8

const (
	FreshnessUnknown FetchFreshnessKind = 1
	Cached           FetchFreshnessKind = 2
	Refreshed        FetchFreshnessKind = 3
)

func (v FetchFreshnessKind) Valid() bool { return v >= FreshnessUnknown && v <= Refreshed }

type RefKind uint8

const (
	LocalBranchKind  RefKind = 1
	LocalTagKind     RefKind = 2
	CachedRemoteKind RefKind = 3
)

func (v RefKind) Valid() bool { return v >= LocalBranchKind && v <= CachedRemoteKind }

type CommitTraversal uint8

const (
	FirstParent CommitTraversal = 1
	AllParents  CommitTraversal = 2
)

func (v CommitTraversal) Valid() bool { return v >= FirstParent && v <= AllParents }

type RetargetPurpose uint8

const (
	RetargetPurposeRetarget RetargetPurpose = 1
	RetargetPurposeDeploy   RetargetPurpose = 2
	RetargetPurposePull     RetargetPurpose = 3
)

func (v RetargetPurpose) Valid() bool {
	return v >= RetargetPurposeRetarget && v <= RetargetPurposePull
}

type DirtyPolicy uint8

const (
	RefuseDirty          DirtyPolicy = 1
	OfferStashThenDeploy DirtyPolicy = 2
)

func (v DirtyPolicy) Valid() bool { return v >= RefuseDirty && v <= OfferStashThenDeploy }

type StageAction uint8

const (
	Stage   StageAction = 1
	Unstage StageAction = 2
)

func (v StageAction) Valid() bool { return v >= Stage && v <= Unstage }

type CommitIndexPolicy uint8

const (
	ExistingIndex    CommitIndexPolicy = 1
	ObservedStageAll CommitIndexPolicy = 2
)

func (v CommitIndexPolicy) Valid() bool { return v >= ExistingIndex && v <= ObservedStageAll }

type GitMutationKind uint8

const (
	CreateMutation   GitMutationKind = 1
	RetargetMutation GitMutationKind = 2
	StageMutation    GitMutationKind = 3
	CommitMutation   GitMutationKind = 4
	RestoreMutation  GitMutationKind = 5
	StashMutation    GitMutationKind = 6
	BranchMutation   GitMutationKind = 7
	PushMutation     GitMutationKind = 8
)

func (v GitMutationKind) Valid() bool { return v >= CreateMutation && v <= PushMutation }

type GitStepKind uint8

const (
	ObjectTransferStep        GitStepKind = 1
	RecoveryPinStep           GitStepKind = 2
	WorktreeRegistrationStep  GitStepKind = 3
	FilePublicationStep       GitStepKind = 4
	IndexPublicationStep      GitStepKind = 5
	RefPublicationStep        GitStepKind = 6
	HookStep                  GitStepKind = 7
	StashStoreStep            GitStepKind = 8
	StashCleanupStep          GitStepKind = 9
	StashApplyStep            GitStepKind = 10
	StashDropStep             GitStepKind = 11
	BranchCreateStep          GitStepKind = 12
	RemotePushStep            GitStepKind = 13
	UpstreamConfigurationStep GitStepKind = 14
	CommitCandidateStep       GitStepKind = 15
)

func (v GitStepKind) Valid() bool { return v >= ObjectTransferStep && v <= CommitCandidateStep }

type FetchDestinationPolicy uint8

const (
	OperationPrivateRoot  FetchDestinationPolicy = 1
	ConfiguredTrackingRef FetchDestinationPolicy = 2
)

func (v FetchDestinationPolicy) Valid() bool {
	return v >= OperationPrivateRoot && v <= ConfiguredTrackingRef
}

type EndpointUnavailableReason uint8

const (
	EndpointDeleted      EndpointUnavailableReason = 1
	EndpointInaccessible EndpointUnavailableReason = 2
	EndpointMissingField EndpointUnavailableReason = 3
	EndpointUnresolved   EndpointUnavailableReason = 4
)

func (v EndpointUnavailableReason) Valid() bool {
	return v >= EndpointDeleted && v <= EndpointUnresolved
}

type PullRequestState uint8

const (
	PROpen    PullRequestState = 1
	PRClosed  PullRequestState = 2
	PRMerged  PullRequestState = 3
	PRUnknown PullRequestState = 4
)

func (v PullRequestState) Valid() bool { return v >= PROpen && v <= PRUnknown }

type PullRequestFilterState uint8

const (
	FilterOpen   PullRequestFilterState = 1
	FilterClosed PullRequestFilterState = 2
	FilterMerged PullRequestFilterState = 3
	FilterAll    PullRequestFilterState = 4
)

func (v PullRequestFilterState) Valid() bool { return v >= FilterOpen && v <= FilterAll }

type ExpectationResult uint8

const (
	NotRequested          ExpectationResult = 1
	Matched               ExpectationResult = 2
	Mismatched            ExpectationResult = 3
	ExpectationUnresolved ExpectationResult = 4
)

func (v ExpectationResult) Valid() bool { return v >= NotRequested && v <= ExpectationUnresolved }

type ProviderKind uint8

const (
	Npm  ProviderKind = 1
	Make ProviderKind = 2
)

func (v ProviderKind) Valid() bool { return v >= Npm && v <= Make }

type PRRole uint8

const (
	PRHeadRole PRRole = 1
	PRBaseRole PRRole = 2
)

func (v PRRole) Valid() bool { return v >= PRHeadRole && v <= PRBaseRole }

type RelationshipEvidence uint8

const (
	RelationComplete RelationshipEvidence = 1
	RelationUnknown  RelationshipEvidence = 2
)

func (v RelationshipEvidence) Valid() bool { return v >= RelationComplete && v <= RelationUnknown }

type WorktreeRelationKind uint8

const (
	ExactSelectedRevision             WorktreeRelationKind = 1
	SameScopedBranchDifferentRevision WorktreeRelationKind = 2
	Unrelated                         WorktreeRelationKind = 3
	UnknownRelation                   WorktreeRelationKind = 4
)

func (v WorktreeRelationKind) Valid() bool { return v >= ExactSelectedRevision && v <= UnknownRelation }

type TerminalDisposition uint8

const (
	Succeeded          TerminalDisposition = 1
	Failed             TerminalDisposition = 2
	TerminalCanceled   TerminalDisposition = 3
	TerminalSuperseded TerminalDisposition = 4
)

func (v TerminalDisposition) Valid() bool { return v >= Succeeded && v <= TerminalSuperseded }

type InvalidationKind uint8

const (
	GitInvalidated       InvalidationKind = 1
	RemoteInvalidated    InvalidationKind = 2
	StorageInvalidated   InvalidationKind = 3
	DiscoveryInvalidated InvalidationKind = 4
	SessionsInvalidated  InvalidationKind = 5
	ContextInvalidated   InvalidationKind = 6
)

func (v InvalidationKind) Valid() bool { return v >= GitInvalidated && v <= ContextInvalidated }
