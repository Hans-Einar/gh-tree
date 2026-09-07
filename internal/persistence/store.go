package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
)

var (
	errUnsupportedProfile = errors.New("unsupported native storage profile")
	errBindingChanged     = errors.New("storage binding changed")
	errInvalidRequest     = errors.New("invalid storage request")
	errLockBusy           = errors.New("storage lock busy")
	errCleanupIncomplete  = errors.New("storage resource cleanup incomplete")
)

// Options contains explicit Composition-selected absolute locations and bounded
// construction settings. Zero limits select the frozen initial budgets. New
// never reads environment/default paths or creates directories or store files.
type Options struct {
	UserConfigPath     string
	PreferencesPath    string
	LockWait           time.Duration
	RecoveryMaxRecords int
	RecoveryMaxBytes   int64
}

// Store retains no native handles across requests. Its two user bindings are
// fixed at construction; run requests independently acquire their Git-issued
// root and the sole literal .gh-tree/run.json child.
type Store struct {
	user, preferences storeBinding
	options           Options
	hook              func(string) error // immutable private test instrumentation
}

func New(ctx context.Context, options Options) (*Store, error) {
	if options.LockWait == 0 {
		options.LockWait = 5 * time.Second
	}
	if options.RecoveryMaxRecords == 0 {
		options.RecoveryMaxRecords = 256
	}
	if options.RecoveryMaxBytes == 0 {
		options.RecoveryMaxBytes = 1 << 30
	}
	if options.LockWait <= 0 || options.LockWait > 5*time.Second || options.RecoveryMaxRecords < 1 || options.RecoveryMaxRecords > 256 || options.RecoveryMaxBytes < 1 || options.RecoveryMaxBytes > 1<<30 {
		return nil, errors.New("invalid storage construction limits")
	}
	user, err := bindExplicit(ctx, api.UserConfig, options.UserConfigPath)
	if err != nil {
		return nil, err
	}
	preferences, err := bindExplicit(ctx, api.Preferences, options.PreferencesPath)
	if err != nil {
		return nil, err
	}
	overlap, err := bindingsOverlap(ctx, user, preferences)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, errors.New("storage family bindings overlap")
	}
	return &Store{user: user, preferences: preferences, options: options}, nil
}

func (s *Store) LoadUserConfig(ctx context.Context) (ports.LoadedUserConfig, error) {
	d, o, err := s.load(ctx, api.UserConfig, api.WorktreeScope{})
	value := api.None[api.UserConfigDocument]()
	if d != nil {
		value = api.Some(d.user)
	}
	r, buildErr := ports.NewLoadedUserConfig(o, value)
	return r, errors.Join(err, buildErr)
}
func (s *Store) LoadPreferences(ctx context.Context) (ports.LoadedPreferences, error) {
	d, o, err := s.load(ctx, api.Preferences, api.WorktreeScope{})
	value := api.None[api.PreferencesDocument]()
	if d != nil {
		value = api.Some(d.preferences)
	}
	r, buildErr := ports.NewLoadedPreferences(o, value)
	return r, errors.Join(err, buildErr)
}
func (s *Store) LoadRunConfig(ctx context.Context, scope api.WorktreeScope) (ports.LoadedRunConfig, error) {
	d, o, err := s.load(ctx, api.RunConfig, scope)
	value := api.None[api.RunConfigDocument]()
	if d != nil {
		value = api.Some(d.run)
	}
	r, buildErr := ports.NewLoadedRunConfig(scope, o, value)
	return r, errors.Join(err, buildErr)
}

