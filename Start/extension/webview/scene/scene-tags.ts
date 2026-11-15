import { polarToCart } from "../../../../Categories/Vectors/polar-convert";

export interface LabelPos { row: number; label: string; px: number; py: number; }

export const SCENE_NODE_TAG = "sceneNode";

export const SCENE_RING_TAG = "sceneRing";

export const SCENE_EDGE_TAG = "sceneEdgeRow";

export const NODE_SPHERE_RADIUS = 12;

export const NODE_RING_TUBE_RATIO = 0.08;

export const HOVER_COLOR = "#aaddff";
export const HOVER_RING_TUBE_RATIO = 0.14;

export const DIRECTION_ZERO_EPS = 1e-6;


export function poleAxis(phi: number, theta: number): [number, number, number] {
  return polarToCart(1, phi, theta);
}
