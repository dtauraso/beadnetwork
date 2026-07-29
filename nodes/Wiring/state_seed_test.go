package Wiring

import "testing"

// state_seed_test.go — locks in: a `data.state` seed (e.g. Held on Time/HoldFlip/Pacer) is
// OPTIONAL. When the spec omits the key, the seed accessor must leave the kind's own
// default untouched (the empty sentinel, NoValue = -1 for held-bearing kinds) rather than
// returning Go's int zero-value 0 — 0 is a legitimate held bead value, so defaulting to 0
// would emit a phantom bead. Only a key PRESENT in data.state overrides the default.
//
// Was populate_data_held_seed_test.go, which called the reflection-era populateData with a
// tagged fixture struct. That function is deleted; the same rule now lives in
// BuildArgs.StateSeed, which each kind calls explicitly. Same three cases, same intent —
// only the seam moved.
func TestStateSeedIsOptional(t *testing.T) {
	cases := []struct {
		name  string
		state map[string]int
		want  int
	}{
		{"absent key keeps the kind's own default (NoValue)", nil, NoValue},
		{"authored held:0 is honored, not treated as unset", map[string]int{"held": 0}, 0},
		{"authored held:-1 stays -1", map[string]int{"held": -1}, -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := BuildArgs{data: &NodeData{State: c.state}}
			if got := a.StateSeed("held", NoValue); got != c.want {
				t.Fatalf("StateSeed(\"held\", NoValue) = %d, want %d", got, c.want)
			}
		})
	}
}

// TestStateSeedNoDataBlock: a spec node with no data block at all must also keep the
// kind's default — the nil-data path is how most nodes are authored.
func TestStateSeedNoDataBlock(t *testing.T) {
	a := BuildArgs{}
	if got := a.StateSeed("held", NoValue); got != NoValue {
		t.Fatalf("StateSeed with no data block = %d, want %d", got, NoValue)
	}
}
