import React, { useState } from "react";
import { popoverRowStyle } from "./overlay-chrome";

// pill-rows.tsx — the ROW SHAPES inside a pill's popover, shared by every control that has
// one (TiltVectorAnglePanel, DistanceHomePanel). overlay-chrome.ts holds the pill, the
// popover and the group heading; this holds what goes in the rows.
//
// StepperRow is the one row shape all of them use: a name on its own line, and below it the
// value with the ▲/▼ that change it. It is stacked rather than laid across one line because
// a single line has to decide what fills the space between the name and the value, and the
// answer keeps being "nothing should" — the version that gave the name `flex: 1 1 auto`
// stretched it to the popover's full width and opened a gap across every row.

/** The value half of a row, sized to `widest` rather than to what it shows.
 *
 *  The shown value is taken OUT OF FLOW over an invisible copy of the widest string the
 *  readout can ever hold, so the box is that wide whatever is in it. The ▲/▼ after it
 *  therefore keep one position while the value steps, instead of sliding as "0" becomes
 *  "-11π/12" and back — an arrow that moves under the cursor is a different button than the
 *  one clicked. Rendering the widest string measures it in the real font, which a `ch`/`em`
 *  guess does not, and "π" is exactly where such a guess is wrong. `tabular-nums` holds the
 *  digits to one width so the shown value does not shift inside the reserved box either. */
export function ValueBox({ shown, widest }: { shown: string; widest: string }) {
  return (
    <span style={valueBoxStyle}>
      <span aria-hidden style={{ visibility: "hidden" }}>{widest}</span>
      <span style={valueTextStyle}>{shown}</span>
    </span>
  );
}

/** A name, its value, and the ▲/▼ that step it. `widest` is the value reservation (see
 *  ValueBox); `upLabel`/`downLabel` are the arrows' accessible names, which differ per
 *  caller ("node-1 theta up", "time distance up"). An arrow with no handler renders
 *  disabled — visibly inert rather than a click that silently does nothing. */
export function StepperRow({
  name,
  shown,
  widest,
  upLabel,
  downLabel,
  onUp,
  onDown,
}: {
  name: string;
  shown: string;
  widest: string;
  upLabel: string;
  downLabel: string;
  onUp?: () => void;
  onDown?: () => void;
}) {
  const [hover, setHover] = useState(false);
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
      <span>{name}</span>
      <span style={valueLineStyle}>
        <ValueBox shown={shown} widest={widest} />
        <span style={arrowGroupStyle}>
          <ArrowButton glyph="▲" label={upLabel} onClick={onUp} />
          <ArrowButton glyph="▼" label={downLabel} onClick={onDown} />
        </span>
      </span>
    </div>
  );
}

function ArrowButton({
  glyph,
  label,
  onClick,
}: {
  glyph: string;
  label: string;
  onClick?: () => void;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      disabled={!onClick}
      onClick={(e) => {
        e.stopPropagation();
        onClick?.();
      }}
      style={onClick ? arrowBtnStyle : arrowBtnDisabledStyle}
    >
      {glyph}
    </button>
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

// A row's second line: the value at the left, the arrows at the RIGHT EDGE of the popover.
// The line takes the full row width and the arrow group is pushed to its end, so every ▲/▼
// in a popover lands on one right-hand column whatever its row's value or name is.
const valueLineStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  width: "100%",
  // When the value and its arrows do not both fit the popover's width, the arrow group
  // drops to the next line (still right-aligned there) rather than widening the popover.
  flexWrap: "wrap",
  rowGap: 2,
};

// The ▲/▼ pair, held together and pushed to the line's right end. `marginLeft: auto` eats
// whatever space is left over, which is what puts them on the shared right-hand column.
const arrowGroupStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  marginLeft: "auto",
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

// Same chrome as arrowBtnStyle but visibly inert — used at a stepper's bounds so a disabled
// arrow reads as disabled rather than as a click that silently does nothing.
const arrowBtnDisabledStyle: React.CSSProperties = {
  ...arrowBtnStyle,
  opacity: 0.35,
  cursor: "default",
};
