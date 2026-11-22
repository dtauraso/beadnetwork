import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getEdgeStreamAccessor } from "../../Node/Edge/edge-stream-blocks";
import { nodeF32 } from "../../Node/Block/read/node-leaves";
import { checkEdgeLandsOnNode } from "../../Node/Edge/check-edge-lands-on-node";
import { EDGE_LINE_COLOR, INSTANCE_TINT_BASE } from "../../Ring/Bead/bead-style";
import { overlayFlag } from "../../Scene/View/Flags/overlay-flags";
import { previewBytes, previewI32, previewI32Run, previewU8 } from "../../Scene/Preview/preview-leaves";

import { DIRECTION_ZERO_EPS, SCENE_EDGE_TAG } from "../../../Start/extension/webview/scene/scene-tags";

const EDGE_LINE_RADIUS = 1.5;
const ARROW_HEAD_RADIUS = 3;
const ARROW_HEAD_LENGTH = ARROW_HEAD_RADIUS * 2;

const AXIS_DEFAULT = new THREE.Vector3(0, 1, 0);
const HOT_WIDTH = 2.5;
const HOT_COLOR = "#ffd54a";
const EDGE_UNPICKED = 1, EDGE_PICKED = 2, EDGE_OPEN = 3;
const STATE_STYLE: Record<number, { color: string; width: number }> = {
  [EDGE_UNPICKED]: { color: "#d5d8dd", width: 0.001 },
  [EDGE_PICKED]: { color: "#a7dfe5", width: 1.5 },
  [EDGE_OPEN]: { color: "#1565c0", width: 2 },
};
const DASH_COLOR = "#d5d8dd";
const DASH_WIDTH = 0.7;
const DASHES = 8;
const noRaycast = () => {};

function edgeStates(): Map<number, number> {
  const out = new Map<number, number>();
  const src = previewI32Run("edgeSrc"), dst = previewI32Run("edgeDst"), st = previewBytes("edgeState");
  if (!src || !dst || !st) return out;
  for (let i = 0; i < src.length && i < dst.length && i < st.byteLength; i++) {
    const s = st.getUint8(i);
    out.set(src[i]! * 65536 + dst[i]!, s);
    out.set(dst[i]! * 65536 + src[i]!, s);
  }
  return out;
}


