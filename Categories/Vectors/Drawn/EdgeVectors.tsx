import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getEdgeStreamAccessor } from "../../Node/Edge/edge-stream-blocks";
import { checkEdgeLandsOnNode } from "../../Node/Edge/check-edge-lands-on-node";
import { EDGE_LINE_COLOR, INSTANCE_TINT_BASE } from "../../Ring/Bead/bead-style";

import { DIRECTION_ZERO_EPS } from "../../../Start/extension/webview/scene/scene-tags";

const EDGE_LINE_RADIUS = 1.5;
const ARROW_HEAD_RADIUS = 3;
const ARROW_HEAD_LENGTH = ARROW_HEAD_RADIUS * 2;

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

    checkEdgeLandsOnNode();

    const edges = getEdgeStreamAccessor();
    if (!edges) { line.count = 0; head.count = 0; return; }

    const n = Math.min(edges.edgeCount, capacity);
    let drawn = 0;
    for (let row = 0; row < n; row++) {
      const [sx, sy, sz, ex, ey, ez] = edges.segment(row);
      dir.current.set(ex - sx, ey - sy, ez - sz);
      const len = dir.current.length();

      if (len <= DIRECTION_ZERO_EPS) continue;
      dir.current.divideScalar(len);
      quat.current.setFromUnitVectors(AXIS_DEFAULT, dir.current);

      const shaft = Math.max(len - ARROW_HEAD_LENGTH, 0);
      pos.current.set(sx, sy, sz).addScaledVector(dir.current, shaft / 2);
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
        <cylinderGeometry args={[EDGE_LINE_RADIUS, EDGE_LINE_RADIUS, 1, 8]} />
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
