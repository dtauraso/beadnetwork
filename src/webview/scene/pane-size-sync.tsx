import { useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import { postLog } from "../log/post";

// PLACEMENT: src/webview/scene/ — mounted inside the Canvas by ThreeView, which owns

const APPEARS_COLLAPSED_W = 300;
const APPEARS_COLLAPSED_H = 150;

export function PaneSizeSync() {
  const setSize = useThree((s) => s.setSize);
  const size = useThree((s) => s.size);
  const lastReport = useRef("");

  useFrame(() => {
    const doc = document.documentElement;
    const w = Math.max(1, doc.clientWidth);
    const h = Math.max(1, doc.clientHeight);
    if (w === Math.round(size.width) && h === Math.round(size.height)) return;
    if (w <= APPEARS_COLLAPSED_W && h <= APPEARS_COLLAPSED_H) return;

    const site = `${size.width}x${size.height}->${w}x${h}`;
    if (site !== lastReport.current) {
      lastReport.current = site;
      const collapsed = Math.round(size.width) === APPEARS_COLLAPSED_W
        && Math.round(size.height) === APPEARS_COLLAPSED_H;
      postLog("pane-size-corrected", {
        was: `${size.width}x${size.height}`,
        pane: `${w}x${h}`,
        factor: Number((w / Math.max(1, size.width)).toFixed(3)),
        note: collapsed
          ? "renderer was at the intrinsic canvas size; nothing had measured the pane"
          : "renderer size disagreed with the pane it is pinned to",
      });
    }
    setSize(w, h);
  });

  return null;
}
