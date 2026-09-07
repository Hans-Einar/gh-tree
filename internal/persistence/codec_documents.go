package persistence

import (
	"bytes"
	"errors"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func decodeUserConfig(raw []byte) (api.UserConfigDocument, error) {
	o, schema, err := documentObject(raw)
	if err != nil {
		return api.UserConfigDocument{}, err
	}
	d := api.UserConfigDocumentData{SchemaVersion: schema}
	if v, present := o.take("stripPrefixes"); present {
		if bytes.Equal(v, []byte("null")) {
			d.StripPrefixes = api.NullField[[]string]()
		} else {
			values, err := stringsValue(v)
			if err != nil {
				return api.UserConfigDocument{}, err
			}
			d.StripPrefixes = api.PresentField(values)
		}
	}
	d.LegacyRepos, err = field(&o, "repos", entries(func(key string, raw []byte) (api.LegacyRepositoryConfig, error) {
		v, err := decodeConfigured(raw)
		if err != nil {
			return api.LegacyRepositoryConfig{}, err
		}
		return api.NewLegacyRepositoryConfig(api.LegacyRepositoryConfigData{Key: key, Value: v})
	}))
	if err != nil {
		return api.UserConfigDocument{}, err
	}
	d.ScopedRepos, err = field(&o, "scopedRepos", entries(func(key string, raw []byte) (api.ScopedRepositoryConfig, error) {
		id, err := parseScopeKey(key)
		if err != nil {
			return api.ScopedRepositoryConfig{}, err
		}
		v, err := decodeConfigured(raw)
		if err != nil {
			return api.ScopedRepositoryConfig{}, err
		}
		return api.NewScopedRepositoryConfig(api.ScopedRepositoryConfigData{RepositoryID: id, Value: v})
	}))
	if err != nil {
		return api.UserConfigDocument{}, err
	}
	d.UnknownMembers, err = o.unknown()
	if err != nil {
		return api.UserConfigDocument{}, err
	}
	return api.NewUserConfigDocument(d)
}

func decodeConfigured(raw []byte) (api.ConfiguredRepository, error) {
	o, err := object(raw)
	if err != nil {
		return api.ConfiguredRepository{}, err
	}
	d := api.ConfiguredRepositoryData{}
	d.Worktrees, err = field(&o, "worktrees", entries(func(name string, raw []byte) (api.ConfiguredWorktreeTarget, error) {
		o, err := object(raw)
		if err != nil {
			return api.ConfiguredWorktreeTarget{}, err
		}
		d := api.ConfiguredWorktreeTargetData{Name: name}
		d.Path, err = field(&o, "path", textValue)
		if err != nil {
			return api.ConfiguredWorktreeTarget{}, err
		}
		d.Branch, err = field(&o, "branch", textValue)
		if err != nil {
			return api.ConfiguredWorktreeTarget{}, err
		}
		d.UnknownMembers, err = o.unknown()
		if err != nil {
			return api.ConfiguredWorktreeTarget{}, err
		}
		return api.NewConfiguredWorktreeTarget(d)
	}))
	if err != nil {
		return api.ConfiguredRepository{}, err
	}
	d.UnknownMembers, err = o.unknown()
	if err != nil {
		return api.ConfiguredRepository{}, err
	}
	return api.NewConfiguredRepository(d)
}

func decodePreferences(raw []byte) (api.PreferencesDocument, error) {
	o, schema, err := documentObject(raw)
	if err != nil {
		return api.PreferencesDocument{}, err
	}
	d := api.PreferencesDocumentData{SchemaVersion: schema}
	decodeLegacy := entries(func(key string, raw []byte) (api.LegacyStringPreference, error) {
		value, err := textValue(raw)
		if err != nil {
			return api.LegacyStringPreference{}, err
		}
		return api.NewLegacyStringPreference(api.LegacyStringPreferenceData{Key: key, Value: value})
	})
	d.LegacyFolders, err = field(&o, "lastFolders", decodeLegacy)
	if err != nil {
		return api.PreferencesDocument{}, err
	}
	d.LegacyWorktrees, err = field(&o, "lastWorktrees", decodeLegacy)
	if err != nil {
		return api.PreferencesDocument{}, err
	}
	d.ScopedPreferences, err = field(&o, "scopedPreferences", entries(decodeScopedPreference))
	if err != nil {
		return api.PreferencesDocument{}, err
	}
	d.UnknownMembers, err = o.unknown()
	if err != nil {
		return api.PreferencesDocument{}, err
	}
	return api.NewPreferencesDocument(d)
}

func decodeScopedPreference(key string, raw []byte) (api.ScopedPreference, error) {
	id, err := parseScopeKey(key)
	if err != nil {
		return api.ScopedPreference{}, err
	}
	o, err := object(raw)
	if err != nil {
		return api.ScopedPreference{}, err
	}
	d := api.ScopedPreferenceData{RepositoryID: id}
	d.LastFolder, err = field(&o, "lastFolder", textValue)
	if err != nil {
		return api.ScopedPreference{}, err
	}
	d.ActiveWorktree, err = field(&o, "activeWorktree", decodeActive)
	if err != nil {
		return api.ScopedPreference{}, err
	}
	d.UnknownMembers, err = o.unknown()
	if err != nil {
		return api.ScopedPreference{}, err
	}
	return api.NewScopedPreference(d)
}
func decodeActive(raw []byte) (api.StoredActiveWorktree, error) {
	o, err := object(raw)
	if err != nil {
		return api.StoredActiveWorktree{}, err
	}
	d := api.StoredActiveWorktreeData{}
	d.AdministrativeKey, err = requiredText(&o, "administrativeKey")
	if err != nil {
		return api.StoredActiveWorktree{}, err
	}
	d.LastKnownPath, err = requiredText(&o, "lastKnownPath")
	if err != nil {
		return api.StoredActiveWorktree{}, err
	}
	d.UnknownMembers, err = o.unknown()
	if err != nil {
		return api.StoredActiveWorktree{}, err
	}
	return api.NewStoredActiveWorktree(d)
}

func decodeRunConfig(raw []byte) (api.RunConfigDocument, error) {
	o, schema, err := documentObject(raw)
	if err != nil {
		return api.RunConfigDocument{}, err
	}
	d := api.RunConfigDocumentData{SchemaVersion: schema}
	d.Default, err = field(&o, "default", textValue)
	if err != nil {
		return api.RunConfigDocument{}, err
	}
	d.Launch, err = field(&o, "launch", entries(func(alias string, raw []byte) (api.SavedLaunchEntry, error) {
		definition, err := decodeDefinition(raw)
		if err != nil {
			return api.SavedLaunchEntry{}, err
		}
		return api.NewSavedLaunchEntry(api.SavedLaunchEntryData{Alias: alias, Definition: definition})
	}))
	if err != nil {
		return api.RunConfigDocument{}, err
	}
	d.UnknownMembers, err = o.unknown()
	if err != nil {
		return api.RunConfigDocument{}, err
	}
	return api.NewRunConfigDocument(d)
}
func decodeDefinition(raw []byte) (api.SavedLaunchDefinition, error) {
	o, err := object(raw)
	if err != nil {
		return api.SavedLaunchDefinition{}, err
	}
	d := api.SavedLaunchDefinitionData{}
	d.Provider, err = requiredText(&o, "provider")
	if err != nil {
		return api.SavedLaunchDefinition{}, err
	}
	d.Dir, err = field(&o, "dir", textValue)
	if err != nil {
		return api.SavedLaunchDefinition{}, err
	}
	d.Script, err = field(&o, "script", textValue)
	if err != nil {
		return api.SavedLaunchDefinition{}, err
	}
	d.Targets, err = field(&o, "targets", stringsValue)
	if err != nil {
		return api.SavedLaunchDefinition{}, err
	}
	d.Command, err = field(&o, "command", textValue)
	if err != nil {
		return api.SavedLaunchDefinition{}, err
	}
	d.UnknownMembers, err = o.unknown()
	if err != nil {
		return api.SavedLaunchDefinition{}, err
	}
	return api.NewSavedLaunchDefinition(d)
}

func encodeUserConfig(v api.UserConfigDocument) ([]byte, error) {
	if !v.Valid() || v.Data().SchemaVersion != 1 {
		return nil, errors.New("storage requires a valid schema1 user document")
	}
	d := v.Data()
	w := objectWriter{}
	w.add("schemaVersion", []byte("1"), nil)
	put(&w, "stripPrefixes", d.StripPrefixes, jsonStrings)
	put(&w, "repos", d.LegacyRepos, func(values []api.LegacyRepositoryConfig) ([]byte, error) {
		return mapValue(values, func(v api.LegacyRepositoryConfig) (string, []byte, error) {
			d := v.Data()
			raw, err := encodeConfigured(d.Value)
			return d.Key, raw, err
		})
	})
	put(&w, "scopedRepos", d.ScopedRepos, func(values []api.ScopedRepositoryConfig) ([]byte, error) {
		return mapValue(values, func(v api.ScopedRepositoryConfig) (string, []byte, error) {
			d := v.Data()
			key, err := scopeKey(d.RepositoryID)
			if err != nil {
				return "", nil, err
			}
			raw, err := encodeConfigured(d.Value)
			return key, raw, err
		})
	})
	w.retained(d.UnknownMembers)
	return checkedDocument(w.finish())
}
func jsonStrings(values []string) ([]byte, error) {
	// Present nil is an explicitly empty JSON array, never the Null field tag.
	if values == nil {
		values = []string{}
	}
	return jsonValue(values)
}
func encodeConfigured(v api.ConfiguredRepository) ([]byte, error) {
	d := v.Data()
	w := objectWriter{}
	put(&w, "worktrees", d.Worktrees, func(values []api.ConfiguredWorktreeTarget) ([]byte, error) {
		return mapValue(values, func(v api.ConfiguredWorktreeTarget) (string, []byte, error) {
			d := v.Data()
			w := objectWriter{}
			put(&w, "path", d.Path, jsonValue[string])
			put(&w, "branch", d.Branch, jsonValue[string])
			w.retained(d.UnknownMembers)
			raw, err := w.finish()
			return d.Name, raw, err
		})
	})
	w.retained(d.UnknownMembers)
	return w.finish()
}
func encodePreferences(v api.PreferencesDocument) ([]byte, error) {
	if !v.Valid() || v.Data().SchemaVersion != 1 {
		return nil, errors.New("storage requires a valid schema1 preferences document")
	}
	d := v.Data()
	w := objectWriter{}
	w.add("schemaVersion", []byte("1"), nil)
	encodeLegacy := func(values []api.LegacyStringPreference) ([]byte, error) {
		return mapValue(values, func(v api.LegacyStringPreference) (string, []byte, error) {
			d := v.Data()
			raw, err := jsonValue(d.Value)
			return d.Key, raw, err
		})
	}
	put(&w, "lastFolders", d.LegacyFolders, encodeLegacy)
	put(&w, "lastWorktrees", d.LegacyWorktrees, encodeLegacy)
	put(&w, "scopedPreferences", d.ScopedPreferences, func(values []api.ScopedPreference) ([]byte, error) {
		return mapValue(values, func(v api.ScopedPreference) (string, []byte, error) {
			d := v.Data()
			key, err := scopeKey(d.RepositoryID)
			if err != nil {
				return "", nil, err
			}
			w := objectWriter{}
			put(&w, "lastFolder", d.LastFolder, jsonValue[string])
			put(&w, "activeWorktree", d.ActiveWorktree, encodeActive)
			w.retained(d.UnknownMembers)
			raw, err := w.finish()
			return key, raw, err
		})
	})
	w.retained(d.UnknownMembers)
	return checkedDocument(w.finish())
}
func encodeActive(v api.StoredActiveWorktree) ([]byte, error) {
	d := v.Data()
	w := objectWriter{}
	raw, err := jsonValue(d.AdministrativeKey)
	w.add("administrativeKey", raw, err)
	raw, err = jsonValue(d.LastKnownPath)
	w.add("lastKnownPath", raw, err)
	w.retained(d.UnknownMembers)
	return w.finish()
}
func encodeRunConfig(v api.RunConfigDocument) ([]byte, error) {
	if !v.Valid() || v.Data().SchemaVersion != 1 {
		return nil, errors.New("storage requires a valid schema1 run document")
	}
	d := v.Data()
	w := objectWriter{}
	w.add("schemaVersion", []byte("1"), nil)
	put(&w, "default", d.Default, jsonValue[string])
	put(&w, "launch", d.Launch, func(values []api.SavedLaunchEntry) ([]byte, error) {
		return mapValue(values, func(v api.SavedLaunchEntry) (string, []byte, error) {
			d := v.Data()
			raw, err := encodeDefinition(d.Definition)
			return d.Alias, raw, err
		})
	})
	w.retained(d.UnknownMembers)
	return checkedDocument(w.finish())
}
func encodeDefinition(v api.SavedLaunchDefinition) ([]byte, error) {
	d := v.Data()
	w := objectWriter{}
	raw, err := jsonValue(d.Provider)
	w.add("provider", raw, err)
	put(&w, "dir", d.Dir, jsonValue[string])
	put(&w, "script", d.Script, jsonValue[string])
	put(&w, "targets", d.Targets, jsonStrings)
	put(&w, "command", d.Command, jsonValue[string])
	w.retained(d.UnknownMembers)
	return w.finish()
}
