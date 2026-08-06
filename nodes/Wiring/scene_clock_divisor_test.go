package Wiring

// scene_clock_divisor_test.go — the PAIR scene's clock runs SLOWER than the ring's at the
// same user-chosen speed, an entirely GO-OWNED scaling (never crosses the bridge, never
// touches view/speed.json). Tests here cover: the shared arithmetic (EffectiveClockSpeed),
// unknown-scene resolution (SceneClockDivisor), and that a save/reload cycle never compounds
// the divisor — the on-disk value stays the USER's unscaled speed.
//
// Nothing here pins the pair's divisor to a LITERAL. It is a watched-it tuning number (see
// SceneTab.ClockDivisor) and has already moved once; a test asserting the literal fails on
// every tuning change while testing nothing about the behaviour. What is asserted instead is
// the CONSTRAINT — the ring is unscaled, the pair is slower than the ring, an unknown scene
// is unscaled — and the arithmetic is exercised against whatever the table currently says.

import (
	"path/filepath"
	"testing"
)

// TestEffectiveClockSpeedRingAndPair: the ring's divisor is a no-op; the pair's divides the
// user's speed by whatever the SceneTabs table currently holds, across the fractional table
// the six-value slider actually sends (0, 0.25, 0.5, 0.75, 1, 2) plus 0 itself.
func TestEffectiveClockSpeedRingAndPair(t *testing.T) {
	pairDivisor := SceneClockDivisor(filepath.Join("/anywhere", "topology-pair"))
	cases := []float64{0, 0.25, 0.5, 0.75, 1, 2}
	for _, userSpeed := range cases {
		if got := EffectiveClockSpeed(userSpeed, 1); got != userSpeed {
			t.Fatalf("ring divisor=1: EffectiveClockSpeed(%v, 1) = %v, want %v (no scaling)", userSpeed, got, userSpeed)
		}
		want := userSpeed / pairDivisor
		if got := EffectiveClockSpeed(userSpeed, pairDivisor); got != want {
			t.Fatalf("pair divisor=%v: EffectiveClockSpeed(%v, %v) = %v, want %v", pairDivisor, userSpeed, pairDivisor, got, want)
		}
	}
}

// TestEffectiveClockSpeedGuardsInvalidDivisor: a 0 or negative divisor must never reach a
// division — EffectiveClockSpeed treats it as "no scaling" instead.
func TestEffectiveClockSpeedGuardsInvalidDivisor(t *testing.T) {
	for _, divisor := range []float64{0, -1, -4} {
		if got := EffectiveClockSpeed(1, divisor); got != 1 {
			t.Fatalf("EffectiveClockSpeed(1, %v) = %v, want 1 (guarded, no scaling)", divisor, got)
		}
	}
}

// TestSceneClockDivisorKnownScenes: the ring is unscaled and the pair is SLOWER than the
// ring. The pair's exact value is a tuning number and is deliberately not asserted — what
// must hold is the relationship (and that the value can never be one a division would choke
// on, which is the case the guard above covers from the other side).
func TestSceneClockDivisorKnownScenes(t *testing.T) {
	ring := SceneClockDivisor(filepath.Join("/anywhere", "topology"))
	if ring != 1 {
		t.Fatalf("ring SceneClockDivisor = %v, want 1 (the ring is the unscaled reference)", ring)
	}
	pair := SceneClockDivisor(filepath.Join("/anywhere", "topology-pair"))
	if pair <= ring {
		t.Fatalf("pair SceneClockDivisor = %v, want > the ring's %v — the pair scene is smaller and must run slower at the same slider setting", pair, ring)
	}
}

// TestSceneClockDivisorUnknownSceneIsOne: a non-tabbed / unrecognised tree (a test fixture,
// a one-off path) resolves to divisor 1, never 0 — a 0 divisor reaching EffectiveClockSpeed
// would be a latent division-by-zero even though the guard above also catches it.
func TestSceneClockDivisorUnknownSceneIsOne(t *testing.T) {
	got := SceneClockDivisor(filepath.Join("/anywhere", "some-one-off-fixture"))
	if got != 1 {
		t.Fatalf("unknown scene SceneClockDivisor = %v, want 1", got)
	}
}

// TestLoadSpeedDoesNotCompoundDivisorAcrossReload: seed at load, save a new user speed,
// reload — the on-disk value and md.ui.speed both stay the USER's unscaled number, not
// divided again on the second load. This is the compounding failure mode the divisor change
// could introduce: divisor is applied ONLY on the way to the speed CHANNELS
// (EffectiveClockSpeed), never to what LoadSpeed reads from or writes to disk.
func TestLoadSpeedDoesNotCompoundDivisorAcrossReload(t *testing.T) {
	root := writeTree(t) // unrecognised fixture path -> divisor 1, but the compounding bug
	// would show up at ANY divisor since it is about disk storage, not the arithmetic itself.
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)

	const userSpeed = 0.5
	md.persist.speed.schedule(userSpeed)

	onDisk, found := loadSceneSpeed(speedFilePath(root))
	if !found || onDisk != userSpeed {
		t.Fatalf("after first save: loadSceneSpeed = (%v, %v), want (%v, true)", onDisk, found, userSpeed)
	}

	// Reload once.
	first := loadTreeMD(t, root)
	first.LoadSpeed(root, nil, nil)
	if first.ui.speed != userSpeed {
		t.Fatalf("first LoadSpeed: ui.speed = %v, want unscaled %v", first.ui.speed, userSpeed)
	}

	// Reload again from the SAME file, simulating a second respawn (e.g. a tab switch and
	// back) — if LoadSpeed ever wrote a scaled value back to disk this second read would see
	// userSpeed/divisor instead of userSpeed.
	second := loadTreeMD(t, root)
	second.LoadSpeed(root, nil, nil)
	if second.ui.speed != userSpeed {
		t.Fatalf("second LoadSpeed (compounding check): ui.speed = %v, want unscaled %v", second.ui.speed, userSpeed)
	}
	onDiskAfter, found := loadSceneSpeed(speedFilePath(root))
	if !found || onDiskAfter != userSpeed {
		t.Fatalf("after reloads: on-disk speed = (%v, %v), want (%v, true) — divisor must never reach disk", onDiskAfter, found, userSpeed)
	}
}
