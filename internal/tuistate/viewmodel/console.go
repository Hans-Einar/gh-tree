package viewmodel

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"unicode/utf8"
)

// OutputInputSpec is a copied constructor/accessor record.
type OutputInputSpec struct {
	SessionID              domain.SessionID
	SourceGeneration       SourceGeneration
	PresentationGeneration PresentationGeneration
	ContentGeneration      ContentGeneration
	Offset                 uint64
	Bytes                  []byte
	Stream                 StreamKind
	Gaps                   []ByteRange
	End                    bool
}
type OutputInput struct {
	data  OutputInputSpec
	valid bool
}

func NewOutputInput(s OutputInputSpec) (OutputInput, error) {
	if !validOutputInput(s) {
		return OutputInput{}, invalid("OutputInput")
	}
	return OutputInput{cloneOutputInput(s), true}, nil
}
func validOutputInput(s OutputInputSpec) bool {
	return s.SessionID.Valid() && s.SourceGeneration.Valid() && s.PresentationGeneration.Valid() && s.ContentGeneration.Valid() && s.Stream.Valid() && uint64(len(s.Bytes)) <= ^uint64(0)-s.Offset && validGaps(s.Gaps, 0, s.Offset)
}
func cloneOutputInput(s OutputInputSpec) OutputInputSpec {
	s.Bytes = copySlice(s.Bytes)
	s.Gaps = copySlice(s.Gaps)
	return s
}
func (v OutputInput) Valid() bool             { return v.valid && validOutputInput(v.data) }
func (v OutputInput) Fields() OutputInputSpec { return cloneOutputInput(v.data) }
func (v OutputInput) Clone() OutputInput      { return OutputInput{cloneOutputInput(v.data), v.valid} }

// ConsoleLineSpec is a copied constructor/accessor record.
type ConsoleLineSpec struct {
	Text        string
	Stream      StreamKind
	SourceRange ByteRange
	Continued   bool
}
type ConsoleLine struct {
	data  ConsoleLineSpec
	valid bool
}

func NewConsoleLine(s ConsoleLineSpec) (ConsoleLine, error) {
	if !validConsoleLine(s) {
		return ConsoleLine{}, invalid("ConsoleLine")
	}
	return ConsoleLine{cloneConsoleLine(s), true}, nil
}
func validConsoleLine(s ConsoleLineSpec) bool {
	return safeLine(s.Text) && s.Stream.Valid() && s.SourceRange.Valid()
}
func cloneConsoleLine(s ConsoleLineSpec) ConsoleLineSpec { ; return s }
func (v ConsoleLine) Valid() bool                        { return v.valid && validConsoleLine(v.data) }
func (v ConsoleLine) Fields() ConsoleLineSpec            { return cloneConsoleLine(v.data) }
func (v ConsoleLine) Clone() ConsoleLine                 { return ConsoleLine{cloneConsoleLine(v.data), v.valid} }

// ConsolePresentationSpec is a copied constructor/accessor record.
type ConsolePresentationSpec struct {
	SessionID              domain.SessionID
	SourceGeneration       SourceGeneration
	PresentationGeneration PresentationGeneration
	ContentGeneration      ContentGeneration
	Range                  ByteRange
	Lines                  []ConsoleLine
	Gaps                   []ByteRange
	Truncated              bool
	End                    bool
	Notices                []StatusNotice
}
type ConsolePresentation struct {
	data  ConsolePresentationSpec
	valid bool
}

