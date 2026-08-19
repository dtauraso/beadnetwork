import type { BufferLabelPos } from "../buffer-scene";
import * as T from "../../controls/chrome-theme";
import { registerLabelElement } from "./label-elements";

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
          ref={(el) => registerLabelElement(pos.row, el)}
          style={{
            position: "absolute",
            left: 0,
            top: 0,
            transform: `translate(${pos.px}px, ${pos.py - 4}px) translate(-50%, -100%)`,
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
