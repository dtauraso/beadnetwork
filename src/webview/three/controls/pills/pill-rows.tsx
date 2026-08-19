import React, { useState } from "react";
import { popoverRowStyle } from "./overlay-chrome";
import * as T from "../chrome-theme";

export function ValueBox({ shown, widest }: { shown: string; widest: string }) {
  return (
    <span style={valueBoxStyle}>
      <span aria-hidden style={{ visibility: "hidden" }}>{widest}</span>
      <span style={valueTextStyle}>{shown}</span>
    </span>
  );
}

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

const valueBoxStyle: React.CSSProperties = {
  position: "relative",
  display: "inline-block",
  fontVariantNumeric: "tabular-nums",
  whiteSpace: "nowrap",
};

const valueTextStyle: React.CSSProperties = {
  position: "absolute",
  left: 0,
  top: 0,
};

const valueLineStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  width: "100%",

  flexWrap: "wrap",
  rowGap: 2,
};

const arrowGroupStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 4,
  marginLeft: "auto",
};

const arrowBtnStyle: React.CSSProperties = {
  background: T.HOVER_ROW,
  border: "none",
  borderRadius: T.RADIUS_ITEM,
  color: T.TEXT,
  fontSize: T.FONT_SIZE_GLYPH,
  lineHeight: 1,
  padding: T.PAD_ITEM,
  cursor: "pointer",
};

const arrowBtnDisabledStyle: React.CSSProperties = {
  ...arrowBtnStyle,
  opacity: T.DISABLED_OPACITY,
  cursor: "default",
};
