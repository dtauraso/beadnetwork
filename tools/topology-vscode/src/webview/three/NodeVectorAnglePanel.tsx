import React from "react";
import { postGoRecord } from "../vscode-api";
import { encodeNodeVectorAdjust } from "../../schema/input-layout";
import { CURVE_PARAM_VECTOR_ANGLE_STEP } from "../../schema/curve-params";
import { useNodeVectorRows } from "./overlay-flags";

// NodeVectorAnglePanel — a per-node vector-direction panel, sibling of DistanceHomePanel
// (same style constants: small dark rounded panel, monospace, ▲/▼ arrows).
//
// WHICH nodes it can adjust is Go's answer, same data-driven shape as DistanceHomePanel:
// it reflects every node whose VectorLen > 0 (useNodeVectorRows, overlay-flags.ts — the
// SAME column NodeVectors.tsx gates its own draw on). A scene whose nodes all stream
// VectorLen 0 (no vectors drawn at all) yields an EMPTY row list, and this panel renders
// nothing — no scene branch on either side, just the shared "no rows" signal
// DistanceHomePanel's "no groups" check uses.
//
// θ/φ are displayed as an INTEGER MULTIPLE of Go's own step
// (nodes/Wiring.CurveParamVectorAngleStep, mirrored here as the generated
// CURVE_PARAM_VECTOR_ANGLE_STEP — memory/feedback_abc_times_constant_not_rederive.md):
// the index is the thing being adjusted, not the radians, so it is shown as "5π/12" rather
// than a decimal. TS computes the DISPLAYED index by dividing Go's own streamed radians by
// Go's own streamed step — a read-side format transform, not authored angle state (the
// same kind of transform NodeVectors.tsx already does converting the same θ/φ to a
// cartesian arrow direction).
//
// Clicking an arrow fire-and-forgets an edit-update(nodeVector, theta|phi) record naming
// the target node's buffer ROW (never its id/name) and the direction; Go owns the step
// and the index arithmetic (node_mover.go's moveMsgKindVectorAngle handler) — this
// component sends no angle value, only which node + which axis + which direction.
//
// EVERY node with a vector is listed at once, θ and φ each on their own line under the
// node's name. It used to show one node at a time behind ◀/▶, which meant comparing the two
// ends of a pair — the thing these angles are usually being set relative to — required
// flipping back and forth and remembering the other one. There is no local state at all
// now: the panel is a pure function of the reflected rows.
const DENOM = Math.max(1, Math.round(Math.PI / CURVE_PARAM_VECTOR_ANGLE_STEP));

function formatAngle(radians: number): string {
  const idx = Math.round(radians / CURVE_PARAM_VECTOR_ANGLE_STEP);
  if (idx === 0) return "0";
  const sign = idx < 0 ? "-" : "";
  return `${sign}${Math.abs(idx)}π/${DENOM}`;
}

export function NodeVectorAnglePanel() {
  const rows = useNodeVectorRows();

  // Data-driven "no rows" render-nothing, same shape as DistanceHomePanel's all-zero
  // check: null (no node frame decoded yet) or an empty list (this scene draws no
  // vectors at all) both mean nothing to show.
  if (!rows || rows.length === 0) return null;

  const adjust = (row: number, axis: "theta" | "phi", dir: "up" | "down") => {
    postGoRecord(encodeNodeVectorAdjust(row, axis, dir));
  };

  return (
    <div style={panelStyle}>
      {rows.map((node, i) => (
        <React.Fragment key={node.row}>
          {/* A rule between nodes, so two nodes' θ/φ lines cannot be misread as one node's
              four. Skipped before the first. */}
          {i > 0 && <div style={sepStyle} />}
          <div style={headerStyle}>{node.label || String(node.row)}</div>
          {(["theta", "phi"] as const).map((axis) => (
            <div style={rowStyle} key={axis}>
              <span style={labelStyle}>{axis}</span>
              <span style={valueStyle}>{formatAngle(axis === "theta" ? node.theta : node.phi)}</span>
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
            </div>
          ))}
        </React.Fragment>
      ))}
    </div>
  );
}

// Styling mirrors DistanceHomePanel's own panelStyle (itself mirroring the camera
// HomeButton): a dark rounded pill, 11px monospace, #ddd, vertical list. Positioned
// BELOW DistanceHomePanel's slot (top:66) — the two panels never render in the same
// scene today (distance groups are ring-only, this panel's rows are the pair scene's
// vectors), but stacking downward keeps this panel out of DistanceHomePanel's spot in
// case a future scene streams both.
const panelStyle: React.CSSProperties = {
  position: "absolute",
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

// A node's name above its own two angle lines, so which θ/φ belongs to which node is read
// off the layout rather than remembered.
const headerStyle: React.CSSProperties = {
  color: "#fff",
  opacity: 0.85,
  paddingTop: 1,
};

const sepStyle: React.CSSProperties = {
  height: 1,
  background: "rgba(255,255,255,0.18)",
  margin: "3px 0",
};
