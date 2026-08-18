import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "../../src/webview/three/scene/nodes/node-frame-aggregate";
import { overlayFlag } from "../../src/webview/three/controls/flags/overlay-flags";
import { readNodeCX, readNodeCY, readNodeCZ } from "../../Buffer/buffer-layout";
import { DIRECTION_ZERO_EPS } from "../../src/webview/three/scene/buffer-scene-shared";

const CHANNEL_LINE_RADIUS = 0.5;
const CHANNEL_HEAD_RADIUS = 1.6;
const CHANNEL_HEAD_LENGTH = CHANNEL_HEAD_RADIUS * 2;
const CHANNEL_COLOR = "#7b6bd6";

const AXIS_DEFAULT = new THREE.Vector3(0, 1, 0);

export function RuleChannelLines({ capacity }: { capacity: number }) {
  const lineRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const mat = useRef(new THREE.Matrix4());
  const pos = useRef(new THREE.Vector3());
  const dir = useRef(new THREE.Vector3());
  const quat = useRef(new THREE.Quaternion());
  const scl = useRef(new THREE.Vector3());

  useFrame(() => {
    const line = lineRef.current;
    const head = headRef.current;
    if (!line || !head) return;

    if (!overlayFlag("ruleChannels")) {
      line.count = 0;
      head.count = 0;
      return;
    }

    const decoded = getNodeFrame();
    if (!decoded) {
      line.count = 0;
      head.count = 0;
      return;
    }
    const { nodeCount, nodeView } = decoded;

    let drawn = 0;
    for (let a = 0; a < nodeCount && drawn < capacity; a++) {
      const ax = readNodeCX(nodeView, a);
      const ay = readNodeCY(nodeView, a);
      const az = readNodeCZ(nodeView, a);
      for (let b = a + 1; b < nodeCount && drawn < capacity; b++) {
        const bx = readNodeCX(nodeView, b);
        const by = readNodeCY(nodeView, b);
        const bz = readNodeCZ(nodeView, b);

        dir.current.set(bx - ax, by - ay, bz - az);
        const len = dir.current.length();
        if (len <= DIRECTION_ZERO_EPS) continue;
        dir.current.divideScalar(len);
        quat.current.setFromUnitVectors(AXIS_DEFAULT, dir.current);

        pos.current.set(ax, ay, az).addScaledVector(dir.current, len / 2);
        scl.current.set(1, len, 1);
        mat.current.compose(pos.current, quat.current, scl.current);
        line.setMatrixAt(drawn, mat.current);

        pos.current.set(bx, by, bz).addScaledVector(dir.current, -CHANNEL_HEAD_LENGTH);
        scl.current.set(1, 1, 1);
        mat.current.compose(pos.current, quat.current, scl.current);
        head.setMatrixAt(drawn * 2, mat.current);

        quat.current.setFromUnitVectors(AXIS_DEFAULT, dir.current.clone().negate());
        pos.current.set(ax, ay, az).addScaledVector(dir.current, CHANNEL_HEAD_LENGTH);
        mat.current.compose(pos.current, quat.current, scl.current);
        head.setMatrixAt(drawn * 2 + 1, mat.current);

        drawn++;
      }
    }

    line.count = drawn;
    head.count = drawn * 2;
    line.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    if (drawn > 0) {
      line.computeBoundingSphere();
      head.computeBoundingSphere();
    }
  });

  return (
    <>
      <instancedMesh ref={lineRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[CHANNEL_LINE_RADIUS, CHANNEL_LINE_RADIUS, 1, 6]} />
        <meshBasicMaterial color={CHANNEL_COLOR} toneMapped={false} transparent opacity={0.45} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity * 2]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[CHANNEL_HEAD_RADIUS, CHANNEL_HEAD_LENGTH, 8]} />
        <meshBasicMaterial color={CHANNEL_COLOR} toneMapped={false} transparent opacity={0.6} />
      </instancedMesh>
    </>
  );
}
