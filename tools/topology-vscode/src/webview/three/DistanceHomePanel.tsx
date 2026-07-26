import { createPortal } from "react-dom";
import { postGoRecord } from "../vscode-api";
import { encodeDistanceGroupAdjust } from "../../schema/input-layout";
import { useDistanceGroupLens } from "./overlay-flags";

// DistanceHomePanel — the "distance home button" toolbar panel: 3 named groups of
// node-pair distances (time / input / gate — Go's distanceGroupOrder,
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
  { index: 2, label: "gate" },
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
    <span className="distance-home-panel">
      {GROUPS.map(({ index, label }) => {
        const v = valueFor(index);
        return (
          <span className="distance-home-row" key={label}>
            <span className="distance-home-label">{label}</span>
            <span className="distance-home-value">{v === undefined ? "—" : Math.round(v)}</span>
            <button
              type="button"
              aria-label={`${label} distance up`}
              onClick={() => adjust(index, "up")}
            >
              ▲
            </button>
            <button
              type="button"
              aria-label={`${label} distance down`}
              onClick={() => adjust(index, "down")}
            >
              ▼
            </button>
          </span>
        );
      })}
    </span>,
    mount,
  );
}
