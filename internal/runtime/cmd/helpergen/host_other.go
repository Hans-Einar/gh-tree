//go:build !windows

package main

import "fmt"

func nativeHost() error { return fmt.Errorf("canonical helper builder requires native Windows amd64") }
