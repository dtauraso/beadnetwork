import { useCallback, useRef, useState } from "react";
import type { BufferLabelPos } from "../buffer-scene";
import { probeLabelLag } from "./label-lag-probe";
import { applyLabelPositions } from "./label-elements";

function identityKey(positions: BufferLabelPos[]): string {
  let key = "";
  for (const p of positions) key += `${p.row}:${p.label}|`;
  return key;
}

export function useBufferLabelPositions(): [BufferLabelPos[], (positions: BufferLabelPos[]) => void] {
  const [labels, setLabels] = useState<BufferLabelPos[]>([]);
  const lastIdentity = useRef("");

  const onBufferPositions = useCallback((positions: BufferLabelPos[]) => {
    probeLabelLag(positions);
    applyLabelPositions(positions);

    const key = identityKey(positions);
    if (key !== lastIdentity.current) {
      lastIdentity.current = key;
      setLabels(positions);
    }
  }, []);

  return [labels, onBufferPositions];
}
