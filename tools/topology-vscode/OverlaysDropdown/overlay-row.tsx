import React, { useCallback, useState } from "react";
import { fireToggle, useToggleVal, type ToggleCfg } from "./overlay-toggle";
import { popoverRowStyle } from "../src/webview/three/controls/pills/overlay-chrome";
import * as T from "../src/webview/three/controls/chrome-theme";

export function OverlayRow({ cfg, disabled, indent = 0 }: { cfg: ToggleCfg; disabled?: boolean; indent?: number }) {
  const val = useToggleVal(cfg);
  const active = cfg.active(val);
  const [hover, setHover] = useState(false);
  const onClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      if (disabled) return;
      fireToggle(cfg, val);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [val, disabled]
  );
  const labelText = typeof cfg.label === "function" ? cfg.label(val) : cfg.label;
  const iconText = typeof cfg.icon === "function" ? cfg.icon(val) : cfg.icon;
  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      title={cfg.title(active)}
      style={{
        ...popoverRowStyle(hover, !!disabled),
        paddingLeft: 6 + indent * 14,
      }}
    >
      {}
      <span
        style={{
          width: 13,
          height: 13,
          flex: "0 0 auto",

          alignSelf: "flex-start",
          opacity: disabled ? T.DISABLED_OPACITY : 1,
          borderRadius: T.RADIUS_ITEM,
          border: `1px solid ${active ? T.ACCENT : T.TEXT}`,
          background: active ? T.ACCENT : "transparent",
          display: "grid",
          placeItems: "center",
          color: T.ACCENT_INK,
          fontSize: T.FONT_SIZE_GLYPH,
          fontWeight: 900,
          lineHeight: "11px",
        }}
      >
        {active ? "✓" : ""}
      </span>
      {}
      {}
      <span style={{ width: 11, flex: "0 0 auto", textAlign: "center", alignSelf: "flex-start" }}>
        {iconText}
      </span>
      {}
      <span style={{ minWidth: 0, overflowWrap: "break-word" }}>{labelText}</span>
    </div>
  );
}
