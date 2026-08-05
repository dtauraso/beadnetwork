import React from "react";

// panel-styles.ts — the shared look of the right-hand column's LIST PANELS
// (DistanceHomePanel, TiltVectorAnglePanel): a dark rounded pill, 11px monospace, #ddd,
// laid out as a vertical list of label/value/▲/▼ rows. Both panels held their own verbatim
// copy of every constant below, each copy's comment claiming to "mirror" the other. Two
// copies of a look that is meant to be one look drift silently — and did: the same
// `flex: 1` label stretch reached both, so a fix to one panel's spacing left the other
// looking different. One definition instead, imported by both.
//
// The panel is sized by its CONTENT. Rows pack from the left into two aligned columns
// (label, then right-aligned value) with the arrows immediately after, so the pill is only
// as wide as the widest row needs. The label cell deliberately does NOT flex: giving it
// `flex: 1` made it absorb all the panel's slack, pushing the value and arrows out to the
// right edge and opening a gap across the middle of every row that read as empty space.
// A fixed minimum is what makes a COLUMN — the values line up down the panel — without
// that stretch.

// Column widths, in px. Small enough that a two-node tilt panel and a three-group distance
// panel both come out compact, wide enough that "theta"/"input" and a 3-digit value still
// line up rather than ragging.
const LABEL_COL = 34;
const VALUE_COL = 30;

export const panelStyle: React.CSSProperties = {
  // Placed by ThreeView's right-hand flex column, not by a top/right of its own — these
  // panels' heights depend on how many rows Go streams, so nothing below can be positioned
  // against a number known here. The column takes no pointer events; each panel re-enables
  // them for itself so the canvas stays draggable in the gaps.
  pointerEvents: "auto",
  display: "inline-flex",
  flexDirection: "column",
  gap: 2,
  background: "rgba(0,0,0,0.55)",
  borderRadius: 6,
  padding: "3px 6px",
  color: "#ddd",
  fontSize: 11,
  fontFamily: "monospace",
  userSelect: "none",
};

export const rowStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  whiteSpace: "nowrap",
};

// No `flex: 1` — see the file comment. minWidth alone aligns the labels into a column and
// lets the row end where its content ends.
export const labelStyle: React.CSSProperties = { minWidth: LABEL_COL };

export const valueStyle: React.CSSProperties = { minWidth: VALUE_COL, textAlign: "right" };

export const btnStyle: React.CSSProperties = {
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
