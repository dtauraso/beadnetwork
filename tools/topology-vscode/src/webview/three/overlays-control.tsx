// overlays-control.tsx — OVERLAYS CONTROL: split-button (body = master toggle, caret =
// popover) + popover checklist, composed from the grouped overlay defs.

import React, { useCallback, useState } from "react";
import { fireToggle, useToggleVal } from "./overlay-toggle";
import { guidelinesCfg, OVERLAY_GROUPS } from "./overlay-defs";
import { OverlayGroupSection } from "./overlay-group-section";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  inFlowPopoverStyle,
  PILL_ANCHOR_STYLE,
} from "./overlay-chrome";

/** OVERLAYS CONTROL: split-button (body = master toggle, caret = popover) + popover checklist. */
export function OverlaysControl() {
  const [open, setOpen] = useState(false);
  const val = useToggleVal(guidelinesCfg);
  const active = guidelinesCfg.active(val);

  const onBodyClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      fireToggle(guidelinesCfg, val);
    },
    [val]
  );

  const onCaretClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  }, []);

  return (
    // The pill and the popover are SIBLINGS inside the shared-width anchor, so they always
    // come out the same width (PILL_ANCHOR_STYLE). Never nest the popover in the pill:
    // pillContainerStyle sets `overflow: hidden` to clip the chip's rounded corners, which
    // would clip the popover out of existence.
    <div style={PILL_ANCHOR_STYLE}>
      {/* Split button — labeled pill (body = master toggle, caret = popover). Accent fill
          when the master is on, neutral chip when off (overlay-toggle-options.html mock). */}
      <div
        style={{
          // Placed LAST in ThreeView's right-hand flex column, so it sits below every
          // panel above it however tall they are — the stacking is the column's, not a
          // number here that has to be re-derived whenever a panel above grows.
          ...pillContainerStyle(active),
        }}
      >
        {/* Body — master toggle. `flex: "1 1 auto"` so the LABEL takes the pill's slack and
            the caret stays at the far end, the same as the angles pill. Without it the
            caret sits right after the word and slides whenever the shared width changes —
            which is what made it look like the triangle was following the popover. */}
        <div
          onClick={onBodyClick}
          title={guidelinesCfg.title(active)}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Overlays
        </div>
        {/* Caret — popover toggle */}
        <div
          onClick={onCaretClick}
          title={open ? "Close overlay list" : "Open overlay list"}
          style={pillCaretStyle}
        >
          {/* Same disclosure triangles as the group headings (see OverlayGroupSection). */}
          {open ? "▲" : "▼"}
        </div>
      </div>

      {/* Popover — grouped checklist (.pop mock style: panel2 bg, border, shadow). IN FLOW
          under the pill, filling the anchor's width — not the old absolute popover at a
          fixed 150. In flow it is measured, so the anchor sizes to whichever is wider (the
          pill's label or a group heading) and BOTH come out at that width. The rows measure
          as nothing (REVEALED_LIST_STYLE), so expanding a group still changes neither. */}
      {open && (
        <div style={inFlowPopoverStyle()}>
          {/* No dimming for the master-off state: the PILL is that indicator — unlit chip
              means overlays are off — and a second, fainter copy of the same fact inside the
              popover only made the list hard to read. The rows stay inert (`disabled`), they
              just no longer fade to say so; a checkmark still shows each flag's real value. */}
          {OVERLAY_GROUPS.map((group) => (
            <OverlayGroupSection key={group.heading} group={group} disabled={!active} />
          ))}
        </div>
      )}
    </div>
  );
}
