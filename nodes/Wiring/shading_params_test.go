package Wiring

import (
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// Phase 4 verifier (docs/go-authoritative-clock/index.html, Verify row "4 ·
// Shading": "go test — shading params emitted"). Go owns the shading PARAMETER
// values; TS reads them via the generated shading-params.ts and binds them to
// GPU materials. This test asserts every shading param Go now owns is present
// and exactly the value the renderer must apply, so an accidental change to a
// material value can't slip through without turning this red (the deterministic
// half of the gate; pixel fidelity is the manual smoke-check).
//
// The expected values are written out independently here (not referencing the
// consts) so editing a const in shading_params.go without intent breaks this.

func TestShadingParamsFloat(t *testing.T) {
	cases := []struct {
		name string
		got  float64
		want float64
	}{
		// Node-body glass (MeshPhysicalMaterial).
		{"NodeTransmission", ShadingParamNodeTransmission, 1.0},
		{"NodeThickness", ShadingParamNodeThickness, 0.0},
		{"NodeRoughness", ShadingParamNodeRoughness, 0.12},
		{"NodeIor", ShadingParamNodeIor, 1.5},
		{"NodeMetalness", ShadingParamNodeMetalness, 0.0},
		{"NodeClearcoat", ShadingParamNodeClearcoat, 0.0},
		{"NodeClearcoatRoughness", ShadingParamNodeClearcoatRoughness, 0.1},
		{"NodeEnvMapIntensity", ShadingParamNodeEnvMapIntensity, 1.0},
		{"NodeOpacity", ShadingParamNodeOpacity, 0.92},
		// Procedural env map.
		{"EnvSkyTopR", ShadingParamEnvSkyTopR, 0.78},
		{"EnvSkyTopG", ShadingParamEnvSkyTopG, 0.77},
		{"EnvSkyTopB", ShadingParamEnvSkyTopB, 0.74},
		{"EnvSkyBottomR", ShadingParamEnvSkyBottomR, 1.0},
		{"EnvSkyBottomG", ShadingParamEnvSkyBottomG, 0.88},
		{"EnvSkyBottomB", ShadingParamEnvSkyBottomB, 0.75},
		{"EnvSkyRadius", ShadingParamEnvSkyRadius, 50.0},
		{"EnvAmbientIntensity", ShadingParamEnvAmbientIntensity, 0.9},
		{"EnvKeyIntensity", ShadingParamEnvKeyIntensity, 0.45},
		{"EnvRimIntensity", ShadingParamEnvRimIntensity, 0.3},
		{"EnvPmremBlur", ShadingParamEnvPmremBlur, 0.04},
		// Scene lights.
		{"SceneAmbientIntensity", ShadingParamSceneAmbientIntensity, 0.6},
		{"SceneDirIntensity", ShadingParamSceneDirIntensity, 0.8},
		// Wire tube + bead.
		{"TubeEmissiveIntensity", ShadingParamTubeEmissiveIntensity, 0.8},
		{"BeadEmissiveIntensity", ShadingParamBeadEmissiveIntensity, 2.5},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("ShadingParam%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// TestShadingParamNodeRingTubeRatioMatchesTS pins the Go mirror of the TS
// NODE_RING_TUBE_RATIO (buffer-scene-shared.ts) — see ShadingParamNodeRingTubeRatio's
// doc comment for why this now has to stay in sync (the node's torus outer radius is
// load-bearing chain-bead-tangency geometry, not decoration).
func TestShadingParamNodeRingTubeRatioMatchesTS(t *testing.T) {
	if ShadingParamNodeRingTubeRatio != 0.08 {
		t.Fatalf("ShadingParamNodeRingTubeRatio = %v, want 0.08 (TS NODE_RING_TUBE_RATIO)", ShadingParamNodeRingTubeRatio)
	}
}

// TestShadingParamBeadRadiusMatchesDerivation now merely restates ShadingParamBeadRadius's
// own definition: it IS `wire.BeadRadius` (docs/bead-lattice.md "The lattice is derived,
// not the bead") — the AUTHORED primitive, re-exported rather than duplicated as a literal
// — not a literal pinned to a formula by a separate test. This used to run the other
// direction (ShadingParamBeadRadius computed FROM wire.BeadTorusOuterR); the identity
// `wire.BeadTorusOuterR/(1+ratio) == wire.BeadRadius` still holds either way (it is the
// same tangency relationship, just read from the other end), which is why this second
// form is kept alongside the direct comparison below — both must agree with the primitive.
// Left in place as a cheap regression guard against someone reintroducing a hand-computed
// literal here (the failure mode this whole change closes off); the real regression
// coverage for the GENERATOR side — that a non-literal ShadingParam* expression still gets
// evaluated, not silently dropped — is TestParseShadingParams_EvaluatesCrossPackageExpression
// in tools/gen-node-defs.
func TestShadingParamBeadRadiusMatchesDerivation(t *testing.T) {
	if ShadingParamBeadRadius != wire.BeadRadius {
		t.Fatalf("ShadingParamBeadRadius = %v, want wire.BeadRadius = %v", ShadingParamBeadRadius, wire.BeadRadius)
	}
	// The inverse relationship (tangency) still holds, read from the other end.
	wantViaTangency := wire.BeadTorusOuterR / (1 + ShadingParamBeadRingTubeRatio)
	if ShadingParamBeadRadius != wantViaTangency {
		t.Fatalf("ShadingParamBeadRadius = %v, want %v (wire.BeadTorusOuterR / (1+ratio), the tangency identity)", ShadingParamBeadRadius, wantViaTangency)
	}
}

func TestShadingParamsColor(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"EnvAmbientColor", ShadingParamEnvAmbientColor, "#ffffff"},
		{"EnvKeyColor", ShadingParamEnvKeyColor, "#ffeedd"},
		{"EnvRimColor", ShadingParamEnvRimColor, "#aabbff"},
		{"TubeColor", ShadingParamTubeColor, "#5599cc"},
		{"TubeEmissive", ShadingParamTubeEmissive, "#2255aa"},
		{"BeadColor", ShadingParamBeadColor, "#ffffff"},
		{"BeadEmissive", ShadingParamBeadEmissive, "#ffffff"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("ShadingParam%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}
