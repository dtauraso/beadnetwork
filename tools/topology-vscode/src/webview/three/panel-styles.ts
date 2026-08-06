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

// NOTHING HERE DECLARES A WIDTH. A panel is a row of STACKS: each item — a distance group,
// a node's angle axis — is one self-contained box holding its own name, value and ▲▼, and
// each box is as wide as its own widest line.
//
// That is what removes the slack these panels had. Before, every item was a horizontal
// row, so lining the values and arrows up down the panel meant giving the label and value
// cells a fixed `minWidth` — and any width wide enough for the longest string is too wide
// for every other row. A value cell sized for a 3-digit length left ~24px of empty panel
// beside a right-aligned "0". Two rounds of re-guessing those numbers (px, then `ch`)
// changed how much slack there was and could not remove it: the guess itself was the
// defect, and the second guess made one panel WIDER than the px it replaced.
//
// Stacking removes the question rather than answering it better. Nothing has to line up
// ACROSS items, so nothing has to be padded to a shared width.

export const panelStyle: React.CSSProperties = {
  // Placed by ThreeView's right-hand flex column, not by a top/right of its own — these
  // panels' heights depend on how many rows Go streams, so nothing below can be positioned
  // against a number known here. The column takes no pointer events; each panel re-enables
  // them for itself so the canvas stays draggable in the gaps.
  pointerEvents: "auto",
  // Sized by its contents in both directions. Panels that lay their items out side by side
  // add itemRowStyle; ones that stack groups vertically (the tilt panel's nodes) add
  // itemColumnStyle.
  display: "inline-flex",
  background: "rgba(0,0,0,0.55)",
  borderRadius: 6,
  padding: "3px 6px",
  color: "#ddd",
  fontSize: 11,
  fontFamily: "monospace",
  userSelect: "none",
};

// Items laid out side by side, and items stacked one above another. A panel picks whichever
// its own contents call for; the tilt panel uses both, nesting one inside the other.
export const itemRowStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "flex-start",
  gap: 8,
};

// The gap BETWEEN items is deliberately larger than the gap inside one (itemStyle's gap of
// 1, holding a name against its own value). That difference is what groups the two lines of
// an item together and separates it from the next one.
export const itemColumnStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: 7,
};

// ONE ITEM, on two lines: its NAME, and under it its VALUE beside its ▲▼. This box is the
// unit a panel is built from — everything about one thing is inside it, and it is as wide as
// its own widest line. Left-aligned, so the names and the values below them start on the
// same edge and the whole panel reads down that edge.
export const itemStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  alignItems: "flex-start",
  gap: 1,
};

// An item's second line: the value, then the arrows that change it.
export const valueRowStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
};

export const labelStyle: React.CSSProperties = { whiteSpace: "nowrap" };

export const valueStyle: React.CSSProperties = { whiteSpace: "nowrap" };

// The ▲/▼ pair at the bottom of an item, kept together so the arrows stay with the thing
// they act on.
export const arrowCellStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  gap: 2,
};

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
