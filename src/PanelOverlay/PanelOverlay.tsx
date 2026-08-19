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
import {
  drawPointerHighlight, drawPointerTip, pointerTargetCursor, pointerTargetKey,
} from "./draw-pointer-target";
import { drawLabels, labelEpoch } from "../Scene/Labels/label-canvas";
import { postGoRecord } from "../webview/vscode-api";
import { postLog } from "../webview/log/post";
import { encodeSceneViewport } from "../schema/input/input-encode-scene-tilt";

const OVERLAY_SURFACE_W = 4096;
const OVERLAY_SURFACE_H = 4096;

const SCALE_EXACT = 0.02;

function keyField(key: string, index: number): number {
  const n = Number(key.split(",")[index]);
  return Number.isFinite(n) ? n : 0;
}

function reportPanelSize(
  name: string,
  goRectW: number,
  goRectH: number,
  vw: number,
  targetW: number,
  surfaceW: number,
  uSpan: number,
  ratio: number,
  first: { current: number },
  last: { current: string },
): void {
  if (goRectW <= 0) return;
  const surfacePxPerScreenPx = (surfaceW * uSpan) / targetW;
  const onScreenCss = (goRectW * ratio * surfacePxPerScreenPx) / (targetW / vw);
  if (first.current === 0) first.current = onScreenCss;
  const factor = onScreenCss / first.current;
  const status = Math.abs(factor - 1) <= SCALE_EXACT
    ? "held"
    : factor > 1 ? "stretched" : "shrunk";
  const site = `${status}|${factor.toFixed(3)}|${vw}`;
  if (site === last.current) return;
  last.current = site;
  postLog(`panel-overlay-${status}`, {
    panel: name,
    status,
    factor: Number(factor.toFixed(4)),
    onScreenCss: Number(onScreenCss.toFixed(2)),
    firstSeenCss: Number(first.current.toFixed(2)),
    goRect: `${goRectW.toFixed(2)}x${goRectH.toFixed(2)}`,
    viewCss: vw,
    targetW,
    surfaceW,
    uSpan: Number(uSpan.toFixed(4)),
    ratio: Number(ratio.toFixed(4)),
  });
}

export function PanelOverlay() {
  const { gl } = useThree();
  const canvas = useMemo(() => document.createElement("canvas"), []);
  const texRef = useRef<THREE.CanvasTexture | null>(null);
  const sceneRef = useRef(new THREE.Scene());
  const camRef = useRef(new THREE.OrthographicCamera(-1, 1, 1, -1, -1, 1));
  const lastKey = useRef("");
  const lastSize = useRef({ w: 0, h: 0 });
  const firstChip = useRef(0);
  const lastChip = useRef("");
  const firstSpeed = useRef(0);
  const lastSpeed = useRef("");

  useEffect(() => {
    const tex = new THREE.CanvasTexture(canvas);
    tex.colorSpace = THREE.SRGBColorSpace;
    tex.minFilter = THREE.LinearFilter;
    tex.magFilter = THREE.LinearFilter;
    tex.wrapS = THREE.ClampToEdgeWrapping;
    tex.wrapT = THREE.ClampToEdgeWrapping;
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
    const targetW = Math.max(1, el.width);
    const targetH = Math.max(1, el.height);
    const ratio = targetW / vw;
    const bw = OVERLAY_SURFACE_W;
    const bh = OVERLAY_SURFACE_H;

    if (vw !== lastSize.current.w || vh !== lastSize.current.h) {
      lastSize.current = { w: vw, h: vh };
      postGoRecord(encodeSceneViewport(vw, vh));
      postLog("panel-overlay-size", {
        clientCss: `${vw}x${vh}`,
        target: `${targetW}x${targetH}`,
        surface: `${bw}x${bh}`,
        canvasNow: `${canvas.width}x${canvas.height}`,
        ratioMeasured: ratio,
        ratioRenderer: gl.getPixelRatio(),
        devicePixelRatio: window.devicePixelRatio,
        styleCss: `${el.style.width}x${el.style.height}`,
      });
    }

    const key = `${vw}@${bw}x${bh}|${speedPanelKey()}|${tiltPanelKey()}|${anglePillKey()}|${nodesPillKey()}|${overlaysPillKey()}|${fitChipKey()}|${tabStripKey()}|${rulesPanelKey()}|${pointerTargetKey()}|${labelEpoch()}`;
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
        drawLabels(c);
        drawPointerHighlight(c);
        drawSpeedPanel(c);
        drawTiltPanel(c);
        drawAnglePill(c);
        drawNodesPill(c);
        drawOverlaysPill(c);
        drawFitChip(c);
        drawTabStrip(c);
        drawRulesPanel(c);
        drawPointerTip(c);
      }
      el.style.cursor = pointerTargetCursor();
      tex.needsUpdate = true;
    }

    if (targetW > OVERLAY_SURFACE_W || targetH > OVERLAY_SURFACE_H) {
      postLog("panel-overlay-surface-too-small", {
        target: `${targetW}x${targetH}`,
        surface: `${OVERLAY_SURFACE_W}x${OVERLAY_SURFACE_H}`,
        note: "the overlay sheet is smaller than the render target; the panels are being spread across it",
      });
    }

    const uSpan = Math.min(1, targetW / canvas.width);
    const vSpan = Math.min(1, targetH / canvas.height);
    tex.repeat.set(uSpan, vSpan);
    tex.offset.set(0, 1 - vSpan);

    const chipKey = fitChipKey();
    reportPanelSize(
      "fit-chip", keyField(chipKey, 2), 0,
      vw, targetW, canvas.width, uSpan, ratio, firstChip, lastChip,
    );
    const speedKey = speedPanelKey();
    reportPanelSize(
      "speed-panel", keyField(speedKey, 2), keyField(speedKey, 3),
      vw, targetW, canvas.width, uSpan, ratio, firstSpeed, lastSpeed,
    );

    gl.autoClear = false;
    gl.clearDepth();
    gl.render(sceneRef.current, camRef.current);
    gl.autoClear = true;
  }, 1);

  return null;
}
