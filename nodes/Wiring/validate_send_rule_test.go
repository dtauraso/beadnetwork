// validate_send_rule_test.go — validateSpec's sendRules/duplicate-id checks. Split out
// of ports_test.go when ports.go moved to nodes/wire (that file's SendRule-specific
// tests, TestOutGated/TestParseSendRule, stayed in nodes/wire; these two exercise
// validateSpec/topoSpec, which are still Wiring's own loader machinery).

package Wiring

import "testing"

// TestValidateSpecSendRule verifies that validateSpec rejects invalid sendRules
// values and accepts valid/absent ones.
func TestValidateSpecSendRule(t *testing.T) {
	// helper: minimal valid spec with one node of a known kind.
	// We'll use the "Input" kind which exists in all topologies.
	// Instead of depending on a real kind, build a spec that only exercises
	// the sendRules check (Check 4); other checks may emit their own errors
	// but we only care about the sendRules error being present/absent.

	// Invalid sendRule value: expect an error containing the bad value.
	badSpec := &topoSpec{
		Nodes: []specNode{
			{
				ID:   "n1",
				Type: "", // unknown kind — Check 1 fires but we still reach Check 4
				Data: &NodeData{
					SendRules: map[string]string{
						"out": "typo_value",
					},
				},
			},
		},
	}
	err := validateSpec(badSpec)
	if err == nil {
		t.Fatal("validateSpec with bad sendRule: expected error, got nil")
	}
	if !containsStr(err.Error(), "typo_value") {
		t.Errorf("error should name the bad value; got: %v", err)
	}

	// Valid sendRule value: error should NOT mention the sendRules key.
	goodSpec := &topoSpec{
		Nodes: []specNode{
			{
				ID:   "n1",
				Type: "",
				Data: &NodeData{
					SendRules: map[string]string{
						"out": "fireAndForget",
					},
				},
			},
		},
	}
	err = validateSpec(goodSpec)
	// May have errors for unknown kind etc., but NOT for sendRules.
	if err != nil && containsStr(err.Error(), "sendRule") {
		t.Errorf("valid sendRule should not produce a sendRule error; got: %v", err)
	}

	// Missing sendRules map: no sendRules error.
	emptySpec := &topoSpec{
		Nodes: []specNode{
			{ID: "n1", Type: "", Data: &NodeData{}},
		},
	}
	err = validateSpec(emptySpec)
	if err != nil && containsStr(err.Error(), "sendRule") {
		t.Errorf("absent sendRules should not produce a sendRule error; got: %v", err)
	}
}

// TestValidateSpecDuplicateNodeID verifies validateSpec rejects two nodes sharing
// an id (which would otherwise silently last-wins the kind map).
func TestValidateSpecDuplicateNodeID(t *testing.T) {
	dup := &topoSpec{
		Nodes: []specNode{
			{ID: "n1", Type: "", Data: &NodeData{}},
			{ID: "n1", Type: "", Data: &NodeData{}},
		},
	}
	err := validateSpec(dup)
	if err == nil {
		t.Fatal("validateSpec with duplicate node id: expected error, got nil")
	}
	if !containsStr(err.Error(), "duplicate node id") {
		t.Errorf("error should flag the duplicate id; got: %v", err)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
