package jsonpersist

import "testing"

// safe_tree_path_test.go — verifies SafeTreePathComponent rejects path-traversal values.
// Split out of nodes/Wiring/scene_path_safety_test.go: SafeTreePathComponent lives here, in
// jsonpersist, and the write-side test that needs Wiring's own unexported writeQuantOffset
// stayed behind in nodes/Wiring/scene_path_safety_test.go.
func TestSafeTreePathComponent(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"", false},
		{".", false},
		{"..", false},
		{"../x", false},
		{"a/b", false},
		{"/abs", false},
		{`a\b`, false},
		{"n1", true},
		{"portOut", true},
		{"node_2", true},
	}
	for _, c := range cases {
		if got := SafeTreePathComponent(c.s); got != c.want {
			t.Errorf("SafeTreePathComponent(%q) = %v, want %v", c.s, got, c.want)
		}
	}
}
