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

// NOTHING HERE DECLARES A COLUMN WIDTH. The panel is a GRID, and a grid's `auto` column is
// exactly as wide as the widest thing in it — so the columns line up down the panel with no
// width stated anywhere, and no slack.
//
// The slack these panels had came from stating widths at all. Each row was its own flex
// line, so aligning the values and the ▲/▼ across rows meant giving the label and value
// cells a fixed `minWidth` — and any fixed width wide enough for the longest string is too
// wide for every other row. A value cell sized for a 3-digit length left ~24px of empty
// panel beside a right-aligned "0". Two rounds of re-guessing those numbers (px, then `ch`)
// changed the amount of slack and could not remove it: the guess itself was the defect, and
// the second guess made one panel WIDER than the px it replaced.
//
// A grid removes the question. Rows are not separate flex lines any more — every cell is a
// direct child of one grid, so column alignment is structural, and each column shrinks to
// its own content.

export const panelStyle: React.CSSProperties = {
  // Placed by ThreeView's right-hand flex column, not by a top/right of its own — these
  // panels' heights depend on how many rows Go streams, so nothing below can be positioned
  // against a number known here. The column takes no pointer events; each panel re-enables
  // them for itself so the canvas stays draggable in the gaps.
  pointerEvents: "auto",
  // A grid of content-sized columns — ONE PER NAME. Each panel sets its own count with
  // gridColumns() below, since that count is how many things the panel lists. `auto` means
  // "as wide as this column's widest cell", so alignment comes from the grid rather than
  // from any stated width.
  display: "inline-grid",
  alignItems: "center",
  justifyItems: "center",
  columnGap: 8,
  rowGap: 2,
  background: "rgba(0,0,0,0.55)",
  borderRadius: 6,
  padding: "3px 6px",
  color: "#ddd",
  fontSize: 11,
  fontFamily: "monospace",
  userSelect: "none",
};

// gridColumns states how many content-sized columns a panel has — one per name it lists.
export const gridColumns = (n: number): React.CSSProperties => ({
  gridTemplateColumns: `repeat(${n}, auto)`,
});

// A NAME, sitting at the top of its own column as that column's title. The panel reads
// across the top and down: name, then its value, then its arrows. Every cell of a column is
// therefore about one thing, and the name is not repeated beside each row.
export const labelStyle: React.CSSProperties = { whiteSpace: "nowrap" };

// The value under its own name. Centred (via the grid's justifyItems) rather than
// right-aligned: a column holds one value now, so there is no column of digits to align the
// last character of.
export const valueStyle: React.CSSProperties = { whiteSpace: "nowrap" };

// The ▲/▼ pair under a value, kept together as one cell so the arrows stay with the column
// they act on.
export const arrowCellStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  gap: 2,
};

// A cell that spans every column — a node's name, or the rule between two nodes.
export const fullRowStyle: React.CSSProperties = { gridColumn: "1 / -1" };


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