export function EdgeVectors({ capacity }: { capacity: number }) {
  const lineRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const dashRef = useRef<THREE.InstancedMesh>(null);
  const mat = useRef(new THREE.Matrix4());
  const pos = useRef(new THREE.Vector3());
  const dir = useRef(new THREE.Vector3());
  const quat = useRef(new THREE.Quaternion());
  const scl = useRef(new THREE.Vector3());
  const col = useRef(new THREE.Color());
  const rows = useRef<number[]>([]);
  const dashRows = useRef<number[]>([]);

  useFrame(() => {
    const line = lineRef.current;
    const head = headRef.current;
    const dash = dashRef.current;
    if (!line || !head || !dash) return;

    if (!overlayFlag("edgeVectors")) { line.count = 0; head.count = 0; dash.count = 0; return; }

    checkEdgeLandsOnNode();

    const edges = getEdgeStreamAccessor();
    if (!edges) { line.count = 0; head.count = 0; dash.count = 0; return; }
    const states = edgeStates();
    let dashes = 0;

    const n = Math.min(edges.edgeCount, capacity);
    const hotSrc = previewI32("hotSrc", -1);
    const hotDst = previewI32("hotDst", -1);
    const pick = previewU8("edgePick") === 1;
    line.raycast = pick ? (rc, out) => { THREE.InstancedMesh.prototype.raycast.call(line, rc, out); } : noRaycast;
    dash.raycast = pick ? (rc, out) => { THREE.InstancedMesh.prototype.raycast.call(dash, rc, out); } : noRaycast;
    rows.current.length = 0;
    dashRows.current.length = 0;
    let drawn = 0;
    for (let row = 0; row < n; row++) {
      if (!edges.shown(row)) continue;
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

      const tail = ARROW_HEAD_LENGTH;
      const headBase = len - ARROW_HEAD_LENGTH;
      const shaft = Math.max(headBase - tail, 0);

      pos.current.set(sx, sy, sz).addScaledVector(dir.current, tail + shaft / 2);
      const hot = hotSrc >= 0 && ((src === hotSrc && dst === hotDst) || (src === hotDst && dst === hotSrc));
      const state = states.get(src * 65536 + dst) ?? 0;
      const style = STATE_STYLE[state];
      const w = hot ? HOT_WIDTH : style ? style.width : 1;
      const hw = !hot && state === EDGE_UNPICKED ? DASH_WIDTH : w;
      if (!hot && state === EDGE_UNPICKED && shaft > 0) {
        const piece = shaft / (DASHES * 2 - 1);
        for (let k = 0; k < DASHES && dashes < capacity * DASHES; k++) {
          pos.current.set(sx, sy, sz).addScaledVector(dir.current, tail + piece * (2 * k + 0.5));
          scl.current.set(DASH_WIDTH, piece, DASH_WIDTH);
          mat.current.compose(pos.current, quat.current, scl.current);
          dash.setMatrixAt(dashes, mat.current);
          col.current.set(DASH_COLOR);
          dash.setColorAt(dashes, col.current);
          dashRows.current[dashes] = row;
          dashes++;
        }
      }
      scl.current.set(w, shaft, w);
      mat.current.compose(pos.current, quat.current, scl.current);
      line.setMatrixAt(drawn, mat.current);

      pos.current.set(sx, sy, sz).addScaledVector(dir.current, headBase + ARROW_HEAD_LENGTH / 2);
      scl.current.set(hw, hw, hw);
      mat.current.compose(pos.current, quat.current, scl.current);
      head.setMatrixAt(drawn, mat.current);

      const shaftColor = hot ? HOT_COLOR : style ? style.color : EDGE_LINE_COLOR;
      col.current.set(!hot && state === EDGE_UNPICKED ? DASH_COLOR : shaftColor);
      head.setColorAt(drawn, col.current);
      col.current.set(shaftColor);
      line.setColorAt(drawn, col.current);
      rows.current[drawn] = row;

      drawn++;
    }
    line.count = drawn;
    head.count = drawn;
    dash.count = dashes;
    line.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    dash.instanceMatrix.needsUpdate = true;
    if (line.instanceColor) line.instanceColor.needsUpdate = true;
    if (head.instanceColor) head.instanceColor.needsUpdate = true;
    if (dash.instanceColor) dash.instanceColor.needsUpdate = true;
    if (drawn > 0) {
      line.computeBoundingSphere();
      head.computeBoundingSphere();
    }
    if (dashes > 0) dash.computeBoundingSphere();
  });

  return (
    <>
      {}
      <instancedMesh ref={lineRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={noRaycast} userData={{ [SCENE_EDGE_TAG]: rows.current }}>
        <cylinderGeometry args={[EDGE_LINE_RADIUS, EDGE_LINE_RADIUS, 1, 8]} />
        {}
        {}
        <meshBasicMaterial color={INSTANCE_TINT_BASE} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[ARROW_HEAD_RADIUS, ARROW_HEAD_LENGTH, 12]} />
        <meshBasicMaterial color={INSTANCE_TINT_BASE} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      <instancedMesh ref={dashRef} args={[undefined, undefined, capacity * DASHES]} frustumCulled={false} raycast={noRaycast} userData={{ [SCENE_EDGE_TAG]: dashRows.current }}>
        <cylinderGeometry args={[EDGE_LINE_RADIUS, EDGE_LINE_RADIUS, 1, 6]} />
        {}
        {}
        <meshBasicMaterial color={INSTANCE_TINT_BASE} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      {}
    </>
  );
}
