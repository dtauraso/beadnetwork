import React from "react";

// overlay-chrome.ts — the shared LOOK of the right-hand column's PILL+POPOVER controls
// (OverlaysControl in camera-ui.tsx, the tilt-vector angle panel). Both controls are a
// labeled pill with a ▼/▲ disclosure caret that opens a popover; the popover holds
// collapsible ▶/▼ group headings, each holding rows. That chrome used to live only in
// camera-ui.tsx; a second control copying its style constants verbatim is exactly the
// pattern panel-styles.ts's own file comment warns against (two copies of one look drift
// silently). One definition here, imported by both.

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

/** The popover panel anchored under a pill: `position: relative` on the pill container
 *  above, this absolutely positioned at `top: calc(100% + 4px); right: 0`. */
export function popoverStyle(width: number): React.CSSProperties {
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
