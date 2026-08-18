import { useEffect, useMemo, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { columnBytes, columnF32 } from "../Buffer/column-values";
import {
  COL_STREAM_SPEED_PANEL_RECT_X, COL_STREAM_SPEED_PANEL_RECT_Y,
  COL_STREAM_SPEED_PANEL_RECT_W, COL_STREAM_SPEED_PANEL_RECT_H,
  COL_STREAM_SPEED_PANEL_SELECTED,
  COL_STREAM_SPEED_PANEL_NUM_TEXT, COL_STREAM_SPEED_PANEL_NUM_LEN,
  COL_STREAM_SPEED_PANEL_DEN_TEXT, COL_STREAM_SPEED_PANEL_DEN_LEN,
  COL_STREAM_SPEED_PANEL_TRACK_X, COL_STREAM_SPEED_PANEL_TRACK_Y,
  COL_STREAM_SPEED_PANEL_TRACK_W, COL_STREAM_SPEED_PANEL_TRACK_H,
} from "../Buffer/column-streams-gen";

const TEXT = new TextDecoder();

const TICK_FONT_PX = 11;
const TICK_FONT = `${TICK_FONT_PX}px monospace`;
const FRAC_SCALE = 0.62;
const FRAC_GAP = 1;
const INK = "#000";
const TRACK_FILL = "#c8c8c8";
const THUMB_FILL = "#fff";
const THUMB_EDGE = "#999";

function readRun(col: number): Float32Array | null {
  const v = columnBytes(col);
  if (!v || v.byteLength === 0) return null;
  const out = new Float32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getFloat32(i * 4, true);
  return out;
}

function readU32Run(col: number): Uint32Array | null {
  const v = columnBytes(col);
  if (!v || v.byteLength === 0) return null;
  const out = new Uint32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getUint32(i * 4, true);
  return out;
}

function readText(col: number): Uint8Array | null {
  const v = columnBytes(col);
  if (!v) return null;
  return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
}

export function SpeedPanelOverlay() {
  const { size, gl } = useThree();
  const canvas = useMemo(() => document.createElement("canvas"), []);
  const texRef = useRef<THREE.CanvasTexture | null>(null);
  const meshRef = useRef<THREE.Mesh | null>(null);
  const sceneRef = useRef(new THREE.Scene());
  const camRef = useRef(new THREE.OrthographicCamera(0, 1, 0, 1, -1, 1));
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
    const cam = camRef.current;
    cam.left = -1; cam.right = 1; cam.top = 1; cam.bottom = -1;
    cam.updateProjectionMatrix();
    meshRef.current = mesh;
    sceneRef.current.add(mesh);
    return () => {
      sceneRef.current.remove(mesh);
      mesh.geometry.dispose();
      (mesh.material as THREE.Material).dispose();
      tex.dispose();
    };
  }, [canvas]);

  useFrame(({ scene, camera }) => {
    const x = readRun(COL_STREAM_SPEED_PANEL_RECT_X);
    const y = readRun(COL_STREAM_SPEED_PANEL_RECT_Y);
    const w = readRun(COL_STREAM_SPEED_PANEL_RECT_W);
    const h = readRun(COL_STREAM_SPEED_PANEL_RECT_H);
    const sel = columnBytes(COL_STREAM_SPEED_PANEL_SELECTED);
    const numText = readText(COL_STREAM_SPEED_PANEL_NUM_TEXT);
    const numLen = readU32Run(COL_STREAM_SPEED_PANEL_NUM_LEN);
    const denText = readText(COL_STREAM_SPEED_PANEL_DEN_TEXT);
    const denLen = readU32Run(COL_STREAM_SPEED_PANEL_DEN_LEN);

    gl.render(scene, camera);

    const tex = texRef.current;
    const mesh = meshRef.current;
    if (!x || !y || !w || !h || !sel || !numText || !numLen || !denText || !denLen || !tex || !mesh) return;

    const trackX = columnF32(COL_STREAM_SPEED_PANEL_TRACK_X);
    const trackY = columnF32(COL_STREAM_SPEED_PANEL_TRACK_Y);
    const trackW = columnF32(COL_STREAM_SPEED_PANEL_TRACK_W);
    const trackH = columnF32(COL_STREAM_SPEED_PANEL_TRACK_H);

    let selectedIndex = -1;
    for (let i = 0; i < sel.byteLength; i++) if (sel.getUint8(i) !== 0) selectedIndex = i;
    const key = `${size.width}x${size.height}@${dpr}|${selectedIndex}|${x.length}|${trackX},${trackY},${trackW},${trackH}`;

    if (key !== lastKey.current) {
      lastKey.current = key;
      canvas.width = Math.max(1, Math.round(size.width * dpr));
      canvas.height = Math.max(1, Math.round(size.height * dpr));
      const c = canvas.getContext("2d");
      if (c) {
        c.setTransform(dpr, 0, 0, dpr, 0, 0);
        c.clearRect(0, 0, size.width, size.height);

        c.fillStyle = TRACK_FILL;
        c.fillRect(trackX, trackY, trackW, trackH);

        let numOff = 0;
        let denOff = 0;
        for (let i = 0; i < x.length; i++) {
          const num = TEXT.decode(numText.subarray(numOff, numOff + numLen[i]!));
          numOff += numLen[i]!;
          const den = TEXT.decode(denText.subarray(denOff, denOff + denLen[i]!));
          denOff += denLen[i]!;
          const on = sel.getUint8(i) !== 0;

          if (on) {
            c.fillStyle = THUMB_FILL;
            c.strokeStyle = THUMB_EDGE;
            c.lineWidth = 1;
            const cx = x[i]! + w[i]! / 2;
            c.beginPath();
            c.arc(cx, trackY + trackH / 2, 6, 0, Math.PI * 2);
            c.fill();
            c.stroke();
          }

          c.fillStyle = INK;
          c.textAlign = "center";
          const cx = x[i]! + w[i]! / 2;
          if (den === "") {
            c.font = `${on ? "bold " : ""}${TICK_FONT}`;
            c.textBaseline = "top";
            c.fillText(num, cx, y[i]!);
          } else {
            const fs = TICK_FONT_PX * FRAC_SCALE;
            c.font = `${on ? "bold " : ""}${fs}px monospace`;
            c.textBaseline = "top";
            c.fillText(num, cx, y[i]!);
            const barY = y[i]! + fs + FRAC_GAP;
            c.fillRect(cx - fs * 0.6, barY, fs * 1.2, 1);
            c.fillText(den, cx, barY + 1 + FRAC_GAP);
          }
        }
      }
      tex.needsUpdate = true;
    }

    gl.autoClear = false;
    gl.clearDepth();
    gl.render(sceneRef.current, camRef.current);
    gl.autoClear = true;
  }, 1);

  return null;
}
