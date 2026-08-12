import React, { useCallback, useState } from "react";
import { fireToggle, toggleVal } from "./overlay-toggle";
import { useOverlayFlags } from "../flags/overlay-flags";
import type { OverlayGroup } from "./overlay-defs";
import { OverlayRow } from "./overlay-row";
import {
  groupHeadingStyle,
  DISCLOSURE_GLYPH_STYLE,
  CHROME_TEXT,
  REVEALED_LIST_STYLE,
} from "./overlay-chrome";
import * as T from "../chrome-theme";

export function OverlayGroupSection({ group, disabled }: { group: OverlayGroup; disabled?: boolean }) {
  const [open, setOpen] = useState(false);
  const [hover, setHover] = useState(false);
  const [countHover, setCountHover] = useState(false);
  const bufFlags = useOverlayFlags();
  const on = group.cfgs.filter((cfg) => cfg.active(toggleVal(bufFlags, cfg))).length;

  const onCountClick = useCallback(
    (e: React.MouseEvent) => {

      if (disabled) return;
      e.stopPropagation();
      const target = on === 0; 
      for (const cfg of group.cfgs) {
        const val = toggleVal(bufFlags, cfg);
        if (cfg.active(val) !== target) fireToggle(cfg, val);
      }
    },
    [group, bufFlags, on, disabled]
  );
  return (
    <div>
      <div
        onClick={(e) => { e.stopPropagation(); setOpen((o) => !o); }}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        title={open ? `Collapse ${group.heading}` : `Expand ${group.heading}`}
        style={groupHeadingStyle(hover)}
      >
        <span style={DISCLOSURE_GLYPH_STYLE}>{open ? "▼" : "▶"}</span>
        <span style={{ flex: "1 1 auto" }}>{group.heading}</span>
        {}
        <span
          onClick={onCountClick}
          onMouseEnter={() => setCountHover(true)}
          onMouseLeave={() => setCountHover(false)}
          title={disabled ? "" : on > 0 ? `Turn all ${group.heading} off` : `Turn all ${group.heading} on`}
          style={{

            color: on > 0 ? T.ACCENT : CHROME_TEXT,
            fontVariantNumeric: "tabular-nums",

            cursor: "pointer",
            padding: T.PAD_ITEM,
            borderRadius: T.RADIUS_ITEM,
            background: !disabled && countHover ? T.HOVER_CHIP : "transparent",
          }}
        >
          {on}/{group.cfgs.length}
        </span>
      </div>
      {open && (
        <div style={REVEALED_LIST_STYLE}>
          {group.cfgs.map((cfg) => {
            const parent = group.under?.[cfg.flag];

            const parentOff = !!parent && !parent.active(toggleVal(bufFlags, parent));
            return (
              <OverlayRow
                key={cfg.flag}
                cfg={cfg}
                disabled={disabled || parentOff}
                indent={!!parent}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}
