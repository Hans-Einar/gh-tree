//go:build !windows

package main

// Only portable structural tests reach these stubs. The executable admission
// gate refuses actual builds on every noncanonical host before materialization.
type directoryWatch struct{}

func openDirectoryWatch(string) (*directoryWatch, error) { return &directoryWatch{}, nil }
func (*directoryWatch) check() error                     { return nil }
func (*directoryWatch) close() error                     { return nil }
