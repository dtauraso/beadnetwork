package viewstate

import (
	"strings"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// overlay_state_test.go — independent behavior check for the generated overlay
// Toggle wiring in overlay_state.go (moved verbatim from Wiring's overlay_gen_test.go, per
// docs/planning/gesture-actor.md's lift). This is the trust frontier of the generator: it
// exercises the mechanical name-derivation (flag → field / method) that replaced the
// hand-written methods. For EACH overlay flag it asserts that Toggle flips the owned
// OverlayState bool. The RowEvent this toggle implies is written by the caller (Wiring's
// stdin_dispatch.go's applyUpdate), not by OverlayState itself, so it is out of scope here.
// The exception flags (scene/node poles breadcrumb) are covered explicitly.

// overlayCase describes one generated overlay flag by its behavior, using closures
// so the test reads/writes the concrete OverlayState field without reflection.
type overlayCase struct {
	name   string
	get    func(*OverlayState) bool
	toggle func(*OverlayState, *T.Trace)
	crumb  string // non-empty => breadcrumb node arg expected on Toggle (poles)
}

var overlayCases = []overlayCase{
	{
		name:   "tori",
		get:    func(o *OverlayState) bool { return o.SceneToriVisible },
		toggle: (*OverlayState).ToggleSceneTori,
	},
	{
		name:   "scenePoles",
		get:    func(o *OverlayState) bool { return o.ScenePolesVisible },
		toggle: (*OverlayState).ToggleScenePoles,
		crumb:  "scene",
	},
	{
		name:   "nodePoles",
		get:    func(o *OverlayState) bool { return o.NodePolesVisible },
		toggle: (*OverlayState).ToggleNodePoles,
		crumb:  "nodes",
	},
	{
		name:   "selSpherePoles",
		get:    func(o *OverlayState) bool { return o.SelSpherePolesVisible },
		toggle: (*OverlayState).ToggleSelSpherePoles,
	},
	{
		name:   "handholds",
		get:    func(o *OverlayState) bool { return o.HandholdsVisible },
		toggle: (*OverlayState).ToggleHandholds,
	},
	{
		name:   "labelsGlobal",
		get:    func(o *OverlayState) bool { return o.LabelsGlobalVisible },
		toggle: (*OverlayState).ToggleLabelsGlobal,
	},
	{
		name:   "overlays",
		get:    func(o *OverlayState) bool { return o.OverlaysVisible },
		toggle: (*OverlayState).ToggleOverlaysVis,
	},
}

func TestOverlayToggleFlips(t *testing.T) {
	for _, c := range overlayCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			// Start from a known false, so a flip lands on true.
			var o OverlayState
			var dbg strings.Builder
			tr := T.New()
			tr.SetSink(&dbg)
			c.toggle(&o, tr)

			if !c.get(&o) {
				t.Fatalf("%s: Toggle did not flip field false->true", c.name)
			}
			if c.crumb != "" {
				s := dbg.String()
				if !strings.Contains(s, "pole-toggle-go") || !strings.Contains(s, `"node":"`+c.crumb+`"`) {
					t.Fatalf("%s: Toggle did not emit breadcrumb for scope %q; sink=%q", c.name, c.crumb, s)
				}
			}
		})
	}
}

// TestDefaultOverlayState pins the startup snapshot: all 7 flags default ON.
func TestDefaultOverlayState(t *testing.T) {
	d := DefaultOverlayState()
	on := []bool{
		d.SceneToriVisible, d.ScenePolesVisible, d.NodePolesVisible,
		d.SelSpherePolesVisible, d.HandholdsVisible,
		d.LabelsGlobalVisible, d.OverlaysVisible,
	}
	for i, v := range on {
		if !v {
			t.Fatalf("DefaultOverlayState field #%d should default ON", i)
		}
	}
}
