import React from "react";
import * as T from "../chrome-theme";

export const CHROME_FONT_STACK = T.FONT_STACK;

export const CHROME_TEXT = T.TEXT;

export function pillContainerStyle(active: boolean): React.CSSProperties {
  return {
    position: "relative",
    zIndex: 20,
    pointerEvents: "auto",
    display: "flex",
    alignItems: "stretch",
    borderRadius: T.RADIUS_CHIP,
    overflow: "hidden",
    fontSize: T.FONT_SIZE,
    fontWeight: T.FONT_WEIGHT_LABEL,
    fontFamily: CHROME_FONT_STACK,
    background: active ? T.ACCENT : T.CHIP,
    border: `1px solid ${active ? T.ACCENT : T.BORDER}`,

    color: active ? T.ACCENT_INK : CHROME_TEXT,
    userSelect: "none",
  };
}

export const pillBodyStyle: React.CSSProperties = {
  padding: T.PAD_PILL_BODY,
  cursor: "pointer",
  display: "flex",
  alignItems: "center",
};

export const pillCaretStyle: React.CSSProperties = {
  padding: "3px 7px 3px 4px",
  cursor: "pointer",
  display: "flex",
  alignItems: "center",
  fontSize: T.FONT_SIZE_GLYPH,

};

export function popoverStyle(width: number | string): React.CSSProperties {
  return {
    position: "absolute",
    top: "calc(100% + 4px)",
    right: 0,
    zIndex: 21,
    pointerEvents: "auto",
    width,
    background: T.SURFACE,
    border: `1px solid ${T.BORDER}`,
    borderRadius: T.RADIUS_PANEL,
    padding: T.PAD_PANEL,
    boxShadow: T.PANEL_SHADOW,
    fontFamily: CHROME_FONT_STACK,
    userSelect: "none",
  };
}

export function groupHeadingStyle(hover: boolean): React.CSSProperties {
  return {
    display: "flex",
    alignItems: "center",
    gap: 5,
    fontSize: T.FONT_SIZE_HEADING,
    textTransform: "uppercase",
    letterSpacing: T.HEADING_TRACKING,
    color: CHROME_TEXT,
    padding: T.PAD_HEADING,
    cursor: "pointer",
    borderRadius: T.RADIUS_ITEM,
    background: hover ? T.HOVER_ROW : "transparent",
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
    padding: T.PAD_ROW,
    cursor: disabled ? "default" : "pointer",
    color: CHROME_TEXT,
    borderRadius: T.RADIUS_ITEM,
    background: !disabled && hover ? T.HOVER_ROW : "transparent",
    userSelect: "none",
    fontSize: T.FONT_SIZE,
  };
}