func NewConsolePresentation(s ConsolePresentationSpec) (ConsolePresentation, error) {
	if !validConsolePresentation(s) {
		return ConsolePresentation{}, invalid("ConsolePresentation")
	}
	return ConsolePresentation{cloneConsolePresentation(s), true}, nil
}
func validConsolePresentation(s ConsolePresentationSpec) bool {
	return s.SessionID.Valid() && s.SourceGeneration.Valid() && s.PresentationGeneration.Valid() && s.ContentGeneration.Valid() && s.Range.Valid() && allValid(s.Lines) && consoleLineRanges(s.Lines, s.Range) && validGaps(s.Gaps, s.Range.start, s.Range.end) && allValid(s.Notices)
}
func cloneConsolePresentation(s ConsolePresentationSpec) ConsolePresentationSpec {
	s.Lines = copyValues(s.Lines)
	s.Gaps = copySlice(s.Gaps)
	s.Notices = copyValues(s.Notices)
	return s
}
func (v ConsolePresentation) Valid() bool { return v.valid && validConsolePresentation(v.data) }
func (v ConsolePresentation) Fields() ConsolePresentationSpec {
	return cloneConsolePresentation(v.data)
}
func (v ConsolePresentation) Clone() ConsolePresentation {
	return ConsolePresentation{cloneConsolePresentation(v.data), v.valid}
}

// ConsoleSummarySpec is a copied constructor/accessor record.
type ConsoleSummarySpec struct {
	Label             string
	ExecutableDisplay string
	ArgumentDisplay   []string
	CwdDisplay        string
	WorktreeID        domain.WorktreeID
	Terminal          bool
}
type ConsoleSummary struct {
	data  ConsoleSummarySpec
	valid bool
}

func NewConsoleSummary(s ConsoleSummarySpec) (ConsoleSummary, error) {
	if !validConsoleSummary(s) {
		return ConsoleSummary{}, invalid("ConsoleSummary")
	}
	return ConsoleSummary{cloneConsoleSummary(s), true}, nil
}
func validConsoleSummary(s ConsoleSummarySpec) bool { return s.WorktreeID.Valid() }
func cloneConsoleSummary(s ConsoleSummarySpec) ConsoleSummarySpec {
	s.ArgumentDisplay = copySlice(s.ArgumentDisplay)
	return s
}
func (v ConsoleSummary) Valid() bool                { return v.valid && validConsoleSummary(v.data) }
func (v ConsoleSummary) Fields() ConsoleSummarySpec { return cloneConsoleSummary(v.data) }
func (v ConsoleSummary) Clone() ConsoleSummary {
	return ConsoleSummary{cloneConsoleSummary(v.data), v.valid}
}

// ConsoleModelSpec is a copied constructor/accessor record.
type ConsoleModelSpec struct {
	SessionID              domain.SessionID
	SourceGeneration       SourceGeneration
	PresentationGeneration PresentationGeneration
	ContentGeneration      ContentGeneration
	Phase                  SessionPhase
	Cleanup                CleanupState
	Activity               Activity
	ExitCode               Optional[int]
	ExitSignal             Optional[string]
	Capabilities           ConsoleCapabilities
	InputFocused           bool
	Summary                ConsoleSummary
	Presentation           Optional[ConsolePresentation]
	Scroll                 Scroll
	Notices                []StatusNotice
}
type ConsoleModel struct {
	data  ConsoleModelSpec
	valid bool
}

func NewConsoleModel(s ConsoleModelSpec) (ConsoleModel, error) {
	if !validConsoleModel(s) {
		return ConsoleModel{}, invalid("ConsoleModel")
	}
	return ConsoleModel{cloneConsoleModel(s), true}, nil
}
func validConsoleModel(s ConsoleModelSpec) bool {
	return s.SessionID.Valid() && s.SourceGeneration.Valid() && s.PresentationGeneration.Valid() && s.ContentGeneration.Valid() && s.Phase.Valid() && s.Cleanup.Valid() && s.Activity.Valid() && s.Capabilities.Valid() && (!s.ExitSignal.present || s.ExitSignal.value != "") && s.Summary.Valid() && (!s.Capabilities.resize || s.Summary.data.Terminal) && (!s.InputFocused || s.Capabilities.input && s.Phase == RunningSession) && (s.Phase == CleanedSession) == (s.Cleanup == CleanupComplete) && optionalValid(s.Presentation) && consoleCorrelation(s) && s.Scroll.Valid() && allValid(s.Notices)
}
func cloneConsoleModel(s ConsoleModelSpec) ConsoleModelSpec {
	s.Summary = s.Summary.Clone()
	s.Presentation = copyOptional(s.Presentation)
	s.Notices = copyValues(s.Notices)
	return s
}
func (v ConsoleModel) Valid() bool              { return v.valid && validConsoleModel(v.data) }
func (v ConsoleModel) Fields() ConsoleModelSpec { return cloneConsoleModel(v.data) }
func (v ConsoleModel) Clone() ConsoleModel      { return ConsoleModel{cloneConsoleModel(v.data), v.valid} }

