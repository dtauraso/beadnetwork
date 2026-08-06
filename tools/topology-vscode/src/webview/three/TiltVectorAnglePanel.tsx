import React, { useState } from "react";
import { postGoRecord } from "../vscode-api";
import { encodeTiltVectorAdjust } from "../../schema/input-layout";
import { CURVE_PARAM_TILT_VECTOR_ANGLE_STEP } from "../../schema/curve-params";
import { useTiltVectorRows, type TiltVectorRow } from "./overlay-flags";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  popoverStyle,
  groupHeadingStyle,
  DISCLOSURE_GLYPH_STYLE,
  popoverRowStyle,
} from "./overlay-chrome";

// The two angle axes a node exposes, in column order.
const AXES = ["theta", "phi"] as const;

// TiltVectorAnglePanel — the PAIR tab's tilt-vector-direction control, built on the SAME
// pill + popover chrome as OverlaysControl (overlay-chrome.ts): a labeled pill in ThreeView's
// right-hand column that opens a popover, one collapsible group per node, one row per axis.
// This control has no master toggle (there is nothing to turn on/off, only angles to read and
// adjust), so unlike OverlaysControl's split button, the WHOLE pill — body and caret alike —
// just opens/closes the popover.
//
// WHICH nodes it can adjust is Go's answer, same data-driven shape as DistanceHomePanel:
// it reflects every node whose TopTiltVectorLen > 0 (useTiltVectorRows, overlay-flags.ts —
// the SAME column TiltVectors.tsx gates its own draw on). A scene whose nodes all stream
// TopTiltVectorLen 0 (no tilt vectors drawn at all) yields an EMPTY row list, and the whole
// pill renders nothing — no scene branch on either side, just the shared "no rows" signal
// DistanceHomePanel's "no groups" check uses.
//
// θ/φ are displayed as an INTEGER MULTIPLE of Go's own step
// (nodes/Wiring.CurveParamTiltVectorAngleStep, mirrored here as the generated
// CURVE_PARAM_TILT_VECTOR_ANGLE_STEP — memory/feedback_abc_times_constant_not_rederive.md):
// the index is the thing being adjusted, not the radians, so it is shown as "5π/12" rather
// than a decimal. TS computes the DISPLAYED index by dividing Go's own streamed radians by
// Go's own streamed step — a read-side format transform, not authored angle state.
//
// Clicking an arrow fire-and-forgets an edit-update(tiltVector, theta|phi) record naming
// the target node's buffer ROW (never its id/name) and the direction; Go owns the step
// and the index arithmetic (node_mover.go's moveMsgKindTiltVectorAngle handler) — this
// component sends no angle value, only which node + which axis + which direction.
const DENOM = Math.max(1, Math.round(Math.PI / CURVE_PARAM_TILT_VECTOR_ANGLE_STEP));

function formatAngle(radians: number): string {
  const idx = Math.round(radians / CURVE_PARAM_TILT_VECTOR_ANGLE_STEP);
  if (idx === 0) return "0";
  const sign = idx < 0 ? "-" : "";
  return `${sign}${Math.abs(idx)}π/${DENOM}`;
}

/** One axis row inside a node's group: axis name, its formatted value, ▲/▼ adjust buttons.
 *  Styled like OverlayRow (popoverRowStyle), minus the checkbox glyph — there is nothing to
 *  check here, only a value to show and adjust. */
function AxisRow({ node, axis }: { node: TiltVectorRow; axis: (typeof AXES)[number] }) {
  const [hover, setHover] = useState(false);
  const adjust = (dir: "up" | "down") => {
    postGoRecord(encodeTiltVectorAdjust(node.row, axis, dir));
  };
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={popoverRowStyle(hover, false)}
    >
      <span style={{ flex: "1 1 auto" }}>{axis}</span>
      <span style={{ fontVariantNumeric: "tabular-nums" }}>
        {formatAngle(axis === "theta" ? node.theta : node.phi)}
      </span>
      <button
        type="button"
        aria-label={`${node.label || node.row} ${axis} up`}
        onClick={(e) => { e.stopPropagation(); adjust("up"); }}
        style={arrowBtnStyle}
      >
        ▲
      </button>
      <button
        type="button"
        aria-label={`${node.label || node.row} ${axis} down`}
        onClick={(e) => { e.stopPropagation(); adjust("down"); }}
        style={arrowBtnStyle}
      >
        ▼
      </button>
    </div>
  );
}

/** One collapsible node group, styled like OverlayGroupSection's heading. Collapsed by
 *  default, same as the overlay groups. */
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
        <span style={{ flex: "1 1 auto" }}>{heading}</span>
      </div>
      {open && AXES.map((axis) => <AxisRow key={axis} node={node} axis={axis} />)}
    </div>
  );
}

/** ANGLES CONTROL: a labeled pill (no master toggle — the whole pill opens the popover) +
 *  popover of per-node collapsible groups, one row per axis. Same pill/popover/heading/row
 *  chrome as OverlaysControl (overlay-chrome.ts). */
export function TiltVectorAnglePanel() {
  const rows = useTiltVectorRows();
  const [open, setOpen] = useState(false);

  // Data-driven "no rows" render-nothing, same shape as DistanceHomePanel's all-zero
  // check: null (no node frame decoded yet) or an empty list (this scene draws no tilt
  // vectors at all) both mean nothing to show — the whole pill, not just the popover.
  if (!rows || rows.length === 0) return null;

  const onToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  };

  return (
    // The popover is a SIBLING of the pill, never a child: pillContainerStyle sets
    // `overflow: hidden` (it clips the split-button's own rounded corners), which also clips
    // an absolutely-positioned popover inside it out of existence — the caret flipped and
    // nothing appeared. This wrapper is what the popover anchors to instead, and it sets no
    // overflow of its own.
    <div style={anchorStyle}>
      <div style={pillContainerStyle(false)}>
        {/* No master toggle: the whole pill (body + caret) opens/closes the popover. */}
        <div onClick={onToggle} title={open ? "Close angles" : "Open angles"} style={pillBodyStyle}>
          Angles
        </div>
        <div onClick={onToggle} title={open ? "Close angles" : "Open angles"} style={pillCaretStyle}>
          {open ? "▲" : "▼"}
        </div>
      </div>

      {open && (
        <div style={popoverStyle(170)}>
          {rows.map((node) => (
            <NodeGroupSection key={node.row} node={node} />
          ))}
        </div>
      )}
    </div>
  );
}

// What the popover is positioned against: a wrapper around the pill with no overflow of its
// own, so `popoverStyle`'s absolute `top: calc(100% + 4px)` measures from the BOTTOM OF THE
// PILL and is not clipped by it. It must stay pointer-transparent itself — ThreeView's
// right-hand column takes no pointer events and each widget re-enables them for its own box,
// so a wrapper that swallowed them would cover the canvas behind this panel.
const anchorStyle: React.CSSProperties = {
  position: "relative",
  pointerEvents: "none",
};

const arrowBtnStyle: React.CSSProperties = {
  background: "rgba(255,255,255,0.12)",
  border: "none",
  borderRadius: 4,
  color: "#e7e7ea",
  fontSize: 10,
  lineHeight: 1,
  padding: "2px 5px",
  cursor: "pointer",
};
