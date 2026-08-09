// overlay-row.tsx — a single popover row: square checkbox + icon + label, fires the row's
// toggle on click.

import React, { useCallback, useState } from "react";
import { fireToggle, useToggleVal, type ToggleCfg } from "./overlay-toggle";
import { popoverRowStyle } from "./overlay-chrome";

/** A single row inside the popover: square checkbox + label, fires the row's op on click.
 *  Styled to match the recommended mock (overlay-toggle-options.html): custom .cb checkbox
 *  that fills accent + ✓ when checked, with a subtle row-hover background. */
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
      {/* The checkbox is the ONLY thing that fades when the row is inert. The label stays
          full strength — fading it was what made the open list hard to read — and the pill
          still says whether the master gate is on. Here the fade is on the one element whose
          job is to be clicked, so it reads as "this box is not taking clicks" rather than as
          the whole row receding. */}
      <span
        style={{
          width: 13,
          height: 13,
          flex: "0 0 auto",
          // Beside the first line too, for the same reason as the icon below it.
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
      {/* The icon LEADS, in a column of its own — never inside the label's text run. A
          glyph sharing that run is just another word, so a label that wrapped put its
          second line under the icon; here the icon is a sibling the text never flows
          beneath, and the words wrap under the words. Fixed-width so every row's words
          start at one x whatever glyph precedes them. */}
      {/* `alignSelf: flex-start` so a wrapped row keeps its glyph beside the FIRST line
          rather than floating to the middle of two. Identical to the row's centering while
          the label is one line, which is every row until one wraps. */}
      <span style={{ width: 11, flex: "0 0 auto", textAlign: "center", alignSelf: "flex-start" }}>
        {iconText}
      </span>
      {/* Wraps rather than widening: the row lives in a box that measures as nothing, so a
          label longer than the popover has to break onto the next line or it would just
          overflow the edge. `minWidth: 0` is what lets a flex item shrink below its own
          content — without it the default `min-width: auto` keeps the label at full width
          and the break never happens. */}
      <span style={{ minWidth: 0, overflowWrap: "break-word" }}>{labelText}</span>
    </div>
  );
}
