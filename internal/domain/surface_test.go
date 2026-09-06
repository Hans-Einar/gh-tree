package domain_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func TestInvalidZeroAndClosedPrivateComparableSurface(t *testing.T) {
	// This is an external consumer. A private tagged struct (rather than an
	// interface with a promoted marker method) prevents consumers extending
	// the alternatives through embedding or supplying typed-nil payloads.
	values := []interface{ Valid() bool }{
		domain.RepositoryID{}, domain.WorktreeID{}, domain.BranchID{}, domain.PRID{},
		domain.OID{}, domain.Revision{}, domain.Head{}, domain.StashID{}, domain.LaunchPointID{},
		domain.SessionID{}, domain.ExactTarget{},
	}
	for _, zero := range values {
		typ := reflect.TypeOf(zero)
		if zero.Valid() {
			t.Fatalf("zero %s valid", typ)
		}
		if typ.Kind() != reflect.Struct || !typ.Comparable() {
			t.Fatalf("%s is not a closed comparable value", typ)
		}
		assertPrivateValue(t, typ)
		if reflect.PointerTo(typ).NumMethod() != typ.NumMethod() {
			t.Fatalf("%s has pointer-only methods permitting in-place mutation", typ)
		}
		for i := 0; i < typ.NumMethod(); i++ {
			if strings.HasPrefix(typ.Method(i).Name, "Set") {
				t.Fatalf("%s exports setter %s", typ, typ.Method(i).Name)
			}
		}
	}
	// Explicit comparable constraints also require the compiler to check these
	// representations, independent of reflect's runtime assertions.
	assertComparable[domain.RepositoryID]()
	assertComparable[domain.WorktreeID]()
	assertComparable[domain.BranchID]()
	assertComparable[domain.PRID]()
	assertComparable[domain.OID]()
	assertComparable[domain.Revision]()
	assertComparable[domain.Head]()
	assertComparable[domain.StashID]()
	assertComparable[domain.LaunchPointID]()
	assertComparable[domain.SessionID]()
	assertComparable[domain.ExactTarget]()
	for _, value := range []interface{ Valid() bool }{
		domain.RepositoryScope(0), domain.RepositoryScope(255), domain.BranchKind(0), domain.BranchKind(255),
		domain.ObjectFormat(0), domain.ObjectFormat(255), domain.HeadKind(0), domain.HeadKind(255), domain.TargetKind(0), domain.TargetKind(255),
	} {
		if value.Valid() {
			t.Fatalf("invalid tag %T(%v) accepted", value, value)
		}
	}
	if domain.ObjectFormat(255).ByteLength() != 0 {
		t.Fatal("invalid object format fabricated a length")
	}
}

func assertComparable[T comparable]() {}

func assertPrivateValue(t *testing.T, typ reflect.Type) {
	t.Helper()
	switch typ.Kind() {
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.IsExported() || field.Anonymous || field.Tag != "" {
				t.Fatalf("%s field %s exposes representation/embedding/schema", typ, field.Name)
			}
			assertPrivateValue(t, field.Type)
		}
	case reflect.Array:
		assertPrivateValue(t, typ.Elem())
	case reflect.String, reflect.Uint8, reflect.Uint64:
		// Exact bytes and copied scalar values have no mutable backing exposed.
	default:
		t.Fatalf("%s contains mutable/open/non-value representation", typ)
	}
}

func TestCopiedNestedAccessorsCannotRewriteIntent(t *testing.T) {
	repo := local("repo")
	original := must(domain.NewAttachedHead(branch(repo, domain.Local, "main"), revision(repo)))
	copy := original
	b, _ := copy.Branch()
	r, _ := copy.Revision()
	bytes := r.OID().Bytes()
	bytes[0] = 0
	b = branch(repo, domain.Local, "other")
	r = revision(local("other"))
	copy = must(domain.NewUnbornHead(b))
	if original == copy || original.Repository() != repo {
		t.Fatal("copy changed original HEAD")
	}
	got, ok := original.Revision()
	if !ok || got != revision(repo) || got == r {
		t.Fatal("nested accessor changed original revision")
	}
	gotB, ok := original.Branch()
	if !ok || gotB.Name() != "main" {
		t.Fatal("branch accessor changed original HEAD")
	}
	target := must(domain.NewBranchTarget(gotB, got))
	copyTarget := target
	copyTarget = must(domain.NewCommitTarget(r))
	if target == copyTarget || target.ExpectedRevision() != got {
		t.Fatal("target copy changed expected revision")
	}
}
