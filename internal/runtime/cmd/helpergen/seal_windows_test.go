package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsSnapshotRetainsReadOnlyOwnership(t *testing.T) {
	owned := t.TempDir()
	dir := filepath.Join(owned, "selected")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "input.go")
	if err := os.WriteFile(path, []byte("recorded"), 0600); err != nil {
		t.Fatal(err)
	}
	dg, err := openImmutableInput(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	defer dg.Close()
	fg, err := openImmutableInput(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer fg.Close()
	if err := os.WriteFile(path, []byte("unrecorded"), 0600); err == nil {
		t.Fatal("write through sealed input succeeded")
	}
	if err := os.Remove(path); err == nil {
		t.Fatal("delete through sealed input succeeded")
	}
	if err := os.Rename(dir, filepath.Join(owned, "moved")); err == nil {
		t.Fatal("rename through sealed input directory succeeded")
	}
	if err := fg.Close(); err != nil {
		t.Fatal(err)
	}
	if err := dg.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("released"), 0600); err != nil {
		t.Fatalf("positive released-writer control: %v", err)
	}
	if err := os.Rename(dir, filepath.Join(owned, "moved")); err != nil {
		t.Fatalf("positive released-directory control: %v", err)
	}
}
