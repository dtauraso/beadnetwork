package Wiring

import "testing"

// cascadeRelayClass is the summary the editor's drag log names the DRAGGED node's relay
// behavior from (Buffer Node.CascadeRelay → AbcDragLabel.tsx). It is a pure function of a
// kind string decided by ONE goroutine's own emit, so this is a plain per-kind table —
// no network, no second goroutine.
//
// The point of the table is that it stays honest about the rules it summarizes: every kind
// that the forwardDelta guard or the moveMsgKindDeltaForward handler stops must read
// terminus, and every kind that routes by sender kind must read routed. A kind added to
// either rule without a case here would silently keep reporting "fan" in the editor.
func TestCascadeRelayClass(t *testing.T) {
	for _, c := range []struct {
		kind string
		want uint8
	}{
		// terminus — never relays onward.
		{"TimeEnd", 2},    // stops in the moveMsgKindDeltaForward handler
		{"PulseLeft", 2},  // stops at the guard atop forwardDelta
		{"PulseRight", 2}, // same guard
		// routed — relays to a single target kind chosen by the SENDER's kind, or drops.
		{"Pulse", 1},     // gate crossover: SelectRight <-> SelectLeft
		{"TimeStart", 1}, // Pulse <-> Time
		// fan — relay to every cascade neighbor except the sender.
		{"Input", 0},
		{"Time", 0},
		{"SelectLeft", 0},
		{"SelectRight", 0},
		// An unknown kind string fans, matching forwardDelta's own fall-through: the
		// per-kind guards are opt-in, so a kind with no rule takes the plain fan.
		{"", 0},
	} {
		if got := cascadeRelayClass(c.kind); got != c.want {
			t.Errorf("cascadeRelayClass(%q) = %d, want %d", c.kind, got, c.want)
		}
	}
}
