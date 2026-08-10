// home-button.tsx — HOME BUTTON: reframes the camera to fit all nodes in view.

import React, { useCallback } from "react";
import * as THREE from "three";
import { sendRawInput, buildHomeRaw } from "../../interaction/raw-input";

/** HOME BUTTON: reframes the camera to fit all nodes in view. */
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
    // Home is a COMMAND to Go. TS sends only render context (fov + aspect); Go frames the
    // scene from its OWN node geometry, installs the pose in the gesture FSM, and streams it
    // back via the buffer's Camera row (BufferCamera). Because the FSM's own pose becomes the
    // framed pose, the next pan/zoom/rotate builds on it (no snap-back). We do NOT mutate the
    // three.js camera or seed a computed pose here.
    sendRawInput(buildHomeRaw(cam.fov, aspect));
  }, [cameraRef, aspect]);

  return (
    <div
      onClick={onClick}
      title="Fit diagram in view"
      style={{
        // Placed by ThreeView's right-hand flex column, not by its own top/right. That
        // column stretches its widgets to one width so the pills match each other; this
        // is not a pill, so it opts out and stays as wide as "⌂ fit".
        alignSelf: "flex-end",
        background: "rgba(0,0,0,0.55)",
        borderRadius: 6,
        padding: "3px 7px",
        cursor: "pointer",
        pointerEvents: "auto",
        zIndex: 20,
        color: "#ddd",
        fontSize: 11,
        fontFamily: "monospace",
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
