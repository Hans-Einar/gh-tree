package domain

import (
	"errors"
	"strconv"
)

// LaunchPointID combines worktree identity with the exact provider, opaque
// project key and member key. Source versions, labels and executable overrides
// are observations or preferences and are deliberately absent.
type LaunchPointID struct {
	worktree WorktreeID
	key      string
}

// NewLaunchPointID permits an empty project key to identify the root project.
// Provider and member keys must be nonempty. All bytes remain literal; adapters
// own project-key canonicalization and provider/member validity.
func NewLaunchPointID(worktree WorktreeID, provider, project, member string) (LaunchPointID, error) {
	if !worktree.Valid() || provider == "" || member == "" {
		return LaunchPointID{}, errors.New("launch identity requires a valid worktree and nonempty provider and member keys")
	}
	key := keyPart(provider) + keyPart(project) + keyPart(member)
	return LaunchPointID{worktree: worktree, key: key}, nil
}

func keyPart(value string) string { return strconv.Itoa(len(value)) + ":" + value }

func (id LaunchPointID) Valid() bool          { return id.worktree.Valid() && id.key != "" }
func (id LaunchPointID) Worktree() WorktreeID { return id.worktree }

// Key is an unambiguous byte-length-delimited identity key, not a path or a
// storage schema. There is no constructor accepting an already encoded key.
func (id LaunchPointID) Key() string                    { return id.key }
func (id LaunchPointID) Equal(other LaunchPointID) bool { return id == other }
