package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func nativeFixture(t *testing.T, format string) (string, *Adapter) {
	t.Helper()
	name := os.Getenv("GH_TREE_TEST_GIT")
	if name == "" {
		name = "git"
	}
	git, err := exec.LookPath(name)
	if err != nil {
		t.Skip("native Git unavailable")
	}
	root := t.TempDir()
	env := commandEnvironment(os.Environ())
	env = append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+filepath.Join(root, "absent-global"), "HOME="+root, "XDG_CONFIG_HOME="+root)
	a, err := New(Options{GitExecutable: git, CurrentDirectory: root, Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"init", "--object-format=" + format, "--initial-branch=main"}
	if storage := os.Getenv("GH_TREE_TEST_REF_STORAGE"); storage != "" {
		args = append(args, "--ref-format="+storage)
	}
	args = append(args, ".")
	r := a.command(context.Background(), root, true, args...)
	if r.err != nil {
		t.Fatalf("Git init %s: %v", format, r.err)
	}
	return root, a
}

func TestNativeReadDoesNotRefreshIndex(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			path := filepath.Join(root, "é file.txt")
			if err := os.WriteFile(path, []byte("original\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if r := a.command(context.Background(), root, true, "add", "--", "é file.txt"); r.err != nil {
				t.Fatal(r.err)
			}
			indexPath := filepath.Join(root, ".git", "index")
			before, err := os.ReadFile(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, []byte("changed\n"), 0600); err != nil {
				t.Fatal(err)
			}
			r := a.command(context.Background(), root, false, "status", "--porcelain=v2", "-z", "--untracked-files=all")
			if r.err != nil || !strings.Contains(string(r.stdout), "é file.txt") {
				t.Fatalf("status: %v %q", r.err, r.stdout)
			}
			after, err := os.ReadFile(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			now, err := os.Stat(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) || !os.SameFile(info, now) || info.ModTime() != now.ModTime() {
				t.Fatal("read refreshed or replaced the index")
			}
			version := a.command(context.Background(), root, false, "--version")
			t.Logf("native profile: %s; object format %s", strings.TrimSpace(string(version.stdout)), format)
		})
	}
}
