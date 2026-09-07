//go:build !windows

package main

import "path/filepath"

func physicalPath(path string) (string, error) { return filepath.EvalSymlinks(path) }
