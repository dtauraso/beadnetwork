import { useCallback, useEffect, useRef, useState } from "react";
import type { BufferLabelPos } from "./buffer-scene";

export function useBufferLabelPositions(): [BufferLabelPos[], (positions: BufferLabelPos[]) => void] {
  const [bufferLabelPositions, setBufferLabelPositions] = useState<BufferLabelPos[]>([]);
  const bufferLabelRaf = useRef<ReturnType<typeof requestAnimationFrame> | null>(null);
  const pendingBufferPositions = useRef<BufferLabelPos[]>([]);

  const onBufferPositions = useCallback((positions: BufferLabelPos[]) => {
    pendingBufferPositions.current = positions;
    if (bufferLabelRaf.current === null) {
      bufferLabelRaf.current = requestAnimationFrame(() => {
        setBufferLabelPositions(pendingBufferPositions.current);
        bufferLabelRaf.current = null;
      });
    }
  }, []);

  useEffect(() => {
    return () => {
      if (bufferLabelRaf.current !== null) {
        cancelAnimationFrame(bufferLabelRaf.current);
        bufferLabelRaf.current = null;
      }
    };
  }, []);

  return [bufferLabelPositions, onBufferPositions];
}
