import { useEffect, useRef, useState } from "react";
import { Canvas } from "@react-three/fiber";
import * as THREE from "three";
import { dropKindFromEvent, fireCreateAt } from "../controls/panels/node-palette-drag";
import { SceneTabs } from "../controls/panels/SceneTabs";
import { useInteractionControls } from "../interaction/interaction-controls";
import type { PickFn } from "../interaction/pick-types";
import { Scene } from "./scene-content";
import { BufferScene, BufferLabelProjector } from "./buffer-scene";
import { ProceduralEnvProvider } from "./scene-env";
import { NavGuides } from "../nav/NavGuides";
import { useOverlayFlags } from "../controls/flags/overlay-flags";
import { useBufferLabelPositions } from "./labels/use-buffer-label-positions";
import { BufferLabelOverlay } from "./labels/BufferLabelOverlay";
import { ScenePanelColumn } from "./ScenePanelColumn";

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

        onDragOver={(e) => {
          if (e.dataTransfer.types.includes("application/x-wirefold-kind")) e.preventDefault();
        }}
        onDrop={(e) => {
          const kindId = dropKindFromEvent(e.nativeEvent);
          if (kindId === null) return;
          e.preventDefault();
          const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
          fireCreateAt(
            kindId,
            ((e.clientX - r.left) / r.width) * 2 - 1,
            -(((e.clientY - r.top) / r.height) * 2 - 1),
          );
        }}
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

      {}
      <ScenePanelColumn cameraRef={cameraRef} aspect={canvasSize.w / canvasSize.h} />
      <SceneTabs />
    </div>
  );
}