func (s *Store) load(ctx context.Context, family api.StorageFamily, scope api.WorktreeScope) (*document, api.StorageLoadObservation, error) {
	o := api.StorageLoadObservationData{State: api.LoadUnavailable}
	var doc *document
	var chain *nativeChain
	var err error
	if s == nil {
		err = errors.Join(errInvalidRequest, errors.New("uninitialized storage"))
	}
	if err == nil {
		switch family {
		case api.UserConfig:
			chain, err = s.user.acquire(ctx)
		case api.Preferences:
			chain, err = s.preferences.acquire(ctx)
		case api.RunConfig:
			chain, err = acquireRun(ctx, scope)
		default:
			err = errors.New("invalid storage family")
		}
	}
	if err == nil {
		if family == api.RunConfig {
			for _, bound := range []storeBinding{s.user, s.preferences} {
				other, acquireErr := bound.acquire(ctx)
				if acquireErr != nil {
					err = acquireErr
					break
				}
				overlap, checkErr := acquiredOverlap(chain, "run.json", other, bound.basename)
				checkErr = errors.Join(checkErr, other.close())
				if checkErr != nil {
					err = checkErr
					break
				}
				if overlap {
					err = errors.Join(errInvalidRequest, errors.New("run scope overlaps a bound user store"))
					break
				}
			}
		}
		basename := "run.json"
		if family == api.UserConfig {
			basename = s.user.basename
		}
		if family == api.Preferences {
			basename = s.preferences.basename
		}
		if err == nil {
			var lock *nativeStoreLock
			if len(chain.remaining) == 0 {
				lock, err = nativeExistingLock(ctx, chain.parent(), basename, s.options.LockWait)
				if errors.Is(err, os.ErrNotExist) {
					err = nil
				}
			}
			if err == nil {
				doc, o, err = loadAcquired(ctx, chain, family, scope, basename)
				if lock != nil {
					locator := filepath.Join(scope.Data().RootLocator, ".gh-tree")
					if family == api.UserConfig {
						locator = s.user.parentPath
					}
					if family == api.Preferences {
						locator = s.preferences.parentPath
					}
					names, _, _, inventoryErr := inventoryRecovery(ctx, chain.parent(), basename, s.options.RecoveryMaxRecords, s.options.RecoveryMaxBytes)
					err = errors.Join(err, inventoryErr)
					for _, name := range names {
						if !manifestName(name) {
							continue
						}
						recovery, recoveryErr := observeManifest(ctx, chain, name, basename, locator, family, scope)
						o.Recovery = append(o.Recovery, recovery...)
						err = errors.Join(err, recoveryErr)
					}
				}
			}
			if lock != nil {
				err = errors.Join(err, lock.close())
			}
		}
		err = errors.Join(err, chain.close())
	}
	if err != nil {
		if nativeUnsupported(err) && o.State == api.LoadUnavailable {
			o.State = api.UnsupportedProfile
		}
		var codec *codecError
		if errors.As(err, &codec) && doc == nil {
			o.State, o.SchemaVersion = codec.state, codec.schema
		}
		diagnostic := storageDiagnostic("load", err)
		o.Diagnostics = append(o.Diagnostics, diagnostic)
		err = errors.Join(diagnostic, err)
	}
	observation, buildErr := api.NewStorageLoadObservation(o)
	return doc, observation, errors.Join(err, buildErr)
}

func loadAcquired(ctx context.Context, c *nativeChain, family api.StorageFamily, scope api.WorktreeScope, basename string) (*document, api.StorageLoadObservationData, error) {
	o := api.StorageLoadObservationData{State: api.LoadUnavailable}
	store, err := acquiredToken(c, basename)
	if err != nil {
		return nil, o, err
	}
	var object *nativeObject
	if len(c.remaining) == 0 {
		object, err = nativeOpenDocument(c.parent(), basename)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, o, err
	}
	if object == nil {
		if err := nativeRevalidate(ctx, c); err != nil {
			return nil, o, err
		}
		version, err := contentVersion(family, store, scope, false, nil)
		if err != nil {
			return nil, o, err
		}
		d, err := emptyDocument(family)
		if err != nil {
			return nil, o, err
		}
		o.State, o.Version = api.LoadAbsent, api.Some(version)
		return &d, o, nil
	}
	raw, readErr := nativeRead(ctx, object)
	closeErr := object.close()
	if readErr != nil {
		return nil, o, errors.Join(readErr, closeErr)
	}
	if err := nativeRevalidate(ctx, c); err != nil {
		return nil, o, errors.Join(err, closeErr)
	}
	version, err := contentVersion(family, store, scope, true, raw)
	if err != nil {
		return nil, o, errors.Join(err, closeErr)
	}
	o.Version = api.Some(version)
	d, err := decodeDocument(family, raw)
	if err != nil {
		var codec *codecError
		if errors.As(err, &codec) {
			o.State, o.SchemaVersion = codec.state, codec.schema
		}
		return nil, o, errors.Join(err, closeErr)
	}
	o.State = api.ValidCurrent
	if d.schema() == 0 {
		o.State = api.ValidLegacy
	}
	o.SchemaVersion = api.Some(d.schema())
	return &d, o, closeErr
}

func storageDiagnostic(stage string, cause error) api.Diagnostic {
	if notices, only := recoveryNotices(cause); only && len(notices) != 0 {
		return notices[0]
	}
	code := api.IOFailure
	if errors.Is(cause, errCleanupIncomplete) {
		code = api.CleanupIncomplete
	} else if nativeUnsupported(cause) {
		code = api.Unsupported
	} else if errors.Is(cause, errLockBusy) || errors.Is(cause, errRecoveryCapacity) {
		code = api.Busy
	} else if errors.Is(cause, errIncompletePreparation) {
		code = api.Unavailable
	} else if errors.Is(cause, errInvalidRequest) {
		code = api.Invalid
	} else if errors.Is(cause, errBindingChanged) {
		code = api.Conflict
	} else if errors.Is(cause, context.Canceled) || errors.Is(cause, context.DeadlineExceeded) {
		code = api.Canceled
	} else if errors.Is(cause, os.ErrPermission) {
		code = api.Permission
	}
	var codec *codecError
	if errors.As(cause, &codec) {
		code = api.Invalid
		if codec.state == api.UnsupportedVersion {
			code = api.Unsupported
		}
	}
	// Native causes remain in the error chain. Do not dump configuration bytes,
	// executable intent or credentials into the portable diagnostic message.
	d, _ := api.NewDiagnostic(api.DiagnosticData{Code: code, Reason: "storage-" + stage, Message: "Storage " + stage + " did not complete without diagnostics."})
	return d
}
