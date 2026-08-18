import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius, readNodeSelected,
} from "../Buffer/buffer-layout";
import { NODE_SPHERE_RADIUS } from "../src/webview/three/scene/buffer-scene-shared";

// The one reader of where a node is and how big it is.
//
// Several things are drawn AT a node's frame -- its body and rings, its highlights, its
// pole frame -- and each used to read these columns itself. None of them derives
// anything: they place a mesh at the centre and scale it by the radius. So the fix is
// not a per-consumer column from Go, which would ship the same three numbers again under
// another name; it is that they ask here. A consumer that wants to COMPUTE a position
// from a node's centre is a different case and belongs in Go, which is where the label
// anchor, the tilt arrows, the interior beads and the channel vectors went.

export function nodeCenterX(nodeView: DataView, row: number): number {
  return readNodeCX(nodeView, row);
}

export function nodeCenterY(nodeView: DataView, row: number): number {
  return readNodeCY(nodeView, row);
}

export function nodeCenterZ(nodeView: DataView, row: number): number {
  return readNodeCZ(nodeView, row);
}

export function nodeRadius(nodeView: DataView, row: number): number {
  return readNodeRadius(nodeView, row) || NODE_SPHERE_RADIUS;
}

// Radius exactly as streamed, with no default filled in, for the one reader that wants
// to know the column is empty rather than be handed a stand-in.
export function nodeRadiusRaw(nodeView: DataView, row: number): number {
  return readNodeRadius(nodeView, row);
}

export function nodeSelected(nodeView: DataView, row: number): boolean {
  return readNodeSelected(nodeView, row) !== 0;
}
