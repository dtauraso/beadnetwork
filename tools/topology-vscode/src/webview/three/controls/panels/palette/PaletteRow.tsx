import React, { useRef, useState } from "react";
import { NODE_DEFS } from "../../../../../schema/node-defs";
import { NODE_PALETTE_KIND_MIME } from "./node-palette-drag";
import { popoverRowStyle, DISCLOSURE_GLYPH_STYLE, REVEALED_LIST_STYLE } from "../../pills/overlay-chrome";

export function PaletteRow({ kind, kindId }: { kind: string; kindId: number }) {
  const [hover, setHover] = useState(false);

  const [open, setOpen] = useState(false);

  const headingRef = useRef<HTMLSpanElement>(null);
  const def = NODE_DEFS[kind];
  return (
    <div
      draggable
      onDragStart={(e) => {
        e.dataTransfer.setData(NODE_PALETTE_KIND_MIME, String(kindId));
        e.dataTransfer.effectAllowed = "copy";

        if (headingRef.current) {
          e.dataTransfer.setDragImage(headingRef.current, 12, 8);
        }
      }}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{ ...popoverRowStyle(hover, false), cursor: "grab", flexDirection: "column", alignItems: "stretch", gap: 2 }}
      title={`Drag ${kind} onto the scene`}
    >
      {}
      <span
        ref={headingRef}
        style={{ display: "flex", alignItems: "center", gap: 7 }}
        onClick={(e) => {
          e.stopPropagation();
          setOpen((o) => !o);
        }}
      >
        <span style={DISCLOSURE_GLYPH_STYLE}>{open ? "▼" : "▶"}</span>
        {}
        <span
          style={{
            width: 11,
            height: 11,
            flex: "0 0 auto",
            borderRadius: 3,
            background: def?.fill ?? "#888",
            border: `1px solid ${def?.stroke ?? "#888"}`,
          }}
        />
        <span style={{ minWidth: 0, overflowWrap: "anywhere" }}>{kind}</span>
      </span>
      {}
      {open && def?.desc && (
        <div style={REVEALED_LIST_STYLE}>
          <span
            style={{
              display: "block",
              opacity: 0.8,

              overflowWrap: "anywhere",
              whiteSpace: "normal",
              lineHeight: 1.35,
              paddingLeft: 18,
            }}
          >
            {def.desc}
          </span>
        </div>
      )}
    </div>
  );
}
