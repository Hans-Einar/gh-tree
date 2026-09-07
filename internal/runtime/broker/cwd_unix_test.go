//go:build linux || darwin || freebsd

package broker

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

func nativeSpec(t *testing.T) StartSpec {
	t.Helper()
	// Keep race instrumentation enabled, but remove its artificial one-second
	// process-exit sleep from these deliberately short owned-helper fixtures.
	// Existing race options/reporting remain present. Product budgets are unchanged.
	t.Setenv("GORACE", strings.TrimSpace(os.Getenv("GORACE")+" atexit_sleep_ms=0"))
	root := must(filepath.EvalSymlinks(t.TempDir()))
	project := filepath.Join(root, " project")
	if err := os.Mkdir(project, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "marker"), []byte("selected-original"), 0600); err != nil {
		t.Fatal(err)
	}
	rootFile := must(os.Open(root))
	defer rootFile.Close()
	projectFile := must(os.Open(project))
	defer projectFile.Close()
	spec := testSpec()
	spec.Environment = append(spec.Environment, "GORACE="+os.Getenv("GORACE"))
	spec.RootLocator = root
	spec.RootIdentity = must(ObserveDirectory(rootFile, ""))
	spec.ProjectIdentity = must(ObserveDirectory(projectFile, ""))
	return spec
}

func closeAcquired(t *testing.T, a *AcquiredDirectory) {
	t.Helper()
	if a != nil {
		if err := a.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func readMarker(t *testing.T, a *AcquiredDirectory) string {
	t.Helper()
	fd, err := unix.Openat(int(a.File().Fd()), "marker", unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		t.Fatal(err)
	}
	f := os.NewFile(uintptr(fd), "fixture-marker")
	defer f.Close()
	return string(must(io.ReadAll(f)))
}

func TestNativeCwdAcquiresOriginalObjectAndRefusesRelocation(t *testing.T) {
	spec := nativeSpec(t)
	before := must(os.Getwd())
	a, err := AcquireCwd(spec)
	defer closeAcquired(t, a)
	if err != nil {
		t.Fatal(err)
	}
	if readMarker(t, a) != "selected-original" {
		t.Fatal("wrong acquired object")
	}
	original := filepath.Join(spec.RootLocator, spec.Components[0])
	moved := filepath.Join(spec.RootLocator, "moved")
	if err := os.Rename(original, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(original, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(original, "marker"), []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if readMarker(t, a) != "selected-original" {
		t.Fatal("retained descriptor followed replacement")
	}
	if err := a.Revalidate(); err == nil {
		t.Fatal("observed relocation accepted")
	}
	if after := must(os.Getwd()); after != before {
		t.Fatal("parent cwd changed")
	}
}

func TestNativeCwdRejectsStaleAndLinkedComponents(t *testing.T) {
	for _, kind := range []string{"stale", "symlink", "root-redirect"} {
		t.Run(kind, func(t *testing.T) {
			spec := nativeSpec(t)
			project := filepath.Join(spec.RootLocator, spec.Components[0])
			switch kind {
			case "stale":
				spec.ProjectIdentity = must(api.NewDirectoryIdentity(api.DirectoryUnix, spec.ProjectIdentity.Device()+1, spec.ProjectIdentity.FileID(), spec.ProjectIdentity.Stamp()))
			case "symlink":
				if err := os.Rename(project, project+"-original"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(project+"-original", project); err != nil {
					t.Fatal(err)
				}
			case "root-redirect":
				alias := filepath.Join(t.TempDir(), "root-link")
				if err := os.Symlink(spec.RootLocator, alias); err != nil {
					t.Fatal(err)
				}
				spec.RootLocator = alias
			}
			a, err := AcquireCwd(spec)
			defer closeAcquired(t, a)
			if err == nil {
				t.Fatal("stale/redirected scope accepted")
			}
		})
	}
}

func TestNativeCwdReplacementAtAcquisitionBarrier(t *testing.T) {
	for _, stage := range []string{"root-acquired", "project-acquired"} {
		t.Run(stage, func(t *testing.T) {
			spec := nativeSpec(t)
			a, err := acquireCwd(spec, func(at string) {
				if at != stage {
					return
				}
				project := filepath.Join(spec.RootLocator, spec.Components[0])
				if err := os.Rename(project, project+"-original"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(project, 0700); err != nil {
					t.Fatal(err)
				}
			})
			defer closeAcquired(t, a)
			if err == nil {
				t.Fatal("barrier substitution accepted")
			}
		})
	}
}

func TestNativeCwdPartialAcquisitionCloseAndUnknownIdentityProfile(t *testing.T) {
	spec := nativeSpec(t)
	spec.Components = append(spec.Components, "absent")
	a, err := AcquireCwd(spec)
	if err == nil || a == nil || len(a.chain) == 0 {
		t.Fatal("partial acquisition ownership lost")
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(); err != nil {
		t.Fatal("non-idempotent close")
	}
	if a.File() != nil || !errors.Is(a.Revalidate(), ErrCwd) {
		t.Fatal("closed capability usable")
	}
	spec = nativeSpec(t)
	spec.ProjectIdentity = must(api.NewDirectoryIdentity(api.DirectoryUnix, spec.ProjectIdentity.Device(), spec.ProjectIdentity.FileID(), "unknown:1"))
	a, err = AcquireCwd(spec)
	defer closeAcquired(t, a)
	if err == nil {
		t.Fatal("unknown stamp profile accepted")
	}
}
