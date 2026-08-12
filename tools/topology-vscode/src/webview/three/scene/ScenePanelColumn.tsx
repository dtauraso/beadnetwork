import type { RefObject } from "react";
import * as THREE from "three";
import { HomeButton } from "../controls/panels/home-button";
import { OverlaysControl } from "../controls/pills/overlays-control";
import { NodePalette } from "../controls/panels/NodePalette";
import { DistanceHomePanel } from "../controls/panels/DistanceHomePanel";
import { TiltVectorAnglePanel } from "../controls/panels/TiltVectorAnglePanel";

export function ScenePanelColumn({ cameraRef, aspect }: { cameraRef: RefObject<THREE.PerspectiveCamera | null>; aspect: number }) {
  return (
    <div
      style={{
        position: "absolute",
        top: 44,
        right: 12,
        zIndex: 20,
        display: "flex",
        flexDirection: "column",

        alignItems: "stretch",
        gap: 6,
        pointerEvents: "none",
      }}
    >
      {}
      <HomeButton cameraRef={cameraRef} aspect={aspect} />
      <DistanceHomePanel />
      <TiltVectorAnglePanel />
      {}
      <NodePalette />
      <OverlaysControl />
    </div>
  );
}
