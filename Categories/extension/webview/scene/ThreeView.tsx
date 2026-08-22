import { useEffect, useRef } from "react";
import { Canvas } from "@react-three/fiber";
import * as THREE from "three";
import { buildDeleteRaw, buildKeyRaw } from "../../../Input/Drag/raw-input-build";
import { sendRawInput } from "../vscode-api";
import { rulesDraftOpen } from "../../../Chrome/Panels/PolarRulesPanel/draw-rules-panel";
import { useInteractionControls } from "./interaction-controls";
import type { PickFn } from "../../../Input/Drag/pick-types";
import { Scene } from "./scene-content";
import { SceneRoot, LabelProjector } from "./scene-root";
import { ProceduralEnvProvider } from "./scene-env";
import { NavGuides } from "../../../Camera/NavGuides";
import { PaneSizeSync } from "./pane-size-sync";

function consumedByDraft(key: string): boolean {
  return key.length === 1 || key === "Enter" || key === "Escape" || key === "Backspace";
}

export function ThreeView() {
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null);
  const pickRequest = useRef<PickFn | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const captureRef = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      if (e.ctrlKey || e.metaKey || e.altKey) return;
      if (rulesDraftOpen() && consumedByDraft(e.key)) {
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
        {/* width/height, not inset. A canvas is a REPLACED element: absolutely positioned
            with an auto width, CSS takes its INTRINSIC size and ignores the opposite
            offset, so `inset: 0` sized it 300x150 — which is the size the viewport
            breadcrumb reported. A percentage is the only thing that stretches it. */}
        <Canvas
          camera={{ fov: 50, near: 0.1, far: 20000, position: [0, 0, 500] }}
          gl={{ antialias: true }}
          style={{ position: "absolute", top: 0, left: 0, width: "100%", height: "100%" }}
          frameloop="always"
        >
          <PaneSizeSync />
          <Scene
            onPickRequest={pickRequest}
          />
          {}
          <NavGuides />
          {}
          <ProceduralEnvProvider>
            <SceneRoot cameraRef={cameraRef} />
          </ProceduralEnvProvider>
          <LabelProjector />
        </Canvas>
      </div>
      {/* Nothing outside the canvas. The labels were the last thing mounted here and are
          drawn now, so the whole editor is one surface. */}
    </div>
  );
}
