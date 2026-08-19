import { useEffect, useMemo, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { drawSpeedPanel, speedPanelKey } from "../SliderPanel/draw-speed-panel";
import { drawTiltPanel, tiltPanelKey } from "../TiltPanel/draw-tilt-panel";
import { drawAnglePill, anglePillKey } from "../AngleDropdown/draw-angle-pill";
import { drawNodesPill, nodesPillKey } from "../NodesDropdown/draw-nodes-pill";
import { drawOverlaysPill, overlaysPillKey } from "../OverlaysDropdown/draw-overlays-pill";
import { drawFitChip, fitChipKey } from "../FitButton/draw-fit-chip";
import { drawTabStrip, tabStripKey } from "../Tabs/draw-tab-strip";
import { drawRulesPanel, rulesPanelKey } from "../PolarRulesPanel/draw-rules-panel";
import { postGoRecord } from "../src/webview/vscode-api";
import { encodeSceneViewport } from "../src/schema/input/input-encode-scene-tilt";
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
  const { gl } = useThree();
  const canvas = useMemo(() => document.createElement("canvas"), []);
  const texRef = useRef<THREE.CanvasTexture | null>(null);
  const sceneRef = useRef(new THREE.Scene());
  const camRef = useRef(new THREE.OrthographicCamera(-1, 1, 1, -1, -1, 1));
  const lastKey = useRef("");
  const lastSize = useRef({ w: 0, h: 0 });

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

    const el = gl.domElement;
    const vw = Math.max(1, el.clientWidth);
    const vh = Math.max(1, el.clientHeight);
    const bw = Math.max(1, el.width);
    const bh = Math.max(1, el.height);
    const scaleX = bw / vw;
    const scaleY = bh / vh;

    if (vw !== lastSize.current.w || vh !== lastSize.current.h) {
      lastSize.current = { w: vw, h: vh };
      postGoRecord(encodeSceneViewport(vw, vh));
    }

    const key = `${vw}x${vh}@${bw}x${bh}|${speedPanelKey()}|${tiltPanelKey()}|${anglePillKey()}|${nodesPillKey()}|${overlaysPillKey()}|${fitChipKey()}|${tabStripKey()}|${rulesPanelKey()}`;
    if (key !== lastKey.current) {
      lastKey.current = key;
      if (canvas.width !== bw || canvas.height !== bh) {
        canvas.width = bw;
        canvas.height = bh;
      }
      const c = canvas.getContext("2d");
      if (c) {
        c.setTransform(scaleX, 0, 0, scaleY, 0, 0);
        c.clearRect(0, 0, vw, vh);
        drawSpeedPanel(c);
        drawTiltPanel(c);
        drawAnglePill(c);
        drawNodesPill(c);
        drawOverlaysPill(c);
        drawFitChip(c);
        drawTabStrip(c);
        drawRulesPanel(c);
        // TEMPORARY: the four numbers the overlay's mapping depends on, drawn where a
        // screenshot can read them. Remove once the panels sit still.
        c.setTransform(1, 0, 0, 1, 0, 0);
        c.fillStyle = "#f00";
        c.font = "16px monospace";
        c.textAlign = "left";
        c.textBaseline = "top";
        c.fillText(
          `client=${vw}x${vh} buffer=${bw}x${bh} dpr=${window.devicePixelRatio} inner=${window.innerWidth}x${window.innerHeight}`,
          4, 4,
        );
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
