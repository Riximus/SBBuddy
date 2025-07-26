package main

import (
	"flag"
	"testing"
)

type multiFlagTest []string

func (m *multiFlagTest) String() string     { return "" }
func (m *multiFlagTest) Set(v string) error { *m = append(*m, v); return nil }

func TestConnectionsFlagParsing(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var conns multiFlagTest
	fs.Var(&conns, "C", "")
	if err := fs.Parse([]string{"-C", "Basel SBB", "-C", "Zürich HB"}); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(conns) != 2 || conns[0] != "Basel SBB" || conns[1] != "Zürich HB" {
		t.Errorf("unexpected parsed values: %v", conns)
	}
}

func TestConnectionsFlagParsingMultiple(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var conns multiFlagTest
	fs.Var(&conns, "C", "")
	args := []string{"-C", "Basel", "-C", "Bern", "-C", "Zürich"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(conns) != 3 || conns[1] != "Bern" {
		t.Errorf("unexpected parsed values: %v", conns)
	}
}

func TestArrivalFlagParsing(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	arrival := fs.Bool("a", false, "")
	if err := fs.Parse([]string{"-a"}); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !*arrival {
		t.Errorf("arrival flag should be true when provided")
	}
}

func TestRandomFlagParsing(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	random := fs.Int("R", 0, "")
	if err := fs.Parse([]string{"-R", "2"}); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if *random != 2 {
		t.Errorf("expected 2 got %d", *random)
	}
}

func TestRandomFlagDefaultParsing(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	random := fs.Int("R", 0, "")
	args := insertDefaultRandom([]string{"-R"})
	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if *random != 0 {
		t.Errorf("expected 0 got %d", *random)
	}
}
