// PolarVectors.tsx — the polarVectors overlay's "emphasise the two polar vectors" half
// (CLAUDE.md/MODEL.md's one-toggle-three-effects overlay; NodeInstances.tsx and
// ChainBeadInstances.tsx carry the other two effects, fading the nodes and the traversal
// animation). Draws exactly the two vectors MODEL.md's "the polar model" names:
//
//   - scene sphere centre --vector--> node   (the node's ONE polar coordinate), one per node.
//   - node --vector--> that edge's FIRST BEAD, one per edge.
//
// GEOMETRY SOURCE: no new domain state. The scene-sphere centre and every node's world
// centre are already streamed (Scene/Node blocks, read the same way SphereRings.tsx and
// buffer-nav.ts's sceneSphereFromSnapshot already do). A bead's centre is
// centre->node + node->bead (chain_beads.go's own doc comment), and an edge's SEGMENT START
// (SX,SY,SZ, Edge block) is anchored to that same first bead's position (chain_beads.go:
// "the same two points chain_beads.go anchors bead 0 and the last bead to") — so the edge's
// own already-streamed SX,SY,SZ IS the first-bead point, with no new column. The Edge block
// carries no source-node-row column (a port stopped being a place,
// docs/channels-not-ports.md), so the source node is resolved as the node whose CENTER is
// nearest SX,SY,SZ (a surface point at exactly that node's own radius from its center) — a
// decode-time geometric lookup over already-decoded points, not new domain state.
//
// Pure buffer→GPU, no state authority: this component reads the buffer every frame and
// draws lines between already-decoded points, imperatively (same timing contract as
// EdgeTube.tsx's LayoutLinkOverlay — a per-slot imperative handle, not React state, so a
// vector tracks its dragged endpoint on the same frame as the node it's attached to).

import React, { useRef, useState, useEffect, forwardRef, useImperativeHandle } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getViewBlocks } from "./view-blocks";
import { getNodeFrame } from "./node-stream-blocks";
import { getEdgeStreamAccessor } from "./edge-stream-blocks";
import { sceneSphereFromSnapshot } from "./buffer-nav";
import { readOverlayPolarVectors, readNodeCX, readNodeCY, readNodeCZ } from "../../schema/buffer-layout";
import {
  SHADING_PARAM_POLAR_VECTOR_COLOR,
  SHADING_PARAM_POLAR_VECTOR_EMISSIVE_INTENSITY,
  SHADING_PARAM_POLAR_VECTOR_TUBE_RADIUS,
  SHADING_PARAM_POLAR_VECTOR_BEAD_VECTOR_COLOR,
  SHADING_PARAM_POLAR_VECTOR_BEAD_VECTOR_EMISSIVE_INTENSITY,
  SHADING_PARAM_POLAR_VECTOR_ARROWHEAD_LENGTH,
  SHADING_PARAM_POLAR_VECTOR_ARROWHEAD_RADIUS,
} from "../../schema/shading-params";
import { DIRECTION_ZERO_EPS } from "./buffer-scene-shared";

interface VecSeg { sx: number; sy: number; sz: number; ex: number; ey: number; ez: number; }
interface VecHandle { update(seg: VecSeg): void }

function sameSeg(a: VecSeg, b: VecSeg): boolean {
  return a.sx === b.sx && a.sy === b.sy && a.sz === b.sz
    && a.ex === b.ex && a.ey === b.ey && a.ez === b.ez;
}

/** Cone apex at `apex`, pointing (apex-ward) along `dir` (normalized) — same construction
 *  EdgeTube.tsx's buildArrow uses for the cascade-link overlay's arrowheads. */
function buildArrow(apex: THREE.Vector3, dir: THREE.Vector3, height: number): {
  center: THREE.Vector3; q: THREE.Quaternion;
} {
  const q = new THREE.Quaternion().setFromUnitVectors(new THREE.Vector3(0, 1, 0), dir);
  const center = apex.clone().addScaledVector(dir, -height / 2);
  return { center, q };
}

/** One drawn vector: a tube from start to end plus a single arrowhead at `end` (unlike the
 *  cascade-link overlay's bidirectional pair — a polar vector has one direction, sceneCentre
 *  ->node or node->bead, so it gets one head, not two). */
