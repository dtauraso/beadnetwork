import { setLabelLayer } from "./label-elements";

export function BufferLabelOverlay() {
  return (
    <div
      ref={setLabelLayer}
      style={{ position: "absolute", inset: 0, pointerEvents: "none" }}
    />
  );
}
