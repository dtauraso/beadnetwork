package build_test

// scene_speed_persist_test.go — round-trip test for the playback-speed persister
// (view/speed.json): a clock/speed edit → the synchronous writer persists the multiplier to
// disk → a reload reads it back and re-seeds every clock-owning goroutine's speed channel.
// Per memory/feedback_headless_repro_verifies_persistence.md, this drives the REAL
// LoadSpeed/writeSceneSpeed/loadSceneSpeed functions against a real on-disk tree (writeTree/
// loadTreeMD, the same real-file harness scene_edit_persist_test.go uses for overlays),
// reading the actual bytes on disk rather than trusting an isolated in-memory stub.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewpersist"
)

// TestPersistSpeedRoundTrips: schedule a speed write -> speed.json carries the exact
// fractional multiplier -> a fresh LoadSpeed call restores md.UI.Speed from disk.
func TestPersistSpeedRoundTrips(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, root)

	md.Persist.Speed().Schedule(0.25)

	got, found := scenepersist.LoadSceneSpeed(scenepaths.SpeedFilePath(root))
	if !found {
		t.Fatalf("loadSceneSpeed found no speed key after flush")
	}
	if got != 0.25 {
		t.Fatalf("speed.json round-trip: got %v, want 0.25", got)
	}

	// A fresh dispatch's LoadSpeed restores md.UI.Speed from the same file — no speedSinks
	// needed for this assertion (nil slice, per LoadSpeed's own nil-tolerant loop).
	fresh := loadTreeMD(t, root)
	scenepersist.InstallSpeed(&fresh.UI, root, nil, nil)
	if fresh.UI.Speed != 0.25 {
		t.Fatalf("LoadSpeed did not restore ui.speed=0.25, got %v", fresh.UI.Speed)
	}
}

// TestLoadSpeedFallsBackQuietlyWhenMissing: a fresh topology with no speed.json falls back
// to the default multiplier (1) rather than failing loudly — this is a PREFERENCE, not a
// structural invariant like counts.json.
func TestLoadSpeedFallsBackQuietlyWhenMissing(t *testing.T) {
	root := writeTree(t) // no view/speed.json
	md := loadTreeMD(t, root)

	scenepersist.InstallSpeed(&md.UI, root, nil, nil)

	if md.UI.Speed != scenepersist.DefaultPlaybackSpeed {
		t.Fatalf("LoadSpeed with no file: got ui.speed=%v, want default %v", md.UI.Speed, scenepersist.DefaultPlaybackSpeed)
	}
}

// TestLoadSpeedSeedsEverySpeedSink: LoadSpeed broadcasts the loaded value to every
// clock-owning goroutine's own speed channel, the SAME Delivery path a live slider edit
// uses (clockAttrHandlers's "speed" case) — this is what makes a persisted speed actually
// take effect on the clocks a fresh process starts, not just on md.UI.Speed.
func TestLoadSpeedSeedsEverySpeedSink(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, root)
	md.Persist.Speed().Schedule(0.5)

	fresh := loadTreeMD(t, root)
	ch := make(chan float64, 1)
	scenepersist.InstallSpeed(&fresh.UI, root, []chan float64{ch}, nil)

	select {
	case got := <-ch:
		if got != 0.5 {
			t.Fatalf("speedSinks channel got %v, want 0.5", got)
		}
	default:
		t.Fatalf("LoadSpeed did not broadcast the loaded speed onto speedSinks")
	}
}
