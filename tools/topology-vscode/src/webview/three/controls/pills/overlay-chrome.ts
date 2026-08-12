import React from "react";

export const CHROME_FONT_STACK = '-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif';

export const CHROME_TEXT = "#e7e7ea";

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

    color: active ? "#04101f" : CHROME_TEXT,
    userSelect: "none",
  };
}

export const pillBodyStyle: React.CSSProperties = {
  padding: "3px 9px",
  cursor: "pointer",
  display: "flex",
  alignItems: "center",
};

export const pillCaretStyle: React.CSSProperties = {
  padding: "3px 7px 3px 4px",
  cursor: "pointer",
  display: "flex",
  alignItems: "center",
  fontSize: 9,

};

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

export const DISCLOSURE_GLYPH_STYLE: React.CSSProperties = { fontSize: 8, width: 8, flex: "0 0 auto" };

export const PILL_ANCHOR_STYLE: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  alignItems: "stretch",
  gap: 4,
  pointerEvents: "none",
};

export function inFlowPopoverStyle(): React.CSSProperties {
  return { ...popoverStyle("100%"), position: "static", boxSizing: "border-box" };
}

export const REVEALED_LIST_STYLE: React.CSSProperties = { width: 0, minWidth: "100%" };

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
