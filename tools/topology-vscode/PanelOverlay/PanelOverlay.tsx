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
import { postLog } from "../src/webview/log/post";
import { columnF32 } from "../Buffer/column-values";
import {
  COL_STREAM_RULES_PANEL_BOX_X, COL_STREAM_RULES_PANEL_BOX_Y,
  COL_STREAM_RULES_PANEL_BOX_W, COL_STREAM_RULES_PANEL_BOX_H,
  COL_STREAM_OVERLAYS_PILL_PILL_X, COL_STREAM_OVERLAYS_PILL_PILL_Y,
  COL_STREAM_OVERLAYS_PILL_PILL_W, COL_STREAM_OVERLAYS_PILL_PILL_H,
} from "../Buffer/column-streams-gen";

const OVERLAY_SURFACE_H = 2048;

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
    const ratio = Math.max(1, Math.min(3, gl.getPixelRatio()));
    const bw = Math.max(1, Math.round(vw * ratio));
    const bh = Math.max(1, Math.round(OVERLAY_SURFACE_H * ratio));

    if (vw !== lastSize.current.w || vh !== lastSize.current.h) {
      const prev = lastSize.current;
      lastSize.current = { w: vw, h: vh };
      postGoRecord(encodeSceneViewport(vw, vh));

      const rect = (x: number, y: number, w: number, h: number) =>
        `${columnF32(x).toFixed(1)},${columnF32(y).toFixed(1)} ${columnF32(w).toFixed(1)}x${columnF32(h).toFixed(1)}`;
      const box = (n: Element | null) =>
        n ? `${Math.round(n.clientWidth)}x${Math.round(n.clientHeight)}` : "none";
      postLog("panel-size-on-resize", {
        viewWas: `${prev.w}x${prev.h}`,
        viewNow: `${vw}x${vh}`,
        renderTarget: `${el.width}x${el.height}`,
        pixelRatio: ratio,
        devicePixelRatio: window.devicePixelRatio,
        window: `${window.innerWidth}x${window.innerHeight}`,
        outer: `${window.outerWidth}x${window.outerHeight}`,
        visualViewport: window.visualViewport
          ? `${Math.round(window.visualViewport.width)}x${Math.round(window.visualViewport.height)} scale=${window.visualViewport.scale}`
          : "none",
        body: box(document.body),
        app: box(document.getElementById("app")),
        canvasParent: box(el.parentElement),
        canvasCss: `${Math.round(el.getBoundingClientRect().width)}x${Math.round(el.getBoundingClientRect().height)}`,
        rulesPanel: rect(
          COL_STREAM_RULES_PANEL_BOX_X, COL_STREAM_RULES_PANEL_BOX_Y,
          COL_STREAM_RULES_PANEL_BOX_W, COL_STREAM_RULES_PANEL_BOX_H,
        ),
        overlaysPill: rect(
          COL_STREAM_OVERLAYS_PILL_PILL_X, COL_STREAM_OVERLAYS_PILL_PILL_Y,
          COL_STREAM_OVERLAYS_PILL_PILL_W, COL_STREAM_OVERLAYS_PILL_PILL_H,
        ),
      });
    }

    const key = `${vw}@${bw}x${bh}|${speedPanelKey()}|${tiltPanelKey()}|${anglePillKey()}|${nodesPillKey()}|${overlaysPillKey()}|${fitChipKey()}|${tabStripKey()}|${rulesPanelKey()}`;
    if (key !== lastKey.current) {
      lastKey.current = key;
      if (canvas.width !== bw || canvas.height !== bh) {
        canvas.width = bw;
        canvas.height = bh;
      }
      const c = canvas.getContext("2d");
      if (c) {
        c.setTransform(1, 0, 0, 1, 0, 0);
        c.clearRect(0, 0, canvas.width, canvas.height);
        c.setTransform(ratio, 0, 0, ratio, 0, 0);
        drawSpeedPanel(c);
        drawTiltPanel(c);
        drawAnglePill(c);
        drawNodesPill(c);
        drawOverlaysPill(c);
        drawFitChip(c);
        drawTabStrip(c);
        drawRulesPanel(c);
      }
      tex.needsUpdate = true;
    }

    const targetW = Math.max(1, el.width);
    const targetH = Math.max(1, el.height);

    const uSpan = Math.min(1, targetW / canvas.width);
    const vSpan = Math.min(1, targetH / canvas.height);
    tex.repeat.set(uSpan, vSpan);
    tex.offset.set(0, 1 - vSpan);

    gl.autoClear = false;
    gl.clearDepth();
    gl.render(sceneRef.current, camRef.current);
    gl.autoClear = true;
  }, 1);

  return null;
}
