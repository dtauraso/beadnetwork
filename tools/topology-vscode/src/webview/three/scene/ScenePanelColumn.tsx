import type { RefObject } from "react";
import * as THREE from "three";
import { FitButton } from "../../../../FitButton/FitButton";
import { OverlaysControl } from "../controls/pills/overlays-control";
import { NodesDropdown } from "../../../../NodesDropdown/NodesDropdown";
import { AngleDropdown } from "../../../../AngleDropdown/AngleDropdown";
import { pillContainerStyle, pillBodyStyle, pillCaretStyle } from "../controls/pills/overlay-chrome";

const COLUMN_GAP = 6;

const PILL_LABELS = ["Distances", "Angles", "Nodes", "Overlays"];

function PillColumnSizer() {
  return (
    <div
      aria-hidden
      style={{ height: 0, marginBottom: -COLUMN_GAP, overflow: "hidden", visibility: "hidden" }}
    >
      {PILL_LABELS.map((label) => (
        <div key={label} style={pillContainerStyle(false)}>
          <div style={{ ...pillBodyStyle, flex: "1 1 auto" }}>{label}</div>
          <div style={pillCaretStyle}>▼</div>
        </div>
      ))}
    </div>
  );
}

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
        gap: COLUMN_GAP,
        pointerEvents: "none",
      }}
    >
      {}
      <PillColumnSizer />
      <FitButton cameraRef={cameraRef} aspect={aspect} />
      <AngleDropdown />
      {}
      <NodesDropdown />
      <OverlaysControl />
    </div>
  );
}
