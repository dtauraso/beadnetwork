import { useEffect, useRef, useState, useCallback } from "react";
import { Canvas } from "@react-three/fiber";
import * as THREE from "three";
import { HomeButton } from "../controls/panels/home-button";
import { OverlaysControl } from "../controls/pills/overlays-control";
import { NodePalette, dropKindFromEvent, fireCreateAt } from "../controls/panels/NodePalette";
import { DistanceHomePanel } from "../controls/panels/DistanceHomePanel";
import { TiltVectorAnglePanel } from "../controls/panels/TiltVectorAnglePanel";
import { SceneTabs } from "../controls/panels/SceneTabs";
import { useInteractionControls } from "../interaction/interaction-controls";
import type { PickFn } from "../interaction/pick-types";
import { Scene } from "./scene-content";
import { BufferScene, BufferLabelProjector } from "./buffer-scene";
import { ProceduralEnvProvider } from "./scene-env";
import type { BufferLabelPos } from "./buffer-scene";
import { NavGuides } from "../nav/NavGuides";
import { useOverlayFlags } from "../controls/flags/overlay-flags";

const PILL_STYLE: React.CSSProperties = {
  background: "rgba(0,0,0,0.55)",
  border: "none",
  borderRadius: 4,
  padding: "3px 6px",
};

export function ThreeView() {

  const [bufferLabelPositions, setBufferLabelPositions] = useState<BufferLabelPos[]>([]);

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
      {!bufLabelsHidden && bufferLabelPositions.map((pos) => (
        <div
          key={pos.row}
          style={{
            position: "absolute",
            left: pos.px,
            top: pos.py - 4,
            transform: "translate(-50%, -100%)",
            fontSize: 11,
            fontFamily: "monospace",
            color: "#e0e0e0",
            pointerEvents: "none",
            lineHeight: 1.25,
            textAlign: "center",
            zIndex: 10,
            ...PILL_STYLE,
          }}
        >
          <div style={{ whiteSpace: "nowrap" }}>{pos.label || String(pos.row)}</div>
        </div>
      ))}

      {}
      <div
        style={{
          position: "absolute",
          top: 44,
          right: 12,
          zIndex: 20,
          display: "flex",
          flexDirection: "column",

          alignItems: "stretch",
          gap: 6,
          pointerEvents: "none",
        }}
      >
        {}
        <HomeButton cameraRef={cameraRef} aspect={canvasSize.w / canvasSize.h} />
        <DistanceHomePanel />
        <TiltVectorAnglePanel />
        {}
        <NodePalette />
        <OverlaysControl />
      </div>
      <SceneTabs />
    </div>
  );
}
