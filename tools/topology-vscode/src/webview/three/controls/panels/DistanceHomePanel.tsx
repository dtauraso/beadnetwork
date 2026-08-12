import React, { useState } from "react";
import { postGoRecord } from "../../../vscode-api";
import { encodeDistanceGroupAdjust } from "../../../../schema/input-encode";
import { useDistanceGroupLens } from "../flags/overlay-flags-distance-groups";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  PILL_ANCHOR_STYLE,
  inFlowPopoverStyle,
} from "../pills/overlay-chrome";
import { StepperRow } from "../pills/pill-rows";























const GROUPS: { index: number; label: string }[] = [
  { index: 0, label: "time" },
  { index: 1, label: "input" },
  { index: 2, label: "select" },
];





function widestLength(values: number[]): string {
  const digits = Math.max(4, ...values.map((v) => String(Math.round(v)).length));
  return "8".repeat(digits);
}

export function DistanceHomePanel() {
  const lens = useDistanceGroupLens();
  const [open, setOpen] = useState(false);









  if (!lens || (lens.time === 0 && lens.input === 0 && lens.gate === 0)) return null;

  const valueFor = (index: number): number => {
    if (index === 0) return lens.time;
    if (index === 1) return lens.input;
    return lens.gate;
  };

  const adjust = (index: number, dir: "up" | "down") => {
    postGoRecord(encodeDistanceGroupAdjust(index, dir));
  };

  const onToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  };

  const widest = widestLength(GROUPS.map(({ index }) => valueFor(index)));

  return (



    <div style={PILL_ANCHOR_STYLE}>
      <div style={pillContainerStyle(false)}>
        {}
        <div
          onClick={onToggle}
          title={open ? "Close distances" : "Open distances"}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Distances
        </div>
        <div
          onClick={onToggle}
          title={open ? "Close distances" : "Open distances"}
          style={pillCaretStyle}
        >
          {open ? "▲" : "▼"}
        </div>
      </div>

      {open && (
        <div style={inFlowPopoverStyle()}>
          {GROUPS.map(({ index, label }) => (
            <StepperRow
              key={label}
              name={label}
              shown={String(Math.round(valueFor(index)))}
              widest={widest}
              upLabel={`${label} distance up`}
              downLabel={`${label} distance down`}
              onUp={() => adjust(index, "up")}
              onDown={() => adjust(index, "down")}
            />
          ))}
        </div>
      )}
    </div>
  );
}
