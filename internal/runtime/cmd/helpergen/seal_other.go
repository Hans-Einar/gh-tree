//go:build !windows

package main

import "os"

// Noncanonical hosts run pure tests only; build admission requires Windows.
func openImmutableInput(path string, directory bool) (*os.File, error) { return os.Open(path) }
