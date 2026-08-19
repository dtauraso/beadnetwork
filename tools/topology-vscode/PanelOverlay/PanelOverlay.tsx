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

    // Take the rectangle from the canvas element itself. Its drawing buffer is exactly what
    // gets rendered and its client box is exactly what the viewer sees, so the two cannot
    // disagree with each other the way a reported size can. A bitmap sized from anything else
    // is stretched onto a viewport it was not drawn for, which moves the panels down the
    // screen and enlarges them — the further off the size, the further down they go.
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
      // The bitmap IS the drawing buffer, so the overlay quad maps it one to one.
      if (canvas.width !== bw || canvas.height !== bh) {
        canvas.width = bw;
        canvas.height = bh;
      }
      const c = canvas.getContext("2d");
      if (c) {
        // Go lays the panels out in CSS pixels; this is the only place that becomes device
        // pixels, and it is derived from the element rather than assumed.
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
