package main

import (
	"os"
	"path/filepath"
	"testing"
)

// --- trace_kinds.go: parseTraceKinds / parseBreadcrumbLabels --------------

func TestParseTraceKinds_OrderMatchesSourceDeclarationOrder(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

const (
	KindRecv = "recv"
	KindFire = "fire"
	KindSend = "send"
)

const KindSlot = "slot"

// Not a Kind* const — must be ignored.
const OtherThing = "other"
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	kinds, err := parseTraceKinds(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"recv", "fire", "send", "slot"}
	if len(kinds) != len(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Errorf("kinds[%d] = %q, want %q (reorder bug: wire numbering IS slice order)", i, kinds[i], want[i])
		}
	}
}

func TestParseTraceKinds_IgnoresTestFiles(t *testing.T) {
	dir := t.TempDir()
	main := `package Trace

const KindReal = "real"
`
	testFile := `package Trace

const KindFromTest = "fromtest"
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(main), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "trace_test.go"), []byte(testFile), 0644); err != nil {
		t.Fatal(err)
	}
	kinds, err := parseTraceKinds(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kinds) != 1 || kinds[0] != "real" {
		t.Fatalf("kinds = %v, want only [\"real\"] (test file constants must not leak in)", kinds)
	}
}

func TestParseTraceKinds_EmptyIsError(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

const NotAKindConst = "x"
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseTraceKinds(dir); err == nil {
		t.Fatal("want error when no Kind* constants are found, got nil")
	}
}

func TestParseBreadcrumbLabels_OrderPreserved(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

var BreadcrumbLabels = []string{"first", "second", "third"}
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	labels, err := parseBreadcrumbLabels(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"first", "second", "third"}
	if len(labels) != len(want) {
		t.Fatalf("labels = %v, want %v", labels, want)
	}
	for i := range want {
		if labels[i] != want[i] {
			t.Errorf("labels[%d] = %q, want %q", i, labels[i], want[i])
		}
	}
}

func TestParseBreadcrumbLabels_NotFoundIsError(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

var SomethingElse = []string{"x"}
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseBreadcrumbLabels(dir); err == nil {
		t.Fatal("want error when BreadcrumbLabels var is absent, got nil")
	}
}
