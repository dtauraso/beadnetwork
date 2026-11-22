import { useEffect } from "react";
import { useThree } from "@react-three/fiber";

let redraw: () => void = () => {};

export function requestRedraw(): void {
  if (!document.hidden) redraw();
}

export function RedrawOnChange() {
  const invalidate = useThree((s) => s.invalidate);
  const gl = useThree((s) => s.gl);
  useEffect(() => {
    redraw = () => invalidate();
    window.BEADNETWORK_BLOCKS?.on(requestRedraw);
    document.addEventListener("visibilitychange", requestRedraw);
    const ro = new ResizeObserver(requestRedraw);
    ro.observe(gl.domElement.parentElement ?? document.documentElement);
    requestRedraw();
    return () => {
      document.removeEventListener("visibilitychange", requestRedraw);
      ro.disconnect();
      redraw = () => {};
    };
  }, [invalidate, gl]);
  return null;
}
