import { getLatestViewFrame, subscribeViewFrame } from "../../snapshot-buffer";
import { decodeViewFrame } from "../decode/buffer-decode-view";

export interface ViewBlocks {
  cameraView: DataView;
  overlayView: DataView;

  ringSurfacePointsView: DataView;
  beadRingSurfacePointsView: DataView;

  sceneTabs: string[];
  sceneTabSelected: number;
}

export function getViewBlocks(): ViewBlocks | null {
  const viewBuf = getLatestViewFrame();
  if (!viewBuf) return null;
  return decodeViewFrame(viewBuf);
}

export function subscribeViewBlocks(fn: () => void): () => void {
  return subscribeViewFrame(fn);
}
