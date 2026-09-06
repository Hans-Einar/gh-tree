//go:build linux || darwin || freebsd

package persistence

import (
	"errors"

	"golang.org/x/sys/unix"
)

// No namespace fallback is permitted. The caller owns the complete preparation,
// version/metadata checks and manifest barriers; this is the native commit point.
// Expected-absence publication deliberately leaves the payload's owned name:
// cleanup and directory durability are separate postcommit facts.
func unixPublish(parent *unixObject, payloadName, targetName string, expectedPresent bool) error {
	if parent == nil || !singleName(payloadName) || !singleName(targetName) || payloadName == targetName {
		return errors.New("invalid native publication arguments")
	}
	if expectedPresent {
		return unix.Renameat(parent.fd(), payloadName, parent.fd(), targetName)
	}
	return unix.Linkat(parent.fd(), payloadName, parent.fd(), targetName, 0)
}

func unixRetainOriginal(original, parent *unixObject, targetName, retainedName string) (*unixObject, error) {
	if original == nil || parent == nil || !singleName(targetName) || !singleName(retainedName) || targetName == retainedName {
		return nil, errors.New("invalid original retention arguments")
	}
	if err := unix.Linkat(parent.fd(), targetName, parent.fd(), retainedName, 0); err != nil {
		return nil, err
	}
	retained, err := unixOpenDocument(parent, retainedName)
	if err != nil {
		return nil, err
	} // Never delete an unobserved recovery object.
	current, err := unixObserve(original.fd())
	// Link creation changes ctime. Observe the still-open original again and
	// compare native identity with the link, not the historical fallback ctime.
	if err != nil || !current.sameObject(retained.observation) {
		return nil, errors.Join(err, errors.New("retained original identity differs"), retained.close())
	}
	return retained, nil
}
