// driven_out.go — DrivenOut is the STRUCTURAL fix for the framing desync documented in
// docs/interior-stream-framing.md: a *wire.Out obtained by a.Out(...)/a.Broadcast(...) is
// written by this node's own Update-loop goroutine (via its shared getStream), while a
// *wire.Out obtained by a.DriveOut(...) is meant to be written by a SEPARATE
// gatecommon.DriveHeld goroutine, on its OWN dedicated per-(node,slot) drive-stream fd
// (newDriveStreamGetter, Buffer.StreamKindDrive). Before this type existed,
// gatecommon.DriveHeld accepted a bare *wire.Out, so nothing stopped a kind from handing it
// the SHARED Out returned by a.Out(...) — exactly the mistake that produced the original
// two-goroutines-one-fd desync, caught afterward only by grepping source text
// (tools/check-driveheld-uses-driveout.sh).
//
// DrivenOut makes that mistake a COMPILE ERROR instead: gatecommon.DriveHeld's signature
// now accepts ONLY Wiring.DrivenOut, never *wire.Out, and the ONLY way to produce a
// DrivenOut is this file's unexported newDrivenOut constructor, called from exactly one
// place — BuildArgs.DriveOut (build_args.go, same package). No package outside
// nodes/Wiring can construct one: DrivenOut's only field is unexported, and Go allows
// neither a struct literal naming an unexported field nor any other construction path from
// a different package (short of the reflect/unsafe packages, which are a different kind of
// violation entirely and outside this guard's scope, same as for any other Go encapsulation
// in this codebase). So `a.Out("Out")` and `a.DriveOut("Out", 0)` are now two INCOMPATIBLE
// types at the call site that hands a value to gatecommon.DriveHeld — passing the former
// where the latter is required fails `go build`, before any goroutine exists, not at
// runtime under load in a live editor session.
//
// DrivenOut deliberately exposes only the THREE operations gatecommon.DriveHeld actually
// needs (Steps, PlaceDrivenAt, Paced) plus Wired (used by callers like Pulse's optional
// OutFanout check) — not an Unwrap()/Out() method back to the bare *wire.Out. That keeps it
// a narrow capability instead of a transparent alias a caller could stash and hand to a
// second goroutine under a different name.
package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// DrivenOut wraps a *wire.Out that has been routed through a dedicated per-DriveHeld-
// goroutine stream fd (BuildArgs.DriveOut) — see this file's header comment for why this
// type exists and what it structurally prevents.
type DrivenOut struct {
	out *wire.Out
}

// newDrivenOut is unexported ON PURPOSE — see this file's header comment. Its only
// PRODUCTION caller is BuildArgs.DriveOut, below in build_args.go (same package).
func newDrivenOut(out *wire.Out) DrivenOut { return DrivenOut{out: out} }

// NewDrivenOutForTest is the one deliberate, clearly-named escape hatch: unit tests that
// build a node kind's own struct directly (bypassing BuildArgs/LoadTopology entirely, the
// same "construct a bare wire.Out by hand" pattern nodes/wire's own NewOutChanForTest and
// NewPacedOutNoGeom already use) need SOME way to populate a Wiring.DrivenOut field without
// going through a real loader. Production code has no reason to call this — the guard this
// file's header comment describes (a plain a.Out(...) result cannot become a DrivenOut) is
// about node-kind SOURCE, not test harnesses, matching the old
// check-driveheld-uses-driveout.sh's own "SCOPE: nodes/*/node.go only... not _test.go
// files" carve-out.
func NewDrivenOutForTest(out *wire.Out) DrivenOut { return DrivenOut{out: out} }

// Wired reports whether the underlying Out is actually connected to a wire — mirrors
// *wire.Out.Wired(), used by kinds with an optional driven output (Pulse's OutFanout).
func (d DrivenOut) Wired() bool { return d.out.Wired() }

// Paced reports whether the underlying Out is in paced (wire) mode vs chan (test) mode —
// gatecommon.DriveHeld's placement strategy depends on this.
func (d DrivenOut) Paced() bool { return d.out.Paced() }

// Steps returns the underlying Out's current bead-step count (its Geom().Steps) — the one
// field gatecommon.DriveHeld's placement pacing needs; narrower than exposing Geom()
// itself, whose return type (wire.outGeom) is unexported and cannot be named outside
// package wire anyway.
func (d DrivenOut) Steps() int { return d.out.Geom().Steps }

// PlaceDrivenAt delegates to the underlying Out's PlaceDrivenAt — see its doc comment
// (nodes/wire/ports.go). This is the one write operation a DriveHeld goroutine performs.
func (d DrivenOut) PlaceDrivenAt(v int, tick int64) wire.DriveItem {
	return d.out.PlaceDrivenAt(v, tick)
}
