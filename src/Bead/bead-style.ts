import {
  SHADING_PARAM_INTERIOR_BEAD_FILL0,
  SHADING_PARAM_INTERIOR_BEAD_FILL1,
  SHADING_PARAM_EDGE_LINE_COLOR,
} from "../schema/buffer-layout/shading-params";

const VALUE_BEAD_STYLE: Record<number, { fill: string; ring: string }> = {
  0: { fill: "#000000", ring: "#000000" },
  1: { fill: "#ffffff", ring: "#000000" },
};

export function beadStyleForValue(v: number | null | undefined): { fill: string; ring: string } | undefined {
  return v == null ? undefined : VALUE_BEAD_STYLE[v];
}

export const EDGE_LINE_COLOR = SHADING_PARAM_EDGE_LINE_COLOR;

export const INSTANCE_TINT_BASE = "#ffffff";

const INTERIOR_VALUE_BEAD_STYLE: Record<number, { fill: string; ring: string }> = {
  0: { fill: SHADING_PARAM_INTERIOR_BEAD_FILL0, ring: "#000000" },
  1: { fill: SHADING_PARAM_INTERIOR_BEAD_FILL1, ring: "#000000" },
};
export function interiorBeadStyleForValue(v: number | null | undefined): { fill: string; ring: string } | undefined {
  return v == null ? undefined : INTERIOR_VALUE_BEAD_STYLE[v];
}
