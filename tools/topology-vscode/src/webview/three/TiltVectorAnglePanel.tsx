import React, { useState } from "react";
import { postGoRecord } from "../vscode-api";
import { encodeTiltVectorAdjust, encodeSceneLatticePoints } from "../../schema/input-layout";
import { useTiltVectorRows, type TiltVectorRow } from "./overlay-flags";
import { formatAngle, widestAngle } from "./tilt-vector-angle-format";
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

/** One axis item inside a node's group, STACKED: the axis name on its own line, its value
 *  and the ▲/▼ that change it on the line below. Same two-line item the pair panels use.
 *
 *  It is stacked rather than laid across one line because a single line has to decide what
 *  fills the space between the name and the value — and the answer keeps being "nothing
 *  should". The first version gave the name `flex: "1 1 auto"`, which stretched it to the
 *  popover's full width and pushed the value and arrows out to the right edge, opening a gap
 *  across every row. Stacking has nothing to stretch: the name is its own line, so the gap
 *  it opened cannot exist. The value line below it does span the row, deliberately — that is
 *  what carries the arrows out to the shared right-hand column (valueLineStyle,
 *  arrowGroupStyle) without a stretched name pushing the value with them.
 *  Styled from popoverRowStyle (hover background, radius, padding) with the
 *  direction overridden — the chrome is shared, only the flow differs. */
/** The value half of an item's second line, sized to `widest` rather than to what it shows.
 *
 *  The shown value is taken OUT OF FLOW over an invisible copy of the widest string the
 *  readout can ever hold, so the box is that wide whatever is in it. Everything after it —
 *  the ▲/▼ — therefore keeps one position while the number steps, instead of sliding as
 *  "0" becomes "-11π/12" and back. Reserving by rendering the widest string measures it in
 *  the real font; a `ch`/`em` guess does not, and "π" is exactly where such a guess is
 *  wrong. `tabular-nums` holds the digits themselves to one width, so the shown value does
 *  not shift inside the reserved box either. */
function ValueBox({ shown, widest }: { shown: string; widest: string }) {
  return (
    <span style={valueBoxStyle}>
      <span aria-hidden style={{ visibility: "hidden" }}>{widest}</span>
      <span style={valueTextStyle}>{shown}</span>
    </span>
  );
}

function AxisRow({ node, axis }: { node: TiltVectorRow; axis: (typeof AXES)[number] }) {
  const [hover, setHover] = useState(false);
  const adjust = (dir: "up" | "down") => {
    postGoRecord(encodeTiltVectorAdjust(node.row, dir));
  };
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        ...popoverRowStyle(hover, false),
        flexDirection: "column",
        alignItems: "stretch",
        gap: 2,
      }}
    >
      <span>{axis}</span>
      <span style={valueLineStyle}>
        <ValueBox
          shown={formatAngle(node.theta, node.points)}
          widest={widestAngle(node.points)}
        />
        <span style={arrowGroupStyle}>
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
  const [hover, setHover] = useState(false);
  const adjust = (delta: number) => {
    const next = Math.min(LATTICE_POINTS_MAX, Math.max(LATTICE_POINTS_MIN, points + delta));
    if (next === points) return;
    postGoRecord(encodeSceneLatticePoints(next));
  };
  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        ...popoverRowStyle(hover, false),
        flexDirection: "column",
        alignItems: "stretch",
        gap: 2,
      }}
    >
      <span>Lattice points</span>
      <span style={valueLineStyle}>
        <ValueBox shown={String(points)} widest={String(LATTICE_POINTS_MAX)} />
        <span style={arrowGroupStyle}>
          <button
            type="button"
            aria-label="lattice points up"
            disabled={points >= LATTICE_POINTS_MAX}
            onClick={(e) => { e.stopPropagation(); adjust(LATTICE_POINTS_STEP); }}
            style={points >= LATTICE_POINTS_MAX ? arrowBtnDisabledStyle : arrowBtnStyle}
          >
            ▲
          </button>
          <button
            type="button"
            aria-label="lattice points down"
            disabled={points <= LATTICE_POINTS_MIN}
            onClick={(e) => { e.stopPropagation(); adjust(-LATTICE_POINTS_STEP); }}
            style={points <= LATTICE_POINTS_MIN ? arrowBtnDisabledStyle : arrowBtnStyle}
          >
            ▼
          </button>
        </span>
      </span>
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

// The value's own box: as wide as the widest value it can hold (see ValueBox), and the
// positioning context the shown value sits in.
const valueBoxStyle: React.CSSProperties = {
  position: "relative",
  display: "inline-block",
  fontVariantNumeric: "tabular-nums",
  whiteSpace: "nowrap",
};

// The shown value, out of flow so only the invisible widest copy sizes the box.
const valueTextStyle: React.CSSProperties = {
  position: "absolute",
  left: 0,
  top: 0,
};

// An item's second line: the value at the left, the arrows that change it at the RIGHT EDGE
// of the dropdown. The line takes the full row width (the rows stretch) and the arrow group
// is pushed to its end, so every ▲/▼ in the popover lands on one right-hand column whatever
// its row's value or label is.
const valueLineStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  width: "100%",
};

// The ▲/▼ pair, held together and pushed to the line's right end. `marginLeft: auto` eats
// whatever space is left over, which is what puts them on the shared right-hand column.
//
// This does NOT replace ValueBox's width reservation. The dropdown is sized by its widest
// content (anchorStyle's max-content), so a value that grows would widen the whole popover
// and carry the right edge — and the arrows with it — outward. The reservation holds that
// width still; this holds the arrows at it.
const arrowGroupStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  marginLeft: "auto",
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

// Same chrome as arrowBtnStyle but visibly inert — used at the lattice-point-count bounds
// so a disabled arrow reads as disabled rather than a click that silently does nothing.
const arrowBtnDisabledStyle: React.CSSProperties = {
  ...arrowBtnStyle,
  opacity: 0.35,
  cursor: "default",
};
