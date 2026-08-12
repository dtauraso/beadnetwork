import React, { useState } from "react";
import { postGoRecord } from "../../../vscode-api";
import { encodeTiltVectorAdjust, encodeSceneLatticePoints } from "../../../../schema/input-encode";
import { useTiltVectorRows, type TiltVectorRow } from "../flags/overlay-flags-tilt-vectors";
import { formatAngle, widestAngle } from "./tilt-vector-angle-format";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  groupHeadingStyle,
  DISCLOSURE_GLYPH_STYLE,
  REVEALED_LIST_STYLE,
  PILL_ANCHOR_STYLE,
  inFlowPopoverStyle,
} from "../pills/overlay-chrome";
import { StepperRow } from "../pills/pill-rows";

const AXES = ["theta"] as const;

function AxisRow({ node, axis }: { node: TiltVectorRow; axis: (typeof AXES)[number] }) {
  const adjust = (dir: "up" | "down") => {
    postGoRecord(encodeTiltVectorAdjust(node.row, dir));
  };
  const who = node.label || node.row;
  return (
    <StepperRow
      name={axis}
      shown={formatAngle(node.theta, node.points)}
      widest={widestAngle(node.points)}
      upLabel={`${who} ${axis} up`}
      downLabel={`${who} ${axis} down`}
      onUp={() => adjust("up")}
      onDown={() => adjust("down")}
    />
  );
}

function NodeGroupSection({ node }: { node: TiltVectorRow }) {
  const [open, setOpen] = useState(false);
  const [hover, setHover] = useState(false);
  const heading = node.label || String(node.row);
  return (
    <div>
      <div
        onClick={(e) => { e.stopPropagation(); setOpen((o) => !o); }}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        title={open ? `Collapse ${heading}` : `Expand ${heading}`}
        style={groupHeadingStyle(hover)}
      >
        <span style={DISCLOSURE_GLYPH_STYLE}>{open ? "▼" : "▶"}</span>
        {}
        <span>{heading}</span>
      </div>
      {}
      {open && (
        <div style={REVEALED_LIST_STYLE}>
          {AXES.map((axis) => <AxisRow key={axis} node={node} axis={axis} />)}
        </div>
      )}
    </div>
  );
}

const LATTICE_POINTS_MIN = 4;
const LATTICE_POINTS_MAX = 64;
const LATTICE_POINTS_STEP = 4;

function LatticePointsRow({ points }: { points: number }) {
  const adjust = (delta: number) => {
    postGoRecord(encodeSceneLatticePoints(points + delta));
  };

  return (
    <StepperRow
      name="Lattice points"
      shown={String(points)}
      widest={String(LATTICE_POINTS_MAX)}
      upLabel="lattice points up"
      downLabel="lattice points down"
      onUp={points >= LATTICE_POINTS_MAX ? undefined : () => adjust(LATTICE_POINTS_STEP)}
      onDown={points <= LATTICE_POINTS_MIN ? undefined : () => adjust(-LATTICE_POINTS_STEP)}
    />
  );
}

export function TiltVectorAnglePanel() {
  const rows = useTiltVectorRows();
  const [open, setOpen] = useState(false);

  if (!rows || rows.length === 0) return null;

  const onToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  };

  return (

    <div style={PILL_ANCHOR_STYLE}>
      <div style={pillContainerStyle(false)}>
        {}
        <div
          onClick={onToggle}
          title={open ? "Close angles" : "Open angles"}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Angles
        </div>
        <div onClick={onToggle} title={open ? "Close angles" : "Open angles"} style={pillCaretStyle}>
          {open ? "▲" : "▼"}
        </div>
      </div>

      {open && (
        <div style={inFlowPopoverStyle()}>
          {}
          <LatticePointsRow points={rows[0]?.points ?? 24} />
          {rows.map((node) => (
            <NodeGroupSection key={node.row} node={node} />
          ))}
        </div>
      )}
    </div>
  );
}

