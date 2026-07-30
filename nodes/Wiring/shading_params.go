// shading_params.go — single source of truth for the scene's shading PARAMETERS
// (glass/material params, environment/lighting params, wire-tube and bead
// appearance) shared between Go and the TS visual layer.
//
// Substance vs. medium (MODEL.md, docs/go-authoritative-clock/index.html "tsgo"):
// the GPU render machinery (three.js materials, PMREMGenerator, mesh creation,
// env-map baking, binding) stays in TS — Go has no GPU. What lives here is the
// shading PARAMETER DATA: Go owns it authoritatively; TS reads these values and
// applies them to GPU materials / bakes the env map from them, sourcing no
// shading values of its own. Per-node fill/stroke already live in NODE_DEFS
// (generated from each nodes/<Kind>/SPEC.md ## View) and stay there; this file
// holds the scene-global shading constants that were previously hardcoded in
// tools/topology-vscode/src/webview/three/scene-content.tsx.
//
// Codegen: tools/gen-node-defs reads this file and emits
// tools/topology-vscode/src/schema/shading-params.ts. Constants are prefixed
// with ShadingParam so gen-node-defs can identify them via the name prefix
// (same mechanism as CurveParam* → curve-params.ts). After changing any
// constant here, regenerate with:
//   cd tools/topology-vscode && npm run gen:node-defs
//
// Colors are hex strings (consumed by THREE.Color); scalars are float/int.

package Wiring

import ()

// --- Node body: glass (MeshPhysicalMaterial) parameters -------------------
// The node sphere is rendered as transmissive glass. These mirror the
// meshPhysicalMaterial props on GraphNode in scene-content.tsx exactly.

// ShadingParamNodeTransmission is the glass transmission (1 = fully transmissive).
const ShadingParamNodeTransmission = 1.0

// ShadingParamNodeThickness is the refraction thickness of the node glass.
const ShadingParamNodeThickness = 0.0

// ShadingParamNodeRoughness is the surface roughness of the node glass.
const ShadingParamNodeRoughness = 0.12

// ShadingParamNodeIor is the index of refraction of the node glass.
const ShadingParamNodeIor = 1.5

// ShadingParamNodeMetalness is the metalness of the node glass (0 = dielectric).
const ShadingParamNodeMetalness = 0.0

// ShadingParamNodeClearcoat is the clearcoat layer strength of the node glass.
const ShadingParamNodeClearcoat = 0.0

// ShadingParamNodeClearcoatRoughness is the clearcoat roughness of the node glass.
const ShadingParamNodeClearcoatRoughness = 0.1

// ShadingParamNodeEnvMapIntensity scales the baked env-map reflection on the node glass.
const ShadingParamNodeEnvMapIntensity = 1.0

// ShadingParamNodeOpacity is the node-body opacity.
const ShadingParamNodeOpacity = 0.92

// --- Node ring: border torus (NodeInstances) -------------------------------
// The border ring drawn around each node body (meshStandardMaterial torus).
// Mirrors the roughness on the ring material in NodeInstances.tsx exactly.

// ShadingParamRingRoughness is the surface roughness of the node border ring.
const ShadingParamRingRoughness = 0.6

// --- Procedural environment map (ProceduralEnvProvider) -------------------
// A tiny gradient-sky scene baked into a PMREM env texture. The env-map vertex
// tint is interpolated between a top color and a bottom color over the sky
// hemisphere; these RGB components mirror the per-channel literals in
// ProceduralEnvProvider exactly (kept as components because the bake lerps them).

// ShadingParamEnvSkyTopR/G/B is the top-of-sky tint (cool neutral).
const (
	ShadingParamEnvSkyTopR = 0.78
	ShadingParamEnvSkyTopG = 0.77
	ShadingParamEnvSkyTopB = 0.74
)

// ShadingParamEnvSkyBottomR/G/B is the horizon tint (warm cream).
const (
	ShadingParamEnvSkyBottomR = 1.0
	ShadingParamEnvSkyBottomG = 0.88
	ShadingParamEnvSkyBottomB = 0.75
)

// ShadingParamEnvSkyRadius is the radius of the baked sky hemisphere.
const ShadingParamEnvSkyRadius = 50.0

// ShadingParamEnvAmbientColor / Intensity is the soft white fill light baked into the env.
const (
	ShadingParamEnvAmbientColor     = "#ffffff"
	ShadingParamEnvAmbientIntensity = 0.9
)

// ShadingParamEnvKeyColor / Intensity is the warm key directional light baked into the env.
const (
	ShadingParamEnvKeyColor     = "#ffeedd"
	ShadingParamEnvKeyIntensity = 0.45
)

// ShadingParamEnvRimColor / Intensity is the cool rim directional light baked into the env.
const (
	ShadingParamEnvRimColor     = "#aabbff"
	ShadingParamEnvRimIntensity = 0.3
)

// ShadingParamEnvPmremBlur is the PMREMGenerator.fromScene blur (sigma) applied
// when baking the env texture.
const ShadingParamEnvPmremBlur = 0.04

// --- Scene lights (Scene component) ---------------------------------------
// The two direct scene lights (separate from the baked env). Mirror the
// <ambientLight> / <directionalLight> in the Scene component.

// ShadingParamSceneAmbientIntensity is the scene ambient-light intensity.
const ShadingParamSceneAmbientIntensity = 0.6

