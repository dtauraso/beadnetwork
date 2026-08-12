


import React, { useCallback, useState } from "react";
import { fireToggle, useToggleVal, type ToggleCfg } from "./overlay-toggle";
import { popoverRowStyle } from "./overlay-chrome";


export function OverlayRow({ cfg, disabled, indent }: { cfg: ToggleCfg; disabled?: boolean; indent?: boolean }) {
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
        paddingLeft: indent ? 20 : 6,
      }}
    >
      {}
      <span
        style={{
          width: 13,
          height: 13,
          flex: "0 0 auto",

          alignSelf: "flex-start",
          opacity: disabled ? 0.45 : 1,
          borderRadius: 3,
          border: `1.5px solid ${active ? "#4ea1ff" : "#9a9aa6"}`,
          background: active ? "#4ea1ff" : "transparent",
          display: "grid",
          placeItems: "center",
          color: "#04101f",
          fontSize: 10,
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
