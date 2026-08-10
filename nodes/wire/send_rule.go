// send_rule.go — ONE JOB: the per-edge send POLICY vocabulary. The SendRule type,
// its closed set of constants, and the parse that turns a persisted string into
// one. It is deliberately separate from the port that carries it (out_port.go's
// Out.Rule): the rule is data the topology file authors, kept in parity with the
// TS SEND_RULES array (tools/bridge/check-send-rule-parity.sh), and the wire itself never
// consults it — the source node does.

package wire

import "fmt"

// SendRule names the per-edge send policy applied by the source node after a
// successful TrySend. The wire stays dumb transport; the node consults the rule.
type SendRule string

const (
	// RuleConsumeGated: default send rule. Kept for compatibility with persisted
	// topology JSON; the consume gate was removed (PacedWire.Done/WaitConsumed
	// are no-ops). The only meaningful distinction is Gated() which gates
	// optional feedback ports.
	RuleConsumeGated SendRule = "consumeGated"
	// RuleFireAndForget: the node sends and does not wait for consumption.
	RuleFireAndForget SendRule = "fireAndForget"
)

// ParseSendRule converts a raw string into a SendRule.
// An empty string returns RuleConsumeGated (preserving the default-when-absent
// behavior). Any string that is not a recognised constant returns an error.
func ParseSendRule(s string) (SendRule, error) {
	switch s {
	case "":
		return RuleConsumeGated, nil
	case string(RuleConsumeGated):
		return RuleConsumeGated, nil
	case string(RuleFireAndForget):
		return RuleFireAndForget, nil
	default:
		return RuleConsumeGated, fmt.Errorf("invalid sendRule %q: must be one of %q, %q",
			s, RuleConsumeGated, RuleFireAndForget)
	}
}
