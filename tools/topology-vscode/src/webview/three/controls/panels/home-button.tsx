import React, { useCallback } from "react";
import * as THREE from "three";
import { sendRawInput, buildHomeRaw } from "../../interaction/raw-input";
import * as T from "../chrome-theme";

export function HomeButton({
  cameraRef,
  aspect,
}: {
  cameraRef: React.MutableRefObject<THREE.PerspectiveCamera | null>;
  aspect: number;
}) {
  const onClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    const cam = cameraRef.current;
    if (!cam) return;

    sendRawInput(buildHomeRaw(cam.fov, aspect));
  }, [cameraRef, aspect]);

  return (
    <div
      onClick={onClick}
      title="Fit diagram in view"
      style={{

        alignSelf: "flex-end",
        background: T.CHIP,
        border: `1px solid ${T.BORDER}`,
        borderRadius: T.RADIUS_CHIP,
        padding: T.PAD_CHIP,
        cursor: "pointer",
        pointerEvents: "auto",
        zIndex: 20,
        color: T.TEXT,
        fontSize: T.FONT_SIZE,
        fontFamily: T.FONT_STACK,
        userSelect: "none",
        display: "flex",
        alignItems: "center",
        gap: 4,
      }}
    >
      ⌂ fit
    </div>
  );
}
