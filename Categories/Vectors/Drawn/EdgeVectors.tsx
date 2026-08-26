import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getEdgeStreamAccessor } from "../../Node/Edge/edge-stream-blocks";
import { nodeF32 } from "../../Node/node-leaves";
import { checkEdgeLandsOnNode } from "../../Node/Edge/check-edge-lands-on-node";
import { EDGE_LINE_COLOR, INSTANCE_TINT_BASE } from "../../Ring/Bead/bead-style";
import { overlayFlag } from "../../Scene/View/Flags/overlay-flags";

import { DIRECTION_ZERO_EPS } from "../../../Start/extension/webview/scene/scene-tags";

const EDGE_LINE_RADIUS = 1.5;
const EDGE_LINE_TIP_RADIUS = 0.2;
const ARROW_HEAD_RADIUS = 3;
const ARROW_HEAD_LENGTH = ARROW_HEAD_RADIUS * 2;
const ARROW_HEAD_GAP = 1;

const AXIS_DEFAULT = new THREE.Vector3(0, 1, 0);


export function EdgeVectors({ capacity }: { capacity: number }) {
  const lineRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const mat = useRef(new THREE.Matrix4());
  const pos = useRef(new THREE.Vector3());
  const dir = useRef(new THREE.Vector3());
  const quat = useRef(new THREE.Quaternion());
  const scl = useRef(new THREE.Vector3());
  const col = useRef(new THREE.Color());

  useFrame(() => {
    const line = lineRef.current;
    const head = headRef.current;
    if (!line || !head) return;

    if (!overlayFlag("edgeVectors")) { line.count = 0; head.count = 0; return; }

    checkEdgeLandsOnNode();

    const edges = getEdgeStreamAccessor();
    if (!edges) { line.count = 0; head.count = 0; return; }

    const n = Math.min(edges.edgeCount, capacity);
    let drawn = 0;
    for (let row = 0; row < n; row++) {
      const [bx, by, bz, cx, cy, cz] = edges.segment(row);

      const src = edges.srcNodeRow(row);
      const dst = edges.dstNodeRow(row);
      const sx = src >= 0 ? nodeF32(src, "poleAnchorX") : bx;
      const sy = src >= 0 ? nodeF32(src, "poleAnchorY") : by;
      const sz = src >= 0 ? nodeF32(src, "poleAnchorZ") : bz;
      const ex = dst >= 0 ? nodeF32(dst, "poleAnchorX") : cx;
      const ey = dst >= 0 ? nodeF32(dst, "poleAnchorY") : cy;
      const ez = dst >= 0 ? nodeF32(dst, "poleAnchorZ") : cz;

      dir.current.set(ex - sx, ey - sy, ez - sz);
      const len = dir.current.length();

      if (len <= DIRECTION_ZERO_EPS) continue;
      dir.current.divideScalar(len);
      quat.current.setFromUnitVectors(AXIS_DEFAULT, dir.current);

      const tail = ARROW_HEAD_LENGTH + ARROW_HEAD_GAP;
      const shaft = Math.max(len - ARROW_HEAD_LENGTH - ARROW_HEAD_GAP - tail, 0);

      pos.current.set(sx, sy, sz).addScaledVector(dir.current, tail + shaft / 2);
      scl.current.set(1, shaft, 1);
      mat.current.compose(pos.current, quat.current, scl.current);
      line.setMatrixAt(drawn, mat.current);

      pos.current.set(ex, ey, ez).addScaledVector(dir.current, -ARROW_HEAD_LENGTH / 2);
      scl.current.set(1, 1, 1);
      mat.current.compose(pos.current, quat.current, scl.current);
      head.setMatrixAt(drawn, mat.current);

      col.current.set(EDGE_LINE_COLOR);
      line.setColorAt(drawn, col.current);
      head.setColorAt(drawn, col.current);

      drawn++;
    }
    line.count = drawn;
    head.count = drawn;
    line.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    if (line.instanceColor) line.instanceColor.needsUpdate = true;
    if (head.instanceColor) head.instanceColor.needsUpdate = true;
    if (drawn > 0) {
      line.computeBoundingSphere();
      head.computeBoundingSphere();
    }
  });

  return (
    <>
      {}
      <instancedMesh ref={lineRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[EDGE_LINE_TIP_RADIUS, EDGE_LINE_RADIUS, 1, 8]} />
        {}
        {}
        <meshBasicMaterial color={INSTANCE_TINT_BASE} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[ARROW_HEAD_RADIUS, ARROW_HEAD_LENGTH, 12]} />
        {}
        {}
        <meshBasicMaterial color={INSTANCE_TINT_BASE} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      {}
    </>
  );
}
