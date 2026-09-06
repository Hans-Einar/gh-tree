package ports

import (
	"errors"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func invalid(s string) error { return errors.New("invalid port value: " + s) }

// Loads preserve result-plus-error observations. A usable document requires a
// valid matching family/content observation; corrupt/unsupported data is absent.
type LoadedUserConfig struct {
	observation api.StorageLoadObservation
	document    api.Optional[api.UserConfigDocument]
}

func NewLoadedUserConfig(o api.StorageLoadObservation, d api.Optional[api.UserConfigDocument]) (LoadedUserConfig, error) {
	if !validLoad(o, api.UserConfig, d.Present()) {
		return LoadedUserConfig{}, invalid("user config load")
	}
	if v, p := d.Value(); p && (!v.Valid() || !schemaMatches(o, v.Data().SchemaVersion)) {
		return LoadedUserConfig{}, invalid("user document")
	}
	return LoadedUserConfig{o, d}, nil
}
func (v LoadedUserConfig) Valid() bool {
	_, e := NewLoadedUserConfig(v.observation, v.document)
	return e == nil
}
func (v LoadedUserConfig) Observation() api.StorageLoadObservation        { return v.observation }
func (v LoadedUserConfig) Document() api.Optional[api.UserConfigDocument] { return v.document }

type LoadedPreferences struct {
	observation api.StorageLoadObservation
	document    api.Optional[api.PreferencesDocument]
}

func NewLoadedPreferences(o api.StorageLoadObservation, d api.Optional[api.PreferencesDocument]) (LoadedPreferences, error) {
	if !validLoad(o, api.Preferences, d.Present()) {
		return LoadedPreferences{}, invalid("preferences load")
	}
	if v, p := d.Value(); p && (!v.Valid() || !schemaMatches(o, v.Data().SchemaVersion)) {
		return LoadedPreferences{}, invalid("preferences document")
	}
	return LoadedPreferences{o, d}, nil
}
func (v LoadedPreferences) Valid() bool {
	_, e := NewLoadedPreferences(v.observation, v.document)
	return e == nil
}
func (v LoadedPreferences) Observation() api.StorageLoadObservation         { return v.observation }
func (v LoadedPreferences) Document() api.Optional[api.PreferencesDocument] { return v.document }

type LoadedRunConfig struct {
	scope       api.WorktreeScope
	observation api.StorageLoadObservation
	document    api.Optional[api.RunConfigDocument]
}

func NewLoadedRunConfig(s api.WorktreeScope, o api.StorageLoadObservation, d api.Optional[api.RunConfigDocument]) (LoadedRunConfig, error) {
	if !s.Valid() || !validLoad(o, api.RunConfig, d.Present()) || !loadMatchesScope(o, s) {
		return LoadedRunConfig{}, invalid("run load")
	}
	if v, p := d.Value(); p && (!v.Valid() || !schemaMatches(o, v.Data().SchemaVersion)) {
		return LoadedRunConfig{}, invalid("run document")
	}
	return LoadedRunConfig{s, o, d}, nil
}
func (v LoadedRunConfig) Valid() bool {
	_, e := NewLoadedRunConfig(v.scope, v.observation, v.document)
	return e == nil
}
func (v LoadedRunConfig) Scope() api.WorktreeScope                      { return v.scope }
func (v LoadedRunConfig) Observation() api.StorageLoadObservation       { return v.observation }
func (v LoadedRunConfig) Document() api.Optional[api.RunConfigDocument] { return v.document }
func validLoad(o api.StorageLoadObservation, f api.StorageFamily, document bool) bool {
	if !o.Valid() {
		return false
	}
	d := o.Data()
	usable := d.State == api.LoadAbsent || d.State == api.ValidLegacy || d.State == api.ValidCurrent
	if usable != document {
		return false
	}
	if v, p := d.Version.Value(); p && v.Family() != f {
		return false
	}
	for _, r := range d.Recovery {
		if r.Data().Family != f {
			return false
		}
	}
	return true
}

type UserConfigCommit struct {
	expected StorageVersion
	document api.UserConfigDocument
}

func NewUserConfigCommit(v StorageVersion, d api.UserConfigDocument) (UserConfigCommit, error) {
	if !v.Valid() || v.Family() != api.UserConfig || !d.Valid() || d.Data().SchemaVersion != 1 {
		return UserConfigCommit{}, invalid("user config commit")
	}
	return UserConfigCommit{v, d}, nil
}
func (v UserConfigCommit) Valid() bool {
	_, e := NewUserConfigCommit(v.expected, v.document)
	return e == nil
}
func (v UserConfigCommit) Expected() StorageVersion         { return v.expected }
func (v UserConfigCommit) Document() api.UserConfigDocument { return v.document }

type PreferencesCommit struct {
	expected StorageVersion
	document api.PreferencesDocument
}

func NewPreferencesCommit(v StorageVersion, d api.PreferencesDocument) (PreferencesCommit, error) {
	if !v.Valid() || v.Family() != api.Preferences || !d.Valid() || d.Data().SchemaVersion != 1 {
		return PreferencesCommit{}, invalid("preferences commit")
	}
	return PreferencesCommit{v, d}, nil
}
func (v PreferencesCommit) Valid() bool {
	_, e := NewPreferencesCommit(v.expected, v.document)
	return e == nil
}
func (v PreferencesCommit) Expected() StorageVersion          { return v.expected }
func (v PreferencesCommit) Document() api.PreferencesDocument { return v.document }

type RunConfigCommit struct {
	scope    api.WorktreeScope
	expected StorageVersion
	document api.RunConfigDocument
}

func NewRunConfigCommit(s api.WorktreeScope, v StorageVersion, d api.RunConfigDocument) (RunConfigCommit, error) {
	if !s.Valid() || !v.Valid() || v.Family() != api.RunConfig || !v.MatchesRunScope(s) || !d.Valid() || d.Data().SchemaVersion != 1 {
		return RunConfigCommit{}, invalid("run config commit")
	}
	return RunConfigCommit{s, v, d}, nil
}
func (v RunConfigCommit) Valid() bool {
	_, e := NewRunConfigCommit(v.scope, v.expected, v.document)
	return e == nil
}
func (v RunConfigCommit) Scope() api.WorktreeScope        { return v.scope }
func (v RunConfigCommit) Expected() StorageVersion        { return v.expected }
func (v RunConfigCommit) Document() api.RunConfigDocument { return v.document }

func loadMatchesScope(o api.StorageLoadObservation, s api.WorktreeScope) bool {
	if !o.Valid() {
		return false
	}
	if v, p := o.Data().Version.Value(); p {
		if !v.MatchesRunScope(s) {
			return false
		}
	}
	for _, r := range o.Data().Recovery {
		if !api.StorageRecoveryMatchesScope(r, s) {
			return false
		}
	}
	return true
}

func schemaMatches(o api.StorageLoadObservation, schema uint32) bool {
	switch o.Data().State {
	case api.ValidLegacy:
		return schema == 0
	case api.ValidCurrent, api.LoadAbsent:
		return schema == 1
	}
	return false
}
