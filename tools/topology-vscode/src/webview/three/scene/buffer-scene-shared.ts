import { readNodeKindId, UNKNOWN_KIND_ID } from "../../../../../../Buffer/buffer-layout";
import { NODE_DEFS_ARRAY } from "../../../schema/node-defs";
import { polarToCart } from "../polar-convert";

export interface BufferLabelPos { row: number; label: string; px: number; py: number; cx: number; cy: number; }

export const BUFFER_NODE_TAG = "bufferNode";

export const BUFFER_RING_TAG = "bufferRing";

export const BUFFER_EDGE_TAG = "bufferEdgeRow";

export const NODE_SPHERE_RADIUS = 12;

export const NODE_RING_TUBE_RATIO = 0.08;

export const RING_PICK_TUBE_RATIO = NODE_RING_TUBE_RATIO;

export const RING_PICK_COLOR = "#00e5a8";

export const RING_PICK_OPACITY = 0.9;

export const RING_BAND_MAJOR = 1 + NODE_RING_TUBE_RATIO * 1.6;
export const RING_BAND_TUBE = NODE_RING_TUBE_RATIO * 0.275;

export const HOVER_COLOR = "#aaddff";
export const HOVER_RING_TUBE_RATIO = 0.14;

export const DIRECTION_ZERO_EPS = 1e-6;

const NODE_DEFAULT_FILL = "#ffffff";
const NODE_DEFAULT_STROKE = "#888888";

export function nodeRowColors(nodeView: DataView, row: number): { fill: string; stroke: string } {
  const kindId = readNodeKindId(nodeView, row);
  const def = kindId === UNKNOWN_KIND_ID ? undefined : NODE_DEFS_ARRAY[kindId];
  return {
    fill: def?.fill ?? NODE_DEFAULT_FILL,
    stroke: def?.stroke ?? NODE_DEFAULT_STROKE,
  };
}

export function poleAxis(phi: number, theta: number): [number, number, number] {
  return polarToCart(1, phi, theta);
}
