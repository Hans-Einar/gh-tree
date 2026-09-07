package launchdiscovery

import (
	"context"
	"strings"
	"testing"
)

func TestStrictManifestAndLiteralMembers(t *testing.T) {
	for _, b := range []string{`{"scripts":{"x":"a","x":"b"}}`, `{"scripts":{"x":"a","\u0078":"b"}}`, `{"unknown":{"x":0,"x":1}}`, `{"scripts":{"\ud800":"a"}}`, "{\"scripts\":{\"\xff\":\"a\"}}", `null`, `{"scripts":null}`, `[]`, strings.Repeat("[", 66) + strings.Repeat("]", 66)} {
		if _, err := parseNpm(context.Background(), []byte(b)); err == nil {
			t.Fatalf("accepted %q", b)
		}
	}
	p, e := parseNpm(context.Background(), []byte(`{"scripts":{" dev:wan ":"echo okay","a/b":"x","😀":"x","bad":null,"-n":"x"}}`))
	if e != nil {
		t.Fatal(e)
	}
	if len(p.members) != 5 || p.members[0].name != " dev:wan " || !p.members[0].valid {
		t.Fatal(p)
	}
	for _, m := range p.members {
		if (m.name == "bad" || m.name == "-n") && m.valid {
			t.Fatal(m)
		}
	}
}
func TestSimpleMakeGrammar(t *testing.T) {
	p, e := parseMake(context.Background(), []byte("all clean: dep\n\t@echo recipe\n.PHONY: all\n-n: x\na/b: x\nx=y: z\nx%: y\ninclude other\nifeq (x,y)\nconditional: x\nendif\n"), 1024)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.members) != 3 || len(p.notices) == 0 {
		t.Fatal(p)
	}
	for _, s := range []string{"", "-n", ".a", "a/b", "a=b", "$(X)", "a;b", "a b", "é"} {
		if safeTarget(s) {
			t.Fatal(s)
		}
	}
	if _, e = parseMake(context.Background(), []byte("long: "+strings.Repeat("x", 1024)), 1024); e == nil {
		t.Fatal("line cap")
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = parseMake(c, []byte("all:"), 1024); e != context.Canceled {
		t.Fatal(e)
	}
}
