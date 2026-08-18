import { polarToCart } from "../polar-convert";

export interface BufferLabelPos { row: number; label: string; px: number; py: number; }

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


export function poleAxis(phi: number, theta: number): [number, number, number] {
  return polarToCart(1, phi, theta);
}
