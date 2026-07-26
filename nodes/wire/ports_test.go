package wire

import (
	"testing"
)

// TestOutGated verifies the node-owned send rule controls Gated():
// the zero value and consumeGated are gated; fireAndForget is not.
func TestOutGated(t *testing.T) {
	cases := []struct {
		name     string
		rule     SendRule
		wantGate bool
	}{
		{"zero value defaults gated", "", true},
		{"consumeGated is gated", RuleConsumeGated, true},
		{"fireAndForget is not gated", RuleFireAndForget, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := &Out{Rule: tc.rule}
			if got := o.Gated(); got != tc.wantGate {
				t.Fatalf("Out{Rule:%q}.Gated() = %v, want %v", tc.rule, got, tc.wantGate)
			}
		})
	}

	// Nil-safe: a nil Out is gated.
	var nilOut *Out
	if !nilOut.Gated() {
		t.Fatalf("(*Out)(nil).Gated() = false, want true")
	}
}

// TestParseSendRule verifies ParseSendRule accepts valid inputs and rejects typos.
func TestParseSendRule(t *testing.T) {
	okCases := []struct {
		input string
		want  SendRule
	}{
		{"", RuleConsumeGated},
		{"consumeGated", RuleConsumeGated},
		{"fireAndForget", RuleFireAndForget},
	}
	for _, tc := range okCases {
		got, err := ParseSendRule(tc.input)
		if err != nil {
			t.Errorf("ParseSendRule(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseSendRule(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}

	badCases := []string{"consumegated", "fireandforget", "typo", "ConsumeGated", "FireAndForget"}
	for _, raw := range badCases {
		if _, err := ParseSendRule(raw); err == nil {
			t.Errorf("ParseSendRule(%q): expected error, got nil", raw)
		}
	}
}
