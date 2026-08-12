import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "../nodes/node-frame-aggregate";
import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius, readNodeSelected, readNodeHovered,
} from "../../../../schema/buffer-layout/buffer-layout";
import { readOverlaySelectionRing, readOverlayHoverRing } from "../../../../schema/buffer-layout/buffer-layout";
import { overlayOn } from "../../controls/flags/overlay-flags";
import { NODE_SPHERE_RADIUS, HOVER_COLOR, HOVER_RING_TUBE_RATIO } from "../buffer-scene-shared";

const SELECTION_RING_TUBE_RATIO = 0.14;
const SELECTION_RING_RADIAL_SEGMENTS = 8;
const SELECTION_RING_TUBULAR_SEGMENTS = 32;
const SELECTION_HALO_R_RATIO = 1.45;
const SELECTION_HALO_WIDTH_SEGMENTS = 16;
const SELECTION_HALO_HEIGHT_SEGMENTS = 16;

export function SelectionHighlight() {
  const groupRef = useRef<THREE.Group | null>(null);

  useFrame(() => {
    const g = groupRef.current;
    if (!g) return;

    const decoded = getNodeFrame();
    let show = false;
    if (decoded) {
      const { nodeCount, nodeView } = decoded;

      let selectedRow = -1;
      for (let i = 0; i < nodeCount; i++) {
        if (readNodeSelected(nodeView, i)) { selectedRow = i; break; }
      }

      if (selectedRow >= 0) {
        const r = readNodeRadius(nodeView, selectedRow) || NODE_SPHERE_RADIUS;
        g.position.set(
          readNodeCX(nodeView, selectedRow),
          readNodeCY(nodeView, selectedRow),
          readNodeCZ(nodeView, selectedRow),
        );

        g.scale.setScalar(r);
        show = true;
      }
    }

    g.visible = show && overlayOn(readOverlaySelectionRing);
  });

  return (
    <group ref={groupRef} visible={false}>
      {}
      <mesh raycast={() => null} frustumCulled={false}>
        <torusGeometry args={[1, SELECTION_RING_TUBE_RATIO, SELECTION_RING_RADIAL_SEGMENTS, SELECTION_RING_TUBULAR_SEGMENTS]} />
        <meshStandardMaterial color="#ffcc00" emissive="#ffcc00" emissiveIntensity={0.3} />
      </mesh>
      {}
      <mesh raycast={() => null} frustumCulled={false}>
        <sphereGeometry args={[SELECTION_HALO_R_RATIO, SELECTION_HALO_WIDTH_SEGMENTS, SELECTION_HALO_HEIGHT_SEGMENTS]} />
        <meshBasicMaterial
          color="#ff5a00"
          transparent
          opacity={0.5}
          side={THREE.DoubleSide}
          depthWrite={false}
        />
      </mesh>
    </group>
  );
}

export function HoverHighlight() {
  const ringRef = useRef<THREE.Mesh>(null);

  useFrame(() => {
    const ring = ringRef.current;
    if (!ring) return;

    const decoded = getNodeFrame();
    let show = false;
    if (decoded) {
      const { nodeCount, nodeView } = decoded;

      let hoveredRow = -1;
      for (let i = 0; i < nodeCount; i++) {
        if (readNodeHovered(nodeView, i)) { hoveredRow = i; break; }
      }

      if (hoveredRow >= 0) {

        const suppressed =
          readNodeSelected(nodeView, hoveredRow) !== 0 && overlayOn(readOverlaySelectionRing);
        if (!suppressed) {
          const r = readNodeRadius(nodeView, hoveredRow) || NODE_SPHERE_RADIUS;
          ring.position.set(
            readNodeCX(nodeView, hoveredRow),
            readNodeCY(nodeView, hoveredRow),
            readNodeCZ(nodeView, hoveredRow),
          );
          ring.scale.setScalar(r); 
          show = true;
        }
      }
    }
    ring.visible = show && overlayOn(readOverlayHoverRing);
  });

  return (
    <mesh ref={ringRef} visible={false} raycast={() => null} frustumCulled={false}>
      {}
      <torusGeometry args={[1, HOVER_RING_TUBE_RATIO, 8, 32]} />
      <meshStandardMaterial color={HOVER_COLOR} emissive={HOVER_COLOR} emissiveIntensity={0.3} />
    </mesh>
  );
}