const VectorOverlay = forwardRef<VecHandle, { color: string; emissiveIntensity: number; tubeRadius: number }>(
  function VectorOverlay({ color, emissiveIntensity, tubeRadius }, ref) {
    const lineMeshRef = useRef<THREE.Mesh>(null);
    const arrowRef = useRef<THREE.Mesh>(null);
    const lastSeg = useRef<VecSeg | null>(null);
    const geoRef = useRef<THREE.TubeGeometry | null>(null);
    const emissiveColor = useRef(new THREE.Color(color));

    useImperativeHandle(ref, () => ({
      update(seg: VecSeg) {
        if (lastSeg.current && sameSeg(lastSeg.current, seg)) return;
        lastSeg.current = seg;

        const start = new THREE.Vector3(seg.sx, seg.sy, seg.sz);
        const end = new THREE.Vector3(seg.ex, seg.ey, seg.ez);
        const curve = new THREE.LineCurve3(start, end);
        const lineGeo = new THREE.TubeGeometry(curve, 1, tubeRadius, 8, false);
        if (geoRef.current) geoRef.current.dispose();
        geoRef.current = lineGeo;
        if (lineMeshRef.current) lineMeshRef.current.geometry = lineGeo;

        const dir = end.clone().sub(start);
        const a = arrowRef.current;
        if (dir.length() >= DIRECTION_ZERO_EPS) {
          const dirNorm = dir.clone().normalize();
          if (a) {
            const { center, q } = buildArrow(end, dirNorm, SHADING_PARAM_POLAR_VECTOR_ARROWHEAD_LENGTH);
            a.position.set(center.x, center.y, center.z);
            a.quaternion.set(q.x, q.y, q.z, q.w);
            a.visible = true;
          }
        } else if (a) {
          a.visible = false;
        }
      },
    }), [tubeRadius]);

    useEffect(() => () => { if (geoRef.current) geoRef.current.dispose(); }, []);

    return (
      <>
        <mesh ref={lineMeshRef} raycast={() => null} frustumCulled={false}>
          <meshStandardMaterial color={color} emissive={emissiveColor.current} emissiveIntensity={emissiveIntensity} />
        </mesh>
        <mesh ref={arrowRef} raycast={() => null} frustumCulled={false} visible={false}>
          <coneGeometry args={[SHADING_PARAM_POLAR_VECTOR_ARROWHEAD_RADIUS, SHADING_PARAM_POLAR_VECTOR_ARROWHEAD_LENGTH, 16]} />
          <meshStandardMaterial color={color} emissive={emissiveColor.current} emissiveIntensity={emissiveIntensity} />
        </mesh>
      </>
    );
  },
);

/** Nearest node row to world point (px,py,pz) by CENTER distance — resolves an edge's
 *  segment-start surface point (Edge block SX,SY,SZ) back to its SOURCE node, since the Edge
 *  block carries no node-row column of its own (docs/channels-not-ports.md). A surface point
 *  sits exactly that node's own radius from its center, closer than any other node's center
 *  in any topology this overlay is meaningful for. */
function nearestNodeRow(nodeView: DataView, nodeCount: number, px: number, py: number, pz: number): number {
  let bestRow = -1;
  let bestDist = Infinity;
  for (let row = 0; row < nodeCount; row++) {
    const dx = readNodeCX(nodeView, row) - px;
    const dy = readNodeCY(nodeView, row) - py;
    const dz = readNodeCZ(nodeView, row) - pz;
    const d = dx * dx + dy * dy + dz * dz;
    if (d < bestDist) { bestDist = d; bestRow = row; }
  }
  return bestRow;
}

export function PolarVectors() {
  const [visible, setVisible] = useState(false);
  const [nodeCount, setNodeCount] = useState(0);
  const [edgeCount, setEdgeCount] = useState(0);
  const nodeHandles = useRef<(VecHandle | null)[]>([]);
  const edgeHandles = useRef<(VecHandle | null)[]>([]);

  useFrame(() => {
    const blocks = getViewBlocks();
    const decodedNode = getNodeFrame();
    if (!blocks || !decodedNode) { if (visible) setVisible(false); return; }

    const on = readOverlayPolarVectors(blocks.overlayView) !== 0;
    if (on !== visible) setVisible(on);
    if (!on) return;

    const { nodeCount: n, nodeView } = decodedNode;
    if (n !== nodeCount) setNodeCount(n);

    const scene = sceneSphereFromSnapshot(blocks);
    for (let row = 0; row < n; row++) {
      const seg: VecSeg = {
        sx: scene.center.x, sy: scene.center.y, sz: scene.center.z,
        ex: readNodeCX(nodeView, row), ey: readNodeCY(nodeView, row), ez: readNodeCZ(nodeView, row),
      };
      nodeHandles.current[row]?.update(seg);
    }

    const edges = getEdgeStreamAccessor();
    const eCount = edges?.edgeCount ?? 0;
    if (eCount !== edgeCount) setEdgeCount(eCount);
    if (edges) {
      for (let row = 0; row < eCount; row++) {
        const [sx, sy, sz] = edges.segment(row);
        const srcRow = nearestNodeRow(nodeView, n, sx, sy, sz);
        if (srcRow < 0) continue;
        const seg: VecSeg = {
          sx: readNodeCX(nodeView, srcRow), sy: readNodeCY(nodeView, srcRow), sz: readNodeCZ(nodeView, srcRow),
          ex: sx, ey: sy, ez: sz,
        };
        edgeHandles.current[row]?.update(seg);
      }
    }
  });

  if (!visible) return null;

  return (
    <>
      {Array.from({ length: nodeCount }, (_, i) => (
        <VectorOverlay
          key={`polar-vec-node-${i}`}
          ref={(h) => { nodeHandles.current[i] = h; }}
          color={SHADING_PARAM_POLAR_VECTOR_COLOR}
          emissiveIntensity={SHADING_PARAM_POLAR_VECTOR_EMISSIVE_INTENSITY}
          tubeRadius={SHADING_PARAM_POLAR_VECTOR_TUBE_RADIUS}
        />
      ))}
      {Array.from({ length: edgeCount }, (_, i) => (
        <VectorOverlay
          key={`polar-vec-bead-${i}`}
          ref={(h) => { edgeHandles.current[i] = h; }}
          color={SHADING_PARAM_POLAR_VECTOR_BEAD_VECTOR_COLOR}
          emissiveIntensity={SHADING_PARAM_POLAR_VECTOR_BEAD_VECTOR_EMISSIVE_INTENSITY}
          tubeRadius={SHADING_PARAM_POLAR_VECTOR_TUBE_RADIUS}
        />
      ))}
    </>
  );
}
