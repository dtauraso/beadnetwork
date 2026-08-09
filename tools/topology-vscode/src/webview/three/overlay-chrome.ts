import React from "react";

// overlay-chrome.ts — the shared LOOK of the right-hand column's PILL+POPOVER controls
// (OverlaysControl in camera-ui.tsx, the tilt-vector angle panel). Both controls are a
// labeled pill with a ▼/▲ disclosure caret that opens a popover; the popover holds
// collapsible ▶/▼ group headings, each holding rows. That chrome used to live only in
// camera-ui.tsx; a second control copying its style constants verbatim is how a look that is
// meant to be one look drifts. One definition here, imported by all three controls (overlays,
// angles, distances) — the row shapes that go inside a popover are pill-rows.tsx.

export const CHROME_FONT_STACK = '-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif';

/** The split-button/pill's outer chip: neutral by default, accent-filled when `active`
 *  (a control with no master toggle, like the angle panel, always passes active=false). */
export function pillContainerStyle(active: boolean): React.CSSProperties {
  return {
    position: "relative",
    zIndex: 20,
    pointerEvents: "auto",
    display: "flex",
    alignItems: "stretch",
    borderRadius: 6,
    overflow: "hidden",
    fontSize: 11,
    fontWeight: 600,
    fontFamily: CHROME_FONT_STACK,
    background: active ? "#4ea1ff" : "#34343d",
    border: `1px solid ${active ? "#4ea1ff" : "#3a3a44"}`,
    color: active ? "#04101f" : "#9a9aa6",
    userSelect: "none",
  };
}

/** The pill's label region (OverlaysControl's master-toggle body). A control with no
 *  master toggle uses this same padding/cursor for its whole clickable label instead. */
export const pillBodyStyle: React.CSSProperties = {
  padding: "3px 9px",
  cursor: "pointer",
  display: "flex",
  alignItems: "center",
};

/** The pill's ▼/▲ caret region. */
export const pillCaretStyle: React.CSSProperties = {
  padding: "3px 7px 3px 4px",
  cursor: "pointer",
  display: "flex",
  alignItems: "center",
  fontSize: 9,
  opacity: 0.85,
};

/** The popover panel anchored under a pill: absolutely positioned at
 *  `top: calc(100% + 4px); right: 0` against the nearest positioned ancestor.
 *
 *  NEVER RENDER THIS INSIDE THE PILL. `pillContainerStyle` sets `overflow: hidden` to clip
 *  the split-button's own rounded corners, and that clips a popover nested within it out of
 *  existence — the caret flips and nothing appears, which is exactly how the angles panel
 *  first shipped. Make the popover a SIBLING of the pill, inside a `position: relative`
 *  wrapper that sets no overflow (TiltVectorAnglePanel's `anchorStyle`). */
// width takes a CSS width, not just a px number: a popover whose contents are narrow and
// variable (the angles panel's node names) wants "max-content" — its own measurement —
// rather than a number picked in advance, which shows up as an empty band down its right.
export function popoverStyle(width: number | string): React.CSSProperties {
  return {
    position: "absolute",
    top: "calc(100% + 4px)",
    right: 0,
    zIndex: 21,
    pointerEvents: "auto",
    width,
    background: "#2f2f37",
    border: "1px solid #3a3a44",
    borderRadius: 8,
    padding: 6,
    boxShadow: "0 8px 24px rgba(0,0,0,0.5)",
    fontFamily: CHROME_FONT_STACK,
    userSelect: "none",
  };
}

/** A collapsible group heading inside a popover (OverlayGroupSection's heading row). */
export function groupHeadingStyle(hover: boolean): React.CSSProperties {
  return {
    display: "flex",
    alignItems: "center",
    gap: 5,
    fontSize: 9.5,
    textTransform: "uppercase",
    letterSpacing: "0.05em",
    color: "#9a9aa6",
    padding: "5px 6px 4px",
    cursor: "pointer",
    borderRadius: 5,
    background: hover ? "rgba(255,255,255,0.05)" : "transparent",
  };
}

// ▶/▼ (U+25B6/U+25BC), not the ▸/▾ small variants: those render as thin arrowheads in
// several of the fonts this stack falls back to, which reads as a link chevron rather
// than a disclosure triangle.
export const DISCLOSURE_GLYPH_STYLE: React.CSSProperties = { fontSize: 8, width: 8, flex: "0 0 auto" };

/** The wrapper a pill and its popover live in, and what gives them ONE WIDTH: a max-content
 *  column whose two children both stretch to it. The width is therefore the widest thing in
 *  either — the pill's label, or a group heading in the popover — so the pill and the
 *  popover always come out the same width.
 *
 *  This only works with the popover IN FLOW (inFlowPopoverStyle below). An absolutely
 *  positioned popover is out of flow, so it contributes its width to nothing: the wrapper
 *  would size to the pill alone and the popover could only be given a width chosen in
 *  advance — the guess that kept leaving a band down its right. In flow, the widest child
 *  sizes the wrapper and the other stretches to match.
 *
 *  ThreeView's right-hand column is built for this: it stacks its widgets, so an open
 *  popover displaces what is below it instead of covering it.
 *
 *  Pointer-transparent itself — the column takes no pointer events and each widget re-enables
 *  them for its own box, so a wrapper that swallowed them would cover the canvas behind it. */
export const PILL_ANCHOR_STYLE: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  alignItems: "stretch",
  width: "max-content",
  gap: 4,
  pointerEvents: "none",
};

/** The popover with the chrome above but not its positioning: in flow inside
 *  PILL_ANCHOR_STYLE, filling the width the anchor resolved to. */
export function inFlowPopoverStyle(): React.CSSProperties {
  return { ...popoverStyle("100%"), position: "static", boxSizing: "border-box" };
}

/** Wrapper for the rows a disclosure triangle REVEALS. They lay out across the popover's
 *  full width but MEASURE AS NOTHING, so expanding a group cannot change any width —
 *  content too long for the popover wraps onto the next line instead of pushing the edge
 *  (and, where the pill shares that width, the pill) outward.
 *
 *  `width: 0` is what buys it: a popover sized by `max-content` counts a child with a
 *  DEFINITE width as that width — zero — rather than as its contents, so the popover stays
 *  sized by what is there before any triangle is clicked (the headings). `minWidth: "100%"`
 *  then expands this box back out to the popover's resolved width. (The effect
 *  `contain: inline-size` describes, in a form that reads as ordinary sizing.)
 *
 *  Rows inside it must be able to BREAK — a flex row of fixed-width children with nothing
 *  wrappable will still overflow. See OverlayRow's label and the angle panel's value line. */
export const REVEALED_LIST_STYLE: React.CSSProperties = { width: 0, minWidth: "100%" };

/** A single popover row (OverlayRow's shape), minus the checkbox glyph — callers that want
 *  a checkbox (OverlaysControl) render their own leading element; callers that don't (the
 *  angle panel's axis rows) render their content directly inside this. */
export function popoverRowStyle(hover: boolean, disabled: boolean): React.CSSProperties {
  return {
    display: "flex",
    alignItems: "center",
    gap: 7,
    padding: "4px 6px",
    cursor: disabled ? "default" : "pointer",
    color: "#e7e7ea",
    borderRadius: 5,
    background: !disabled && hover ? "rgba(255,255,255,0.05)" : "transparent",
    userSelect: "none",
    fontSize: 11.5,
  };
}
