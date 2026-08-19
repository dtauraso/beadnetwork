import { useCallback, useEffect, useRef, useState } from "react";
import type { BufferLabelPos } from "../buffer-scene";
import { probeLabelLag } from "./label-lag-probe";

export function useBufferLabelPositions(): [BufferLabelPos[], (positions: BufferLabelPos[]) => void] {
  const [bufferLabelPositions, setBufferLabelPositions] = useState<BufferLabelPos[]>([]);
  const bufferLabelRaf = useRef<ReturnType<typeof requestAnimationFrame> | null>(null);
  const pendingBufferPositions = useRef<BufferLabelPos[]>([]);
  const committedRef = useRef<BufferLabelPos[]>([]);

  const onBufferPositions = useCallback((positions: BufferLabelPos[]) => {
    probeLabelLag(committedRef.current, positions);
    pendingBufferPositions.current = positions;
    if (bufferLabelRaf.current === null) {
      bufferLabelRaf.current = requestAnimationFrame(() => {
        committedRef.current = pendingBufferPositions.current;
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
