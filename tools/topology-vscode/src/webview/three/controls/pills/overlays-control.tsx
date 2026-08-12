


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




    <div style={PILL_ANCHOR_STYLE}>
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
        <div style={inFlowPopoverStyle()}>
          {}
          {OVERLAY_GROUPS.map((group) => (
            <OverlayGroupSection key={group.heading} group={group} disabled={!active} />
          ))}
        </div>
      )}
    </div>
  );
}
