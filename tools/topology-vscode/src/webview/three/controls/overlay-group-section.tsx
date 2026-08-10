// overlay-group-section.tsx — one collapsible group in the popover: a clickable heading
// (with on/total count, itself a group-wide toggle) that expands to its OverlayRows.

import React, { useCallback, useState } from "react";
import { fireToggle, toggleVal } from "./overlay-toggle";
import { useOverlayFlags } from "./overlay-flags";
import type { OverlayGroup } from "./overlay-defs";
import { OverlayRow } from "./overlay-row";
import {
  groupHeadingStyle,
  DISCLOSURE_GLYPH_STYLE,
  CHROME_TEXT,
  REVEALED_LIST_STYLE,
} from "./overlay-chrome";

/** One collapsible group in the popover: a clickable heading that expands to its rows.
 *
 *  Collapsed is the DEFAULT, and the heading carries an on/total count so collapsing never
 *  hides state — you can read "POLES 2/3" without expanding, which is the thing a plain
 *  dropdown would cost you. Open/closed is view-local `useState`, deliberately NOT a Go
 *  flag: which section a person has twirled open is not part of the model (no buffer
 *  column, nothing streamed, nothing persisted), unlike the overlay flags themselves, which
 *  stay Go-owned. Each ROW still reads its own flag from the buffer — the count here is a
 *  second reader of the same truth, never a cache of it. */
export function OverlayGroupSection({ group, disabled }: { group: OverlayGroup; disabled?: boolean }) {
  const [open, setOpen] = useState(false);
  const [hover, setHover] = useState(false);
  const [countHover, setCountHover] = useState(false);
  const bufFlags = useOverlayFlags();
  const on = group.cfgs.filter((cfg) => cfg.active(toggleVal(bufFlags, cfg))).length;
  // Flip only the members that need flipping — every send is the SAME per-flag toggle record
  // a row click sends (encodeOverlaysToggle), so the group action introduces no second way to
  // set an overlay. Members already in the target state are left alone rather than toggled
  // twice. stopPropagation keeps this off the heading's expand/collapse.
  const onCountClick = useCallback(
    (e: React.MouseEvent) => {
      // The chip is the group's toggle ONLY while overlays are on. With the master off it
      // has no toggle to give, so it does NOT stop the click: it falls through to the
      // heading and expands the group like the words and the triangle beside it. Swallowing
      // it (the old `stopPropagation` before this check) made one part of the heading dead
      // to a click the rest of it answers.
      if (disabled) return;
      e.stopPropagation();
      const target = on === 0; // all off → turn everything on; otherwise turn everything off
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
        {/* The count is also the group's toggle, and it is SYMMETRIC with no remembered
            state: any member on → turn them all off; all off → turn them all on. The
            tempting version ("off, then restore what was on") needs to remember which
            members were on — a cache of Go-owned flags in TS, or a new per-group flag in
            Go. Neither is worth it for something three row clicks already do, so this
            sends nothing but the per-flag toggle records the rows themselves send.
            Accented only when some member is on, so a collapsed group reads at a glance. */}
        <span
          onClick={onCountClick}
          onMouseEnter={() => setCountHover(true)}
          onMouseLeave={() => setCountHover(false)}
          title={disabled ? "" : on > 0 ? `Turn all ${group.heading} off` : `Turn all ${group.heading} on`}
          style={{
            // Accent when some member is on; otherwise the chrome's own text colour. It was
            // #6e6e78 — dimmer than anything else in the popover, so an all-off group's count
            // was the hardest thing to read in it, when "0/3" is exactly what you look for.
            color: on > 0 ? "#4ea1ff" : CHROME_TEXT,
            fontVariantNumeric: "tabular-nums",
            // Pointer either way: with the master off this chip is part of the heading's
            // expand target, so a default cursor here would say "nothing to click" over
            // something that does answer a click.
            cursor: "pointer",
            padding: "1px 4px",
            borderRadius: 4,
            background: !disabled && countHover ? "rgba(255,255,255,0.10)" : "transparent",
          }}
        >
          {on}/{group.cfgs.length}
        </span>
      </div>
      {open && (
        <div style={REVEALED_LIST_STYLE}>
          {group.cfgs.map((cfg) => {
            const parent = group.under?.[cfg.flag];
            // A nested row is dead while its parent is off — the same rule the master
            // `overlays` flag applies to every row, one level down. Read through toggleVal,
            // the same rule a row itself reads by, so parent and child cannot disagree about
            // the parent's value.
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
