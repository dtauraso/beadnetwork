import React, { useCallback, useState } from "react";
import { fireToggle, toggleVal } from "./overlay-toggle";
import { useOverlayFlags } from "../src/webview/three/controls/flags/overlay-flags";
import { firePanelToggle, usePanelOpen } from "./panel-toggle";
import { groupCfgs, type OverlayGroup } from "./overlay-defs";
import { OverlayRow } from "./overlay-row";
import {
  groupHeadingStyle,
  DISCLOSURE_GLYPH_STYLE,
  CHROME_TEXT,
  REVEALED_LIST_STYLE,
} from "../src/webview/three/controls/pills/overlay-chrome";
import * as T from "../src/webview/three/controls/chrome-theme";

export function OverlayGroupSection({
  group,
  disabled,
  depth = 0,
}: {
  group: OverlayGroup;
  disabled?: boolean;
  depth?: number;
}) {
  const open = usePanelOpen(group.panel);
  const [hover, setHover] = useState(false);
  const [countHover, setCountHover] = useState(false);
  const bufFlags = useOverlayFlags();
  const all = groupCfgs(group);
  const on = all.filter((cfg) => cfg.active(toggleVal(bufFlags, cfg))).length;

  const onCountClick = useCallback(
    (e: React.MouseEvent) => {

      if (disabled) return;
      e.stopPropagation();
      const target = on === 0;
      for (const cfg of all) {
        const val = toggleVal(bufFlags, cfg);
        if (cfg.active(val) !== target) fireToggle(cfg, val);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [group, bufFlags, on, disabled]
  );
  return (
    <div>
      <div
        onClick={(e) => { e.stopPropagation(); firePanelToggle(group.panel, open); }}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        title={open ? `Collapse ${group.heading}` : `Expand ${group.heading}`}
        style={{ ...groupHeadingStyle(hover), paddingLeft: 3 + depth * 10 }}
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
          {on}/{all.length}
        </span>
      </div>
      {open && (
        <div style={REVEALED_LIST_STYLE}>
          {group.cfgs.map((cfg) => (
            <OverlayRow key={cfg.flag} cfg={cfg} disabled={disabled} indent={depth} />
          ))}
          {(group.groups ?? []).map((sub) => (
            <OverlayGroupSection
              key={sub.heading}
              group={sub}
              disabled={disabled}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
}
