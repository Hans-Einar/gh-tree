package persistence

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDarwinNativeExtendedACLRefusal(t *testing.T) {
	root := physicalTemp(t)
	parent := acquiredUnix(t, root).parent()
	source := unixTestPayload(t, parent, "source-acl", []byte("old"))
	// Native chmod fixture construction only. Production inspects the retained
	// descriptor using fgetattrlist and never calls chmod or parses its output.
	cmd := exec.Command("/bin/chmod", "+a", "everyone deny write", filepath.Join(root, "source-acl"))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("native ACL fixture: %v %s", err, output)
	}
	if _, err := unixInspectMetadata(source); err == nil {
		t.Fatal("Darwin extended ACL was silently reduced to mode bits")
	}
}
