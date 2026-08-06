import React from "react";
import { postGoRecord } from "../vscode-api";
import { encodeDistanceGroupAdjust } from "../../schema/input-layout";
import { useDistanceGroupLens } from "./overlay-flags";
import {
  panelStyle,
  itemRowStyle,
  itemStyle,
  labelStyle,
  valueStyle,
  arrowCellStyle,
  btnStyle,
} from "./panel-styles";

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
// SpeedSlider/AbcDragLabel's reflect pattern. Rendered as a plain JSX child of
// ThreeView (alongside HomeButton/OverlaysControl in camera-ui.tsx), not a
// portal. That matters for anchoring/scroll: a portal into a static toolbar mount
// elsewhere in the DOM is a different containing block and would not scroll with the
// page. The look itself is shared with TiltVectorAnglePanel (panel-styles.ts).
const GROUPS: { index: number; label: string }[] = [
  { index: 0, label: "time" },
  { index: 1, label: "input" },
  { index: 2, label: "select" },
];

export function DistanceHomePanel() {
  const lens = useDistanceGroupLens();

  // Data-driven, not scene-branching: a scene whose nodes don't resolve any of the 3
  // groups' pairs (e.g. the pair scene, whose only nodes are outside every group in
  // nodes/Wiring/distance_groups.go's distanceGroups table) streams all three
  // GroupLen* columns as 0 forever — distanceGroupMax's `any` never turns true, so
  // ApplyDistanceGroupTarget's math never has anything to act on either. That is
  // exactly the "no groups" condition, so render nothing rather than a panel of
  // dead buttons. A real ring-topology group briefly reading 0 before its first
  // resolvable frame lands is the same signal and self-corrects the next frame.
  if (!lens || (lens.time === 0 && lens.input === 0 && lens.gate === 0)) return null;

  const valueFor = (index: number): number | undefined => {
    if (!lens) return undefined;
    if (index === 0) return lens.time;
    if (index === 1) return lens.input;
    return lens.gate;
  };

  const adjust = (index: number, dir: "up" | "down") => {
    postGoRecord(encodeDistanceGroupAdjust(index, dir));
  };

  return (
    // One STACK per group, side by side. Everything about a group — its name, its length,
    // its ▲▼ — lives inside that group's own box, so nothing is padded to line up with a
    // neighbour.
    <div style={{ ...panelStyle, ...itemRowStyle }}>
      {GROUPS.map(({ index, label }) => {
        const v = valueFor(index);
        return (
          <span style={itemStyle} key={label}>
            <span style={labelStyle}>{label}</span>
            <span style={valueStyle}>{v === undefined ? "—" : Math.round(v)}</span>
            <span style={arrowCellStyle}>
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
            </span>
          </span>
        );
      })}
    </div>
  );
}