type StreamKind uint8

const (
	StdoutStream StreamKind = iota + 1
	StderrStream
	TerminalStream
)

func (v StreamKind) Valid() bool { return v >= StdoutStream && v <= TerminalStream }

type SessionPhase uint8

const (
	StartingSession SessionPhase = iota + 1
	RunningSession
	StoppingSession
	CleanedSession
	CleanupFailedSession
)

func (v SessionPhase) Valid() bool { return v >= StartingSession && v <= CleanupFailedSession }

type CleanupState uint8

const (
	CleanupPending CleanupState = iota + 1
	CleanupComplete
	CleanupResidual
)

func (v CleanupState) Valid() bool { return v >= CleanupPending && v <= CleanupResidual }

type Activity uint8

const (
	ActivityUnknown Activity = iota + 1
	OutputObserved
)

func (v Activity) Valid() bool { return v >= ActivityUnknown && v <= OutputObserved }

// ByteRange is an absolute half-open transport range; empty ranges are valid.
type ByteRange struct{ start, end uint64 }

func NewByteRange(start, end uint64) (ByteRange, error) {
	if end < start {
		return ByteRange{}, invalid("byte range")
	}
	return ByteRange{start, end}, nil
}
func (r ByteRange) Valid() bool   { return r.end >= r.start }
func (r ByteRange) Start() uint64 { return r.start }
func (r ByteRange) End() uint64   { return r.end }
func validGaps(gs []ByteRange, start, end uint64) bool {
	previous := start
	for _, g := range gs {
		if !g.Valid() || g.start < previous || g.start == g.end || g.end > end {
			return false
		}
		previous = g.end
	}
	return true
}
func consoleLineRanges(ls []ConsoleLine, r ByteRange) bool {
	for _, l := range ls {
		if l.data.SourceRange.start < r.start || l.data.SourceRange.end > r.end {
			return false
		}
	}
	return true
}
func safeLine(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r >= 0x7f && r <= 0x9f {
			return false
		}
	}
	return true
}
func consoleCorrelation(s ConsoleModelSpec) bool {
	if !s.Presentation.present {
		return true
	}
	p := s.Presentation.value.data
	return p.SessionID == s.SessionID && p.SourceGeneration == s.SourceGeneration && p.PresentationGeneration == s.PresentationGeneration && p.ContentGeneration == s.ContentGeneration
}

// Capabilities are advertised facts. Tree stop and terminal interrupt differ.
type ConsoleCapabilities struct{ input, resize, interrupt, treeStop, valid bool }

func NewConsoleCapabilities(input, resize, interrupt, treeStop bool) ConsoleCapabilities {
	return ConsoleCapabilities{input, resize, interrupt, treeStop, true}
}
func (c ConsoleCapabilities) Valid() bool     { return c.valid }
func (c ConsoleCapabilities) Input() bool     { return c.input }
func (c ConsoleCapabilities) Resize() bool    { return c.resize }
func (c ConsoleCapabilities) Interrupt() bool { return c.interrupt }
func (c ConsoleCapabilities) TreeStop() bool  { return c.treeStop }
