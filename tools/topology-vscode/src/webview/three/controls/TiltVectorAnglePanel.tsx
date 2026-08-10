import React, { useState } from "react";
import { postGoRecord } from "../../vscode-api";
import { encodeTiltVectorAdjust, encodeSceneLatticePoints } from "../../../schema/input-encode";
import { useTiltVectorRows, type TiltVectorRow } from "./overlay-flags-tilt-vectors";
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
} from "./overlay-chrome";
import { StepperRow } from "./pill-rows";

// The angle axes this panel lets a node's tilt be set on, in display order.
//
// THETA ONLY. There is no φ anywhere in the tilt-vector model any more
// (task/drop-tilt-vector-phi removed it end to end): TiltVectorMsg, the buffer columns,
// and the bridge attribute are all θ-only now, so this is not "a control withheld", it is
// the whole vocabulary. Every derived direction is θ arithmetic — bottom is θ+12, the
// coplanar normal is θ+6, the outgoing vector is that normal −12 — and the acute test that
// decides a step reads the same θ lattice as pure integer index arithmetic
// (Wiring.TiltVectorIsAcute).
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
// θ is displayed as an INTEGER MULTIPLE of THIS NODE'S OWN lattice step — 2π/points, where
// `points` is the LIVE streamed lattice point count (Buffer/layout.go's LatticePoints,
// task/pair-lattice-points), not the fixed compile-time
// CurveParamTiltVectorAngleStep/CURVE_PARAM_TILT_VECTOR_ANGLE_STEP (π/12, a 24-point
// default). That fixed constant is only right at 24 points; deriving from the streamed
// count instead keeps the index and its shown fraction denominator correct at whatever
// count the scene setting currently holds (6 of 24 shows "6π/12", 3 of 12 shows "3π/6" —
// same index, half the denominator, at half the points). TS computes the DISPLAYED index by
// dividing Go's own streamed radians by that per-node step — a read-side format transform,
// not authored angle state.
//
// Clicking an arrow fire-and-forgets an edit-update(tiltVector, theta) record naming
// the target node's buffer ROW (never its id/name) and the direction; Go owns the step
// and the index arithmetic (node_mover.go's moveMsgKindTiltVectorAngle handler) — this
// component sends no angle value, only which node + which direction.
//
// The actual derivation lives in tilt-vector-angle-format.ts (formatAngle, imported above)
// — split out so it has no react/vscode-api dependency and its own unit test doesn't need
// to import a webview module.

/** One axis item inside a node's group: the shared StepperRow (pill-rows.tsx), which every
 *  pill's popover uses — the name on its own line, the value and its ▲/▼ below. */
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
      {/* The list measures as nothing and lays out full width, so expanding θ changes no
          width and what does not fit wraps (REVEALED_LIST_STYLE). */}
      {open && (
        <div style={REVEALED_LIST_STYLE}>
          {AXES.map((axis) => <AxisRow key={axis} node={node} axis={axis} />)}
        </div>
      )}
    </div>
  );
}

// Pair-lattice point count bounds (Buffer/layout.go's LatticePoints /
// nodes/Wiring/scene_lattice_persist.go): valid = a multiple of 4 in 4..64.
const LATTICE_POINTS_MIN = 4;
const LATTICE_POINTS_MAX = 64;
const LATTICE_POINTS_STEP = 4;

/** SCENE-LEVEL row: the pair lattice's current point count, with ▲/▼ that step it by 4
 *  (clamped 4..64) and fire-and-forget an edit-update(scene, latticePoints) record — the
 *  same shape as AxisRow's arrows, but this is a SCENE setting (one value for the whole
 *  scene, not per-node), so it sits once at the top of the popover rather than inside a
 *  node's group. Disabled at each bound rather than letting a click silently do nothing
 *  (memory/feedback_clear_button_armed_only_when_loaded.md's "don't bank an action a
 *  disabled affordance can't take" shape). */
function LatticePointsRow({ points }: { points: number }) {
  const adjust = (delta: number) => {
    postGoRecord(encodeSceneLatticePoints(points + delta));
  };
  // No handler at a bound, which is what renders that arrow disabled (StepperRow) — the
  // affordance says it cannot take the click rather than banking one that does nothing.
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
    <div style={PILL_ANCHOR_STYLE}>
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
        <div style={inFlowPopoverStyle()}>
          {/* Scene-level, not per-node: one row, at the top, using whichever row's own
              streamed count happens to be current — every node in the scene streams the
              same LatticePoints value (task/pair-lattice-points). */}
          <LatticePointsRow points={rows[0]?.points ?? 24} />
          {rows.map((node) => (
            <NodeGroupSection key={node.row} node={node} />
          ))}
        </div>
      )}
    </div>
  );
}

// Nothing about a row's look lives here any more: the value reservation, the right-aligned
// arrows and their disabled state are StepperRow's (pill-rows.tsx), and the pill and its
// dropdown share ONE WIDTH via PILL_ANCHOR_STYLE / inFlowPopoverStyle() (overlay-chrome.ts).
// Both are used by the overlays and distances controls too.
