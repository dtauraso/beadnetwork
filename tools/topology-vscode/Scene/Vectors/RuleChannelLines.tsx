import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "../../src/webview/three/scene/nodes/node-frame-aggregate";
import { overlayFlag } from "../../src/webview/three/controls/flags/overlay-flags";
import {
  readChannelVectorShaftM0, readChannelVectorShaftM1, readChannelVectorShaftM2, readChannelVectorShaftM3,
  readChannelVectorShaftM4, readChannelVectorShaftM5, readChannelVectorShaftM6, readChannelVectorShaftM7,
  readChannelVectorShaftM8, readChannelVectorShaftM9, readChannelVectorShaftM10, readChannelVectorShaftM11,
  readChannelVectorShaftM12, readChannelVectorShaftM13, readChannelVectorShaftM14, readChannelVectorShaftM15,
  readChannelVectorHeadM0, readChannelVectorHeadM1, readChannelVectorHeadM2, readChannelVectorHeadM3,
  readChannelVectorHeadM4, readChannelVectorHeadM5, readChannelVectorHeadM6, readChannelVectorHeadM7,
  readChannelVectorHeadM8, readChannelVectorHeadM9, readChannelVectorHeadM10, readChannelVectorHeadM11,
  readChannelVectorHeadM12, readChannelVectorHeadM13, readChannelVectorHeadM14, readChannelVectorHeadM15,
} from "../../Buffer/buffer-layout";
import {
  SHADING_PARAM_CHANNEL_LINE_RADIUS,
  SHADING_PARAM_CHANNEL_HEAD_RADIUS,
  SHADING_PARAM_CHANNEL_HEAD_LENGTH,
} from "../../Buffer/shading-params";

const CHANNEL_COLOR = "#7b6bd6";

const CHANNEL_LINE_OPACITY = 0.2584;
const CHANNEL_HEAD_OPACITY = 0.3675;

function copyShaft(view: DataView, row: number, mesh: THREE.InstancedMesh, slot: number): void {
  const out = mesh.instanceMatrix.array;
  const b = slot * 16;
  out[b]      = readChannelVectorShaftM0(view, row);
  out[b + 1]  = readChannelVectorShaftM1(view, row);
  out[b + 2]  = readChannelVectorShaftM2(view, row);
  out[b + 3]  = readChannelVectorShaftM3(view, row);
  out[b + 4]  = readChannelVectorShaftM4(view, row);
  out[b + 5]  = readChannelVectorShaftM5(view, row);
  out[b + 6]  = readChannelVectorShaftM6(view, row);
  out[b + 7]  = readChannelVectorShaftM7(view, row);
  out[b + 8]  = readChannelVectorShaftM8(view, row);
  out[b + 9]  = readChannelVectorShaftM9(view, row);
  out[b + 10] = readChannelVectorShaftM10(view, row);
  out[b + 11] = readChannelVectorShaftM11(view, row);
  out[b + 12] = readChannelVectorShaftM12(view, row);
  out[b + 13] = readChannelVectorShaftM13(view, row);
  out[b + 14] = readChannelVectorShaftM14(view, row);
  out[b + 15] = readChannelVectorShaftM15(view, row);
}

function copyHead(view: DataView, row: number, mesh: THREE.InstancedMesh, slot: number): void {
  const out = mesh.instanceMatrix.array;
  const b = slot * 16;
  out[b]      = readChannelVectorHeadM0(view, row);
  out[b + 1]  = readChannelVectorHeadM1(view, row);
  out[b + 2]  = readChannelVectorHeadM2(view, row);
  out[b + 3]  = readChannelVectorHeadM3(view, row);
  out[b + 4]  = readChannelVectorHeadM4(view, row);
  out[b + 5]  = readChannelVectorHeadM5(view, row);
  out[b + 6]  = readChannelVectorHeadM6(view, row);
  out[b + 7]  = readChannelVectorHeadM7(view, row);
  out[b + 8]  = readChannelVectorHeadM8(view, row);
  out[b + 9]  = readChannelVectorHeadM9(view, row);
  out[b + 10] = readChannelVectorHeadM10(view, row);
  out[b + 11] = readChannelVectorHeadM11(view, row);
  out[b + 12] = readChannelVectorHeadM12(view, row);
  out[b + 13] = readChannelVectorHeadM13(view, row);
  out[b + 14] = readChannelVectorHeadM14(view, row);
  out[b + 15] = readChannelVectorHeadM15(view, row);
}

export function RuleChannelLines({ capacity }: { capacity: number }) {
  const lineRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);

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
    const { channelVectorCount, channelVectorView } = decoded;

    let drawn = 0;
    for (let row = 0; row < channelVectorCount && drawn < capacity; row++) {
      copyShaft(channelVectorView, row, line, drawn);
      copyHead(channelVectorView, row, head, drawn);
      drawn++;
    }

    line.count = drawn;
    head.count = drawn;
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
        <cylinderGeometry args={[SHADING_PARAM_CHANNEL_LINE_RADIUS, SHADING_PARAM_CHANNEL_LINE_RADIUS, 1, 6]} />
        <meshBasicMaterial color={CHANNEL_COLOR} toneMapped={false} transparent opacity={CHANNEL_LINE_OPACITY} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[SHADING_PARAM_CHANNEL_HEAD_RADIUS, SHADING_PARAM_CHANNEL_HEAD_LENGTH, 8]} />
        <meshBasicMaterial color={CHANNEL_COLOR} toneMapped={false} transparent opacity={CHANNEL_HEAD_OPACITY} />
      </instancedMesh>
    </>
  );
}
