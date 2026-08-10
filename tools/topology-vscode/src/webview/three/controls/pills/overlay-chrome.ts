import React from "react";

// overlay-chrome.ts — the shared LOOK of the right-hand column's PILL+POPOVER controls
// (OverlaysControl in overlays-control.tsx, the tilt-vector angle panel). Both controls are a
// labeled pill with a ▼/▲ disclosure caret that opens a popover; the popover holds
// collapsible ▶/▼ group headings, each holding rows. That chrome used to live only in
// overlays-control.tsx; a second control copying its style constants verbatim is how a look that is
// meant to be one look drifts. One definition here, imported by all three controls (overlays,
// angles, distances) — the row shapes that go inside a popover are pill-rows.tsx.

export const CHROME_FONT_STACK = '-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif';

// The one text colour for this chrome — pill labels, group headings, rows, and the ▲▼/▶▼
// triangles that sit among them. It used to be two: rows at this value and everything else
// at #9a9aa6, a grey that read as FADED next to them rather than as hierarchy, with the
// carets dimmed a further 0.85 on top. Nothing here is disabled, so nothing here should look
// disabled: the popover already separates a heading from a row by size, weight, case and
// letter-spacing, none of which need dimming to do their job.
//
// Dimming stays where it MEANS something: an inert checkbox (overlay-row.tsx's 0.45), an
// arrow at its bound (arrowBtnDisabledStyle's 0.35).
export const CHROME_TEXT = "#e7e7ea";

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
    // Dark ink on the accent fill when active; otherwise the chrome's own text colour —
    // an unlit pill is not a disabled one.
    color: active ? "#04101f" : CHROME_TEXT,
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
  // No opacity here. It was 0.85, which faded the caret against the label beside it — and
  // the caret is the affordance that says the pill opens, so it is the last thing that
  // should read as half-there. It is already smaller than the label; that is the hierarchy.
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
    color: CHROME_TEXT,
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

/** The wrapper a pill and its popover live in. It sets NO width of its own: ThreeView's
 *  right-hand column stretches every widget in it to one width, so all three pills
 *  (overlays, angles, distances) come out the same width as each other and their popovers
 *  come out the same width as their pills. A `max-content` here — what this used to
 *  carry — would opt each control out of that and size it to its own label again.
 *
 *  Closed, the column measures the pills, so the shared width is "the widest pill" and the
 *  three start out matching. An OPEN popover is measured too, and if its content is wider
 *  it sizes the column — the pill above it grows to fit exactly as it did before the pills
 *  were tied together, and the other pills, sharing the column's width, grow with it. What
 *  a popover never does is change width again once open: the rows a group reveals measure
 *  as nothing (REVEALED_LIST_STYLE), so expanding one wraps rather than widening.
 *
 *  The popover is IN FLOW, not absolutely positioned. Out of flow it would contribute its
 *  width to nothing AND take no space, so it could only be given a width chosen in advance
 *  — the guess that kept leaving a band down its right — and would cover what sits below
 *  it. ThreeView's column is built for the in-flow version: an open popover displaces what
 *  is below it instead.
 *
 *  Pointer-transparent itself — the column takes no pointer events and each widget re-enables
 *  them for its own box, so a wrapper that swallowed them would cover the canvas behind it. */
export const PILL_ANCHOR_STYLE: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  alignItems: "stretch",
  gap: 4,
  pointerEvents: "none",
};

/** The popover with the chrome above but not its positioning: in flow inside
 *  PILL_ANCHOR_STYLE, filling the width the column resolved to — and contributing its own
 *  content to that resolution, so an open popover wider than the pills carries them with it
 *  rather than wrapping its top-level rows. */
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
    color: CHROME_TEXT,
    borderRadius: 5,
    background: !disabled && hover ? "rgba(255,255,255,0.05)" : "transparent",
    userSelect: "none",
    fontSize: 11.5,
  };
}
