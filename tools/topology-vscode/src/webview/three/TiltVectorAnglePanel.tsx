import React from "react";
import { postGoRecord } from "../vscode-api";
import { encodeTiltVectorAdjust } from "../../schema/input-layout";
import { CURVE_PARAM_TILT_VECTOR_ANGLE_STEP } from "../../schema/curve-params";
import { useTiltVectorRows } from "./overlay-flags";
import {
  panelStyle,
  itemRowStyle,
  itemColumnStyle,
  itemStyle,
  labelStyle,
  valueStyle,
  arrowCellStyle,
  btnStyle,
} from "./panel-styles";

// The two angle axes a node exposes, in column order. One column per entry.
const AXES = ["theta", "phi"] as const;

// TiltVectorAnglePanel — a per-node tilt-vector-direction panel, sibling of
// DistanceHomePanel (same style constants: small dark rounded panel, monospace, ▲/▼
// arrows).
//
// WHICH nodes it can adjust is Go's answer, same data-driven shape as DistanceHomePanel:
// it reflects every node whose TopTiltVectorLen > 0 (useTiltVectorRows, overlay-flags.ts —
// the SAME column TiltVectors.tsx gates its own draw on). A scene whose nodes all stream
// TopTiltVectorLen 0 (no tilt vectors drawn at all) yields an EMPTY row list, and this panel
// renders nothing — no scene branch on either side, just the shared "no rows" signal
// DistanceHomePanel's "no groups" check uses.
//
// θ/φ are displayed as an INTEGER MULTIPLE of Go's own step
// (nodes/Wiring.CurveParamTiltVectorAngleStep, mirrored here as the generated
// CURVE_PARAM_TILT_VECTOR_ANGLE_STEP — memory/feedback_abc_times_constant_not_rederive.md):
// the index is the thing being adjusted, not the radians, so it is shown as "5π/12" rather
// than a decimal. TS computes the DISPLAYED index by dividing Go's own streamed radians by
// Go's own streamed step — a read-side format transform, not authored angle state (the
// same kind of transform TiltVectors.tsx already does converting the same θ/φ to a
// cartesian arrow direction).
//
// Clicking an arrow fire-and-forgets an edit-update(tiltVector, theta|phi) record naming
// the target node's buffer ROW (never its id/name) and the direction; Go owns the step
// and the index arithmetic (node_mover.go's moveMsgKindTiltVectorAngle handler) — this
// component sends no angle value, only which node + which axis + which direction.
//
// EVERY node with a tilt vector is listed at once, θ and φ each on their own line under
// the node's name. It used to show one node at a time behind ◀/▶, which meant comparing
// the two ends of a pair — the thing these angles are usually being set relative to —
// required flipping back and forth and remembering the other one. There is no local state
// at all now: the panel is a pure function of the reflected rows.
const DENOM = Math.max(1, Math.round(Math.PI / CURVE_PARAM_TILT_VECTOR_ANGLE_STEP));

function formatAngle(radians: number): string {
  const idx = Math.round(radians / CURVE_PARAM_TILT_VECTOR_ANGLE_STEP);
  if (idx === 0) return "0";
  const sign = idx < 0 ? "-" : "";
  return `${sign}${Math.abs(idx)}π/${DENOM}`;
}

export function TiltVectorAnglePanel() {
  const rows = useTiltVectorRows();

  // Data-driven "no rows" render-nothing, same shape as DistanceHomePanel's all-zero
  // check: null (no node frame decoded yet) or an empty list (this scene draws no tilt
  // vectors at all) both mean nothing to show.
  if (!rows || rows.length === 0) return null;

  const adjust = (row: number, axis: "theta" | "phi", dir: "up" | "down") => {
    postGoRecord(encodeTiltVectorAdjust(row, axis, dir));
  };

  return (
    // NESTED: the panel stacks its NODES; each node holds its own name and, inside it, a row
    // of AXIS stacks. So a node's whole angle state is one box, and each axis inside it is a
    // smaller box of name/value/▲▼ — the same item shape DistanceHomePanel uses, one level
    // down. Which θ/φ belongs to which node is read off the containment, not remembered.
    <div style={{ ...panelStyle, ...itemColumnStyle }}>
      {rows.map((node, i) => (
        <div style={itemColumnStyle} key={node.row}>
          {/* A rule between nodes, so two nodes' axes cannot be misread as one node's four.
              Skipped before the first. */}
          {i > 0 && <div style={sepStyle} />}
          <div style={headerStyle}>{node.label || String(node.row)}</div>
          <div style={itemRowStyle}>
            {AXES.map((axis) => (
              <span style={itemStyle} key={axis}>
                <span style={labelStyle}>{axis}</span>
                <span style={valueStyle}>
                  {formatAngle(axis === "theta" ? node.theta : node.phi)}
                </span>
                <span style={arrowCellStyle}>
                  <button
                    type="button"
                    style={btnStyle}
                    aria-label={`${node.label || node.row} ${axis} up`}
                    onClick={() => adjust(node.row, axis, "up")}
                  >
                    ▲
                  </button>
                  <button
                    type="button"
                    style={btnStyle}
                    aria-label={`${node.label || node.row} ${axis} down`}
                    onClick={() => adjust(node.row, axis, "down")}
                  >
                    ▼
                  </button>
                </span>
              </span>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

// A node's name at the top of that node's own box — no column span needed any more, since
// the node IS a box and its name is simply the first thing in it.
const headerStyle: React.CSSProperties = {
  color: "#fff",
  paddingTop: 1,
};

const sepStyle: React.CSSProperties = {
  height: 1,
  background: "rgba(255,255,255,0.18)",
  margin: "3px 0",
};
