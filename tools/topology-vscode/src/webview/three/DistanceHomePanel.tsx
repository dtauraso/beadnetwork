import type React from "react";
import { createPortal } from "react-dom";
import { postGoRecord } from "../vscode-api";
import { encodeDistanceGroupAdjust } from "../../schema/input-layout";
import { useDistanceGroupLens } from "./overlay-flags";

// DistanceHomePanel — the "distance home button" toolbar panel: 3 named groups of
// node-pair distances (time / input / select — Go's distanceGroupOrder,
// whose third group is still keyed "gate" internally; only the visible label
// reads "select" to match the renamed Select* gate structs —
// nodes/Wiring/distance_groups.go), each showing its CURRENT max pair length
// (read-only reflect of the Overlay block's GroupLenTime/GroupLenInput/GroupLenGate
// columns) with an up/down arrow. Clicking an arrow fire-and-forgets an
// edit-update(distanceGroup, length) record naming the group's WIRE INDEX and the
// direction; Go owns the group definitions AND the ×1.1/÷1.1 math (this component
// sends no length value, only which group + which direction) and repositions every
// pair's target node via RootMove, which rebroadcasts geometry so the node moves and
// its edge redraws. No local domain state (only ephemeral UI, if any) — mirrors
// SpeedSlider/AbcDragLabel's portal + reflect pattern.
const GROUPS: { index: number; label: string }[] = [
  { index: 0, label: "time" },
  { index: 1, label: "input" },
  { index: 2, label: "select" },
];

export function DistanceHomePanel() {
  const lens = useDistanceGroupLens();
  const mount = document.getElementById("distance-home-mount");
  if (!mount) return null;

  const valueFor = (index: number): number | undefined => {
    if (!lens) return undefined;
    if (index === 0) return lens.time;
    if (index === 1) return lens.input;
    return lens.gate;
  };

  const adjust = (index: number, dir: "up" | "down") => {
    postGoRecord(encodeDistanceGroupAdjust(index, dir));
  };

  return createPortal(
    <div style={panelStyle}>
      {GROUPS.map(({ index, label }) => {
        const v = valueFor(index);
        return (
          <div style={rowStyle} key={label}>
            <span style={labelStyle}>{label}</span>
            <span style={valueStyle}>{v === undefined ? "—" : Math.round(v)}</span>
            <button
              type="button"
              style={btnStyle}
              aria-label={`${label} distance up`}
              onClick={() => adjust(index, "up")}
            >
              ▲
            </button>
            <button
              type="button"
              style={btnStyle}
              aria-label={`${label} distance down`}
              onClick={() => adjust(index, "down")}
            >
              ▼
            </button>
          </div>
        );
      })}
    </div>,
    mount,
  );
}

// Styling mirrors the camera HomeButton (camera-ui.tsx): a dark rounded pill,
// 11px monospace, #ddd — but laid out as a VERTICAL LIST (one group per row)
// instead of wrapping inline text.
const panelStyle: React.CSSProperties = {
  // Placed directly BELOW the camera HomeButton (camera-ui.tsx: absolute top:44
  // right:12, inside ThreeView's inset:0 viewport container). Fixed positioning
  // anchors it under that button regardless of the toolbar mount it portals into.
  position: "fixed",
  top: 66,
  right: 12,
  zIndex: 20,
  pointerEvents: "auto",
  display: "inline-flex",
  flexDirection: "column",
  gap: 2,
  background: "rgba(0,0,0,0.55)",
  borderRadius: 6,
  padding: "3px 7px",
  color: "#ddd",
  fontSize: 11,
  fontFamily: "monospace",
  userSelect: "none",
};

const rowStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  whiteSpace: "nowrap",
};

const labelStyle: React.CSSProperties = { flex: 1, minWidth: 40 };

const valueStyle: React.CSSProperties = { minWidth: 34, textAlign: "right" };

const btnStyle: React.CSSProperties = {
  background: "rgba(255,255,255,0.12)",
  border: "none",
  borderRadius: 4,
  color: "#ddd",
  fontSize: 11,
  fontFamily: "monospace",
  lineHeight: 1,
  padding: "2px 5px",
  cursor: "pointer",
};
