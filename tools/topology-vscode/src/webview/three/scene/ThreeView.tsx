import { useEffect, useRef, useState } from "react";
import { Canvas } from "@react-three/fiber";
import * as THREE from "three";
import { sendRawInput, buildDeleteRaw, buildKeyRaw } from "../interaction/raw-input";
import { rulesDraftOpen } from "../../../../PolarRulesPanel/draw-rules-panel";
import { useInteractionControls } from "../interaction/interaction-controls";
import type { PickFn } from "../interaction/pick-types";
import { Scene } from "./scene-content";
import { BufferScene, BufferLabelProjector } from "./buffer-scene";
import { ProceduralEnvProvider } from "./scene-env";
import { NavGuides } from "../nav/NavGuides";
import { useOverlayFlags } from "../controls/flags/overlay-flags";
import { useBufferLabelPositions } from "./labels/use-buffer-label-positions";
import { BufferLabelOverlay } from "./labels/BufferLabelOverlay";

export function ThreeView() {

  const [bufferLabelPositions, onBufferPositions] = useBufferLabelPositions();

  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null);
  const pickRequest = useRef<PickFn | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const captureRef = useRef<HTMLDivElement | null>(null);
  const [canvasSize, setCanvasSize] = useState({ w: 800, h: 600 });

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const obs = new ResizeObserver(() => setCanvasSize({ w: el.clientWidth, h: el.clientHeight }));
    obs.observe(el);
    setCanvasSize({ w: el.clientWidth, h: el.clientHeight });
    return () => obs.disconnect();
  }, []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      if (rulesDraftOpen()) {
        e.preventDefault();
        sendRawInput(buildKeyRaw(e.key));
        return;
      }
      if (e.key !== "Delete" && e.key !== "Backspace") return;
      sendRawInput(buildDeleteRaw());
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const { onPointerDown, onPointerMove, onPointerUp, onPointerCancel, onWheelNative } = useInteractionControls(
    cameraRef,
    pickRequest,
  );

  useEffect(() => {
    const el = captureRef.current;
    if (!el) return;
    el.addEventListener("wheel", onWheelNative, { passive: false });
    return () => el.removeEventListener("wheel", onWheelNative);
  }, [onWheelNative]);

  const bufFlags = useOverlayFlags();
  const bufLabelsHidden = bufFlags?.labelsGlobal ?? false;

  return (
    <div ref={containerRef} style={{ position: "absolute", inset: 0 }}>
      {}
      <div
        ref={captureRef}
        style={{ position: "absolute", inset: 0, touchAction: "none" }}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerCancel={onPointerCancel}
        onContextMenu={(e) => e.preventDefault()}
      >
        <Canvas
          camera={{ fov: 50, near: 0.1, far: 20000, position: [0, 0, 500] }}
          gl={{ antialias: true }}
          style={{ position: "absolute", inset: 0 }}
          frameloop="always"
        >
          <Scene
            onPickRequest={pickRequest}
          />
          {}
          <NavGuides />
          {}
          <ProceduralEnvProvider>
            <BufferScene cameraRef={cameraRef} />
          </ProceduralEnvProvider>
          <BufferLabelProjector onPositions={onBufferPositions} />
        </Canvas>
      </div>

      {}
      {!bufLabelsHidden && <BufferLabelOverlay positions={bufferLabelPositions} />}

    </div>
  );
}
