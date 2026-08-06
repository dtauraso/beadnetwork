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

// The angle axes this panel lets a node's tilt be set on, in display order.
//
// THETA ONLY. φ is deliberately absent: the panel only ever has rows in a scene whose nodes
// draw tilt vectors (TopTiltVectorLen > 0, which Go streams only where SceneTab.UpAxis is
// set — the pair), and there the straightening model turns on θ alone. Every derived
// direction is θ arithmetic with φ carried through untouched — bottom is θ+12, the coplanar
// normal is θ+6, the outgoing vector is that normal −12 — and the two dot products that
// decide a step read the same θ lattice. A φ control offered a knob no rule here consults.
//
// Go still decodes a φ edit on the tiltVector entity; nothing sends one now. That is the
// same shape as the rest of the vocabulary Go accepts without a live TS sender
// (.claude/rules/bridge-surface.md) — the capability is not removed, just not offered.
const AXES = ["theta"] as const;

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

/** One axis item inside a node's group, STACKED: the axis name on its own line, its value
 *  and the ▲/▼ that change it on the line below. Same two-line item the pair panels use.
 *
 *  It is stacked rather than laid across one line because a single line has to decide what
 *  fills the space between the name and the value — and the answer keeps being "nothing
 *  should". The first version gave the name `flex: "1 1 auto"`, which stretched it to the
 *  popover's full width and pushed the value and arrows out to the right edge, opening a gap
 *  across every row. Stacking has nothing to stretch: each line is as wide as its own
 *  content. Styled from popoverRowStyle (hover background, radius, padding) with the
 *  direction overridden — the chrome is shared, only the flow differs. */
function AxisRow({ node, axis }: { node: TiltVectorRow; axis: (typeof AXES)[number] }) {
  const [hover, setHover] = useState(false);
  const adjust = (dir: "up" | "down") => {
    postGoRecord(encodeTiltVectorAdjust(node.row, axis, dir));
  };
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        ...popoverRowStyle(hover, false),
        flexDirection: "column",
        alignItems: "flex-start",
        gap: 2,
      }}
    >
      <span>{axis}</span>
      <span style={valueLineStyle}>
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
      </span>
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
        {/* No `flex: "1 1 auto"` here. Overlays' heading stretches because it has a count
            chip to push to the far end; this heading has nothing after it, so stretching it
            only holds the popover open past its content. */}
        <span>{heading}</span>
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
    // The dropdown is a SIBLING of the pill, never a child: pillContainerStyle sets
    // `overflow: hidden` (it clips the split-button's own rounded corners), which also clips
    // anything positioned inside it out of existence — the caret flipped and nothing
    // appeared. Both are children of the shared-width wrapper instead (anchorStyle).
    <div style={anchorStyle}>
      <div style={pillContainerStyle(false)}>
        {/* No master toggle: the whole pill (body + caret) opens/closes the dropdown. The
            caret is pushed to the far end so the pill fills the shared width rather than
            leaving its own slack — the label and the caret are its only content. */}
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
        <div style={dropdownStyle}>
          {rows.map((node) => (
            <NodeGroupSection key={node.row} node={node} />
          ))}
        </div>
      )}
    </div>
  );
}

// An item's second line: the value, then the arrows that change it, packed together.
const valueLineStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
};

// The pill and its dropdown share ONE WIDTH, and this wrapper defines it: a max-content
// column whose two children both stretch to it. The width is therefore the widest thing in
// either — the pill's label, a node heading, or an axis item — so the pill, the node groups
// (first level) and the axis items (second level) all come out the same width.
//
// That is why the dropdown is IN FLOW here rather than absolutely positioned like the
// overlays popover. An absolute popover is out of flow, so it contributes its width to
// nothing: the wrapper would size to the pill alone, and the dropdown could only be given a
// width chosen in advance — the guess that kept leaving a band down its right. In flow, the
// widest child sizes the wrapper and the other stretches to match.
//
// ThreeView's right-hand column is built for this: it stacks its widgets, and "a panel that
// grows pushes the rest down rather than overlapping them", so an open dropdown displaces
// what is below it instead of covering it.
//
// Pointer-transparent itself — the column takes no pointer events and each widget re-enables
// them for its own box, so a wrapper that swallowed them would cover the canvas behind it.
const anchorStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  alignItems: "stretch",
  width: "max-content",
  gap: 4,
  pointerEvents: "none",
};

// The dropdown takes the overlays popover's CHROME but not its positioning: in flow (see
// anchorStyle) and filling the shared width.
const dropdownStyle: React.CSSProperties = {
  ...popoverStyle("100%"),
  position: "static",
  boxSizing: "border-box",
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
