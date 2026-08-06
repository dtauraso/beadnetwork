package Wiring

// scene_clock_divisor_test.go — the PAIR scene's clock runs 4x slower than the ring's at
// the same user-chosen speed, an entirely GO-OWNED scaling (never crosses the bridge, never
// touches view/speed.json). Tests here cover: the shared arithmetic (EffectiveClockSpeed),
// unknown-scene resolution (SceneClockDivisor), and that a save/reload cycle never compounds
// the divisor — the on-disk value stays the USER's unscaled speed.

import (
	"path/filepath"
	"testing"
)

// TestEffectiveClockSpeedRingAndPair: the ring's divisor (1) is a no-op; the pair's (4)
// quarters the user's speed, across the fractional table the six-value slider actually sends
// (0, 0.25, 0.5, 0.75, 1, 2) plus 0 itself.
func TestEffectiveClockSpeedRingAndPair(t *testing.T) {
	cases := []float64{0, 0.25, 0.5, 0.75, 1, 2}
	for _, userSpeed := range cases {
		if got := EffectiveClockSpeed(userSpeed, 1); got != userSpeed {
			t.Fatalf("ring divisor=1: EffectiveClockSpeed(%v, 1) = %v, want %v (no scaling)", userSpeed, got, userSpeed)
		}
		want := userSpeed / 4
		if got := EffectiveClockSpeed(userSpeed, 4); got != want {
			t.Fatalf("pair divisor=4: EffectiveClockSpeed(%v, 4) = %v, want %v", userSpeed, got, want)
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

// TestSceneClockDivisorKnownScenes: the ring resolves to 1, the pair to 4, matching the
// SceneTabs table directly (a change to that table should move this test, not be silently
// tolerated).
func TestSceneClockDivisorKnownScenes(t *testing.T) {
	ring := SceneClockDivisor(filepath.Join("/anywhere", "topology"))
	if ring != 1 {
		t.Fatalf("ring SceneClockDivisor = %v, want 1", ring)
	}
	pair := SceneClockDivisor(filepath.Join("/anywhere", "topology-pair"))
	if pair != 4 {
		t.Fatalf("pair SceneClockDivisor = %v, want 4", pair)
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
