import { useEffect, useMemo, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { drawSpeedPanel, speedPanelKey } from "../SliderPanel/draw-speed-panel";
import { drawTiltPanel, tiltPanelKey } from "../TiltPanel/draw-tilt-panel";
import { columnF32 } from "../Buffer/column-values";
import {
  COL_STREAM_SPEED_PANEL_BOX_Y, COL_STREAM_SPEED_PANEL_BOX_H,
  COL_STREAM_TILT_PANEL_BOX_Y, COL_STREAM_TILT_PANEL_BOX_H,
} from "../Buffer/column-streams-gen";

const STACK_GAP = 6;

function panelStackBottom(): number {
  const speed = columnF32(COL_STREAM_SPEED_PANEL_BOX_Y) + columnF32(COL_STREAM_SPEED_PANEL_BOX_H);
  const tilt = columnF32(COL_STREAM_TILT_PANEL_BOX_Y) + columnF32(COL_STREAM_TILT_PANEL_BOX_H);
  return Math.max(speed, tilt) + STACK_GAP;
}

export function PanelOverlay() {
  const { size, gl } = useThree();
  const canvas = useMemo(() => document.createElement("canvas"), []);
  const texRef = useRef<THREE.CanvasTexture | null>(null);
  const sceneRef = useRef(new THREE.Scene());
  const camRef = useRef(new THREE.OrthographicCamera(-1, 1, 1, -1, -1, 1));
  const lastKey = useRef("");

  const dpr = Math.max(1, Math.min(3, window.devicePixelRatio || 1));

  useEffect(() => {
    const tex = new THREE.CanvasTexture(canvas);
    tex.minFilter = THREE.LinearFilter;
    tex.magFilter = THREE.LinearFilter;
    texRef.current = tex;
    const mesh = new THREE.Mesh(
      new THREE.PlaneGeometry(2, 2),
      new THREE.MeshBasicMaterial({ map: tex, transparent: true, depthTest: false }),
    );
    camRef.current.updateProjectionMatrix();
    const overlayScene = sceneRef.current;
    overlayScene.add(mesh);
    return () => {
      overlayScene.remove(mesh);
      mesh.geometry.dispose();
      (mesh.material as THREE.Material).dispose();
      tex.dispose();
    };
  }, [canvas]);

  useFrame(({ scene, camera }) => {
    gl.render(scene, camera);

    const tex = texRef.current;
    if (!tex) return;

    const key = `${size.width}x${size.height}@${dpr}|${speedPanelKey()}|${tiltPanelKey()}`;
    if (key !== lastKey.current) {
      lastKey.current = key;
      canvas.width = Math.max(1, Math.round(size.width * dpr));
      canvas.height = Math.max(1, Math.round(size.height * dpr));
      const c = canvas.getContext("2d");
      if (c) {
        c.setTransform(dpr, 0, 0, dpr, 0, 0);
        c.clearRect(0, 0, size.width, size.height);
        drawSpeedPanel(c);
        drawTiltPanel(c);
      }
      document.documentElement.style.setProperty(
        "--panel-stack-bottom",
        `${Math.round(panelStackBottom())}px`,
      );
      tex.needsUpdate = true;
    }

    gl.autoClear = false;
    gl.clearDepth();
    gl.render(sceneRef.current, camRef.current);
    gl.autoClear = true;
  }, 1);

  return null;
}
