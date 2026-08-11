package lattice

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

func TestBeadFractionClampsAtZeroAndOne(t *testing.T) {
	cases := []struct {
		name                               string
		nowTick, placementTick, crossTicks float64
		want                               float64
	}{
		{"before placement clamps to 0", 5, 10, 20, 0},
		{"at placement is 0", 10, 10, 20, 0},
		{"midway is fractional", 20, 10, 20, 0.5},
		{"at deadline is 1", 30, 10, 20, 1},
		{"past deadline clamps to 1", 100, 10, 20, 1},
		{"zero crossTicks returns 0", 10, 10, 0, 0},
		{"negative crossTicks returns 0", 10, 10, -5, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := BeadFraction(c.nowTick, c.placementTick, c.crossTicks)
			if got != c.want {
				t.Fatalf("BeadFraction(%v,%v,%v) = %v, want %v", c.nowTick, c.placementTick, c.crossTicks, got, c.want)
			}
		})
	}
}

func TestSimLatencyMsScalesWithSteps(t *testing.T) {
	zero := SimLatencyMs(0)
	if zero != 0 {
		t.Fatalf("SimLatencyMs(0) = %v, want 0", zero)
	}
	one := SimLatencyMs(1)
	want := DwellTicksPerBead * clock.MsPerTick
	if one != want {
		t.Fatalf("SimLatencyMs(1) = %v, want %v", one, want)
	}
	ten := SimLatencyMs(10)
	if ten != 10*one {
		t.Fatalf("SimLatencyMs(10) = %v, want %v (10x SimLatencyMs(1))", ten, 10*one)
	}
}
