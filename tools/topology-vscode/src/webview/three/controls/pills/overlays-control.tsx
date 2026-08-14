import React, { useCallback } from "react";
import { fireToggle, useToggleVal } from "./overlay-toggle";
import { firePanelToggle, usePanelOpen } from "./panel-toggle";
import { guidelinesCfg, OVERLAY_GROUPS } from "./overlay-defs";
import { OverlayGroupSection } from "./overlay-group-section";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  popoverStyle,
  PILL_ANCHOR_STYLE,
} from "./overlay-chrome";

const OPEN_WIDTH = "150%";

const OPEN_PULL_LEFT = "-50%";

const POPOVER_MAX_HEIGHT = "60vh";

export function OverlaysControl() {
  const open = usePanelOpen("overlays");
  const val = useToggleVal(guidelinesCfg);
  const active = guidelinesCfg.active(val);

  const onBodyClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      fireToggle(guidelinesCfg, val);
    },
    [val]
  );

  const onCaretClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      firePanelToggle("overlays", open);
    },
    [open]
  );

  return (

    <div
      style={{
        ...PILL_ANCHOR_STYLE,
        boxSizing: "border-box",

        position: "relative",

        ...(open ? { width: OPEN_WIDTH, marginLeft: OPEN_PULL_LEFT } : null),
      }}
    >
      {}
      <div
        style={{

          ...pillContainerStyle(active),
        }}
      >
        {}
        <div
          onClick={onBodyClick}
          title={guidelinesCfg.title(active)}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Overlays
        </div>
        {}
        <div
          onClick={onCaretClick}
          title={open ? "Close overlay list" : "Open overlay list"}
          style={pillCaretStyle}
        >
          {}
          {open ? "▲" : "▼"}
        </div>
      </div>

      {}
      {open && (
        <div
          style={{
            ...popoverStyle("100%"),
            boxSizing: "border-box",
            maxHeight: POPOVER_MAX_HEIGHT,
            overflowY: "auto",
          }}
        >
          {}
          {OVERLAY_GROUPS.map((group) => (
            <OverlayGroupSection key={group.heading} group={group} disabled={!active} />
          ))}
        </div>
      )}
    </div>
  );
}
