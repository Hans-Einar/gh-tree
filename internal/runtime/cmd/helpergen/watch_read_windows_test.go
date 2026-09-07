package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryWatchReadProbe(t *testing.T) {
	for _, operation := range []string{"stat-child", "read-self", "read-child", "open-child", "read-file"} {
		t.Run(operation, func(t *testing.T) {
			root := t.TempDir()
			child := filepath.Join(root, "child")
			os.Mkdir(child, 0700)
			os.WriteFile(filepath.Join(child, "file.txt"), []byte("recorded"), 0600)
			w, err := openDirectoryWatch(root)
			if err != nil {
				t.Fatal(err)
			}
			if err := w.check(); err != nil {
				t.Fatal(err)
			}
			switch operation {
			case "stat-child":
				_, err = os.Stat(child)
			case "read-self":
				_, err = os.ReadDir(root)
			case "read-child":
				_, err = os.ReadDir(child)
			case "open-child":
				var f *os.File
				f, err = os.Open(child)
				if err == nil {
					f.Close()
				}
			case "read-file":
				_, err = os.ReadFile(filepath.Join(child, "file.txt"))
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := w.check(); err != nil {
				w.close()
				t.Fatalf("read invalidated file set: %v", err)
			}
			if err := w.close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
