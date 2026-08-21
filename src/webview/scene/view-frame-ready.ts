import { getLatestViewFrame } from "../snapshot-buffer";
import { decodeViewFrame } from "../decode/buffer-decode-view";

export function viewFrameReady(): boolean {
  const buf = getLatestViewFrame();
  return buf !== null && decodeViewFrame(buf) !== null;
}
