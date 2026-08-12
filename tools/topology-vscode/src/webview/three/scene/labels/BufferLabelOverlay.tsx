import type { BufferLabelPos } from "../buffer-scene";
import * as T from "../../controls/chrome-theme";

const PILL_STYLE: React.CSSProperties = {
  background: T.CHIP,
  border: `1px solid ${T.BORDER}`,
  borderRadius: T.RADIUS_ITEM,
  padding: T.PAD_CHIP,
};

export function BufferLabelOverlay({ positions }: { positions: BufferLabelPos[] }) {
  return (
    <>
      {positions.map((pos) => (
        <div
          key={pos.row}
          style={{
            position: "absolute",
            left: pos.px,
            top: pos.py - 4,
            transform: "translate(-50%, -100%)",
            fontSize: T.FONT_SIZE,
            fontFamily: T.FONT_STACK,
            fontVariantNumeric: "tabular-nums",
            color: T.TEXT,
            pointerEvents: "none",
            lineHeight: 1.25,
            textAlign: "center",
            zIndex: 10,
            ...PILL_STYLE,
          }}
        >
          <div style={{ whiteSpace: "nowrap" }}>{pos.label || String(pos.row)}</div>
        </div>
      ))}
    </>
  );
}
