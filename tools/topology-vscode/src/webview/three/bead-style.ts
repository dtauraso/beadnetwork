import {
  SHADING_PARAM_INTERIOR_BEAD_FILL0,
  SHADING_PARAM_INTERIOR_BEAD_FILL1,
  SHADING_PARAM_CHAIN_BEAD_FILL,
} from "../../schema/shading-params";

// bead-style.ts — Single source of truth for bead value → appearance.
// On-wire beads (chain traversal) and interior (held-inside-a-node) beads each get their
// own registry below, even though both start at the same 0/1 tones. This file is still the
// SINGLE place either kind reads from, so a caller cannot invent a third colour.

// Single source of truth for on-wire value→appearance. The animated edge/chain bead derives
// its fill/ring colors here. (The former static data.init bead components were removed
// when node 1's interior switched to the live node-bead stream.)
const VALUE_BEAD_STYLE: Record<number, { fill: string; ring: string }> = {
  0: { fill: "#000000", ring: "#000000" },
  1: { fill: "#ffffff", ring: "#000000" },
};
// Only 0 and 1 are valid bead values. A value outside the map (including a
// missing/undefined value) returns undefined — the caller hides the bead rather
// than drawing a grey/fake fallback. With Go no longer placing -1 on a wire, a
// non-0/1 bead is a bug, not a colour to paint.
export function beadStyleForValue(v: number | null | undefined): { fill: string; ring: string } | undefined {
  return v == null ? undefined : VALUE_BEAD_STYLE[v];
}

// The drawn edge's colour (EdgeLines.tsx): the BEAD SPHERE's own fill — Go's
// ShadingParamChainBeadFill, the pale cyan the chain beads wear. Not the value-1 white: the
// lit tones (black/white) belong to a bead CARRYING a value, and painting the resting edge
// with one of them would say the whole edge is holding a 1.
//
// It is a RENDERED tone chosen off a screenshot, so the material that wears it must stay
// unlit — a lit material multiplies it by incoming light and renders it a second time
// (~0.8x, measured; see that constant's own doc comment in nodes/Wiring/shading_params.go).
// Verified by probe (a magenta value reached the screen), so this constant IS what the line
// draws with — if the line reads differently from the beads it is not the value that
// differs. It is the same fill on a different SHAPE: a chain bead is a sphere wearing a
// BLACK RING (RING_COLOR, ChainBeadInstances), and the ring darkens the whole chain's
// apparent tone; a bare cylinder has nothing doing that, so the identical hex looks lighter.
export const EDGE_LINE_COLOR = SHADING_PARAM_CHAIN_BEAD_FILL;

// Interior (held-inside-a-node) value→appearance. This is a SEPARATE registry from
// VALUE_BEAD_STYLE above, not a reuse of it: an interior bead is seen THROUGH the node's
// glassy transmissive shell (NodeInstances.tsx), which tints whatever is behind it, so the
// on-wire fill and the interior fill can never be allowed to drift together automatically —
// making them equal on screen requires authoring the interior fill separately against the
// rendered shell. Fills come from Go's ShadingParamInteriorBeadFill0/1 (mirrored into TS as
// SHADING_PARAM_INTERIOR_BEAD_FILL0/1); the ring stays the same black ring as the on-wire
// styles, since the ring is not part of what the shell tints away.
const INTERIOR_VALUE_BEAD_STYLE: Record<number, { fill: string; ring: string }> = {
  0: { fill: SHADING_PARAM_INTERIOR_BEAD_FILL0, ring: "#000000" },
  1: { fill: SHADING_PARAM_INTERIOR_BEAD_FILL1, ring: "#000000" },
};
export function interiorBeadStyleForValue(v: number | null | undefined): { fill: string; ring: string } | undefined {
  return v == null ? undefined : INTERIOR_VALUE_BEAD_STYLE[v];
}
