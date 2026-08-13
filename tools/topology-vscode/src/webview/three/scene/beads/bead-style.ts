import {
  SHADING_PARAM_INTERIOR_BEAD_FILL0,
  SHADING_PARAM_INTERIOR_BEAD_FILL1,
  SHADING_PARAM_EDGE_LINE_COLOR,
} from "../../../../schema/buffer-layout/shading-params";

const VALUE_BEAD_STYLE: Record<number, { fill: string; ring: string }> = {
  0: { fill: "#000000", ring: "#000000" },
  1: { fill: "#ffffff", ring: "#000000" },
};

export function beadStyleForValue(v: number | null | undefined): { fill: string; ring: string } | undefined {
  return v == null ? undefined : VALUE_BEAD_STYLE[v];
}

export const EDGE_LINE_COLOR = SHADING_PARAM_EDGE_LINE_COLOR;

// COMM_BEAD_STYLE is the whole chain along a path that carries a position
// rather than a value — drawn at the same size as the animation edge it
// stands in for, so the only difference the eye reads is the colour.
export const COMM_BEAD_STYLE = { fill: "#3fb950", ring: "#14532d" };

const INTERIOR_VALUE_BEAD_STYLE: Record<number, { fill: string; ring: string }> = {
  0: { fill: SHADING_PARAM_INTERIOR_BEAD_FILL0, ring: "#000000" },
  1: { fill: SHADING_PARAM_INTERIOR_BEAD_FILL1, ring: "#000000" },
};
export function interiorBeadStyleForValue(v: number | null | undefined): { fill: string; ring: string } | undefined {
  return v == null ? undefined : INTERIOR_VALUE_BEAD_STYLE[v];
}
