import React, { useCallback, useLayoutEffect, useRef, useState } from "react";
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

const OPEN_WIDTH_RATIO = 1.15;

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

  const pillRef = useRef<HTMLDivElement>(null);
  const [closedWidth, setClosedWidth] = useState(0);

  useLayoutEffect(() => {
    const pill = pillRef.current;
    if (!pill || open) return;
    const measure = () => {
      const w = pill.getBoundingClientRect().width;
      if (w > 0) setClosedWidth(w);
    };
    measure();
    const ro = new ResizeObserver(measure);
    ro.observe(pill);
    return () => ro.disconnect();
  }, [open]);

  const openWidth = Math.round(closedWidth * OPEN_WIDTH_RATIO);

  const wide = open && closedWidth > 0;

  return (

    <div
      style={{
        ...PILL_ANCHOR_STYLE,
        boxSizing: "border-box",

        ...(wide ? { width: openWidth, marginLeft: closedWidth - openWidth } : null),
      }}
    >
      {}
      <div
        ref={pillRef}
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