// ShadingParamSceneDirIntensity is the scene directional-light intensity.
const ShadingParamSceneDirIntensity = 0.8

// --- Wire tube appearance (SingleEdgeTube) --------------------------------
// The always-lit base tube material. Mirrors the meshStandardMaterial on the
// base tube in SingleEdgeTube.

// ShadingParamTubeColor is the wire-tube base color.
const ShadingParamTubeColor = "#5599cc"

// ShadingParamTubeEmissive is the wire-tube emissive color.
const ShadingParamTubeEmissive = "#2255aa"

// ShadingParamTubeEmissiveIntensity is the wire-tube emissive intensity.
const ShadingParamTubeEmissiveIntensity = 0.8

// --- Bead appearance (PulseBead) ------------------------------------------
// The in-flight bead sphere. Mirrors the meshStandardMaterial on PulseBead.

// ShadingParamBeadRadius is the sphere radius of a bead — the 0/1 beads AND the grey chain
// beads, which are the same size and structure by design (a chain bead is a grey version of
// bead 1, not a smaller marker). Mirrored into TS as SHADING_PARAM_BEAD_RADIUS, and read
// Go-side by chain_beads.go to space the chain at exactly one DIAMETER so adjacent beads
// TOUCH: a chain is a solid line of beads, not a dotted one, so there is no gap to tune and
// no second copy of the radius to drift.
const ShadingParamBeadRadius = 4.0

// ShadingParamBeadRingTubeRatio is a bead ring's torus tube radius as a fraction of
// ShadingParamBeadRadius. Same for chain beads as for the 0/1 beads — same structure.
const ShadingParamBeadRingTubeRatio = 0.12

// ShadingParamChainBeadFill is the UNLIT chain bead's fill — a pale cyan, DELIBERATELY not
// ShadingParamTubeColor below.
//
// This constant existed, was deleted on the reasoning that the chain IS the edge visual so a
// second colour would just be a copy free to drift, and is now back because the divergence is
// INTENTIONAL: David picked this tone off a screenshot. Recording that here so the next reader
// does not "fix" it back to the tube colour on the same argument that deleted it.
//
// The value started as the tone sampled from a screenshot (#a7dfe5) and was then taken down
// ~8% by eye because that read too bright in place. It is a chosen appearance, not a
// measurement to be restored. It is NOT a node's source
// fill: the node body that tone came from is glassy transmission material, so matching the
// appearance means matching the rendered tone rather than the input that produced it.
//
// Because this is a RENDERED tone, the bead that wears it is drawn with an UNLIT material
// (ChainBeadInstances.tsx). A lit material would multiply it by incoming light and render it a
// second time — measured at ~0.8x, which is why an earlier attempt with meshStandardMaterial
// came out #8daaad against this #a7dfe5. Change this constant and the pixel changes with it;
// that is only true while the material stays unlit.
const ShadingParamChainBeadFill = "#9acdd3"

// ShadingParamInteriorBeadFill0 and ShadingParamInteriorBeadFill1 are the fills for a bead
// HELD INSIDE a node (InteriorBeadInstances.tsx), deliberately kept SEPARATE from
// bead-style.ts's on-wire 0/1 fills even though they start at the same values. The reason
// is the same shape as ShadingParamChainBeadFill above: an interior bead is seen THROUGH
// the node's glassy transmissive meshPhysicalMaterial shell (NodeInstances.tsx), so its
// pixel is tinted by that shell no matter what material the bead itself uses. A shared
// constant with the on-wire bead cannot make the two look equal, because the interior one
// always has one more optical stage (the shell) sitting in front of it. Equality between
// an interior bead and a wire bead is therefore achieved by AUTHORING these two numbers BY
// EYE against the rendered shell, not by sharing a material or a constant. They start at
// the on-wire values (#000000 / #ffffff) so this commit changes only the MATERIAL an
// interior bead is drawn with (see InteriorBeadInstances.tsx), not its tone — expect these
// two constants to be tuned away from the on-wire values once the shell's tint is visible
// against them.
// PROBE: deliberately absurd values, committed only to answer one question — "does
// InteriorBeadInstances render at all in the live editor?" MUST be reverted to
// #000000/#ffffff before this branch merges.
const ShadingParamInteriorBeadFill0 = "#ff00ff"
const ShadingParamInteriorBeadFill1 = "#ff8800"

// ShadingParamBeadColor is the in-flight bead color.
const ShadingParamBeadColor = "#ffffff"

// ShadingParamBeadEmissive is the in-flight bead emissive color.
const ShadingParamBeadEmissive = "#ffffff"

// ShadingParamBeadEmissiveIntensity is the in-flight bead emissive intensity.
const ShadingParamBeadEmissiveIntensity = 2.5

// --- Layout-link overlay (cyan cascade-link overlay) ------------------------
// The second tube + arrowheads drawn over each cascade-linked LAYOUT pair
// (LayoutLink block), plus the dimmed opacity applied to the real edge tube
// underneath while the overlay is on.

// ShadingParamLayoutLinkColor is the layout-link overlay line/arrowhead color (cyan accent).
const ShadingParamLayoutLinkColor = "#00e5ff"

// ShadingParamLayoutLinkEmissive is the layout-link overlay emissive color.
const ShadingParamLayoutLinkEmissive = "#00e5ff"

// ShadingParamLayoutLinkEmissiveIntensity is the layout-link overlay emissive intensity.
const ShadingParamLayoutLinkEmissiveIntensity = 0.8
