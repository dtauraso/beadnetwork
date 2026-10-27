import { useEffect, useMemo } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { LineSegments2 } from "three/examples/jsm/lines/LineSegments2.js";
import { LineSegmentsGeometry } from "three/examples/jsm/lines/LineSegmentsGeometry.js";
import { LineMaterial } from "three/examples/jsm/lines/LineMaterial.js";
import { ownerCounts } from "../../Scene/owner-counts";
import { ndcToPixel } from "../../Scene/Drag/ndc";
import { nodeBytes } from "../node-leaves";
import type { NodeValueName } from "../node-values-gen";
import { setArcLabels, type ArcLabel } from "../Labels/label-canvas";

const PHI_COLOR = "#FF9F1C";
const KIND_COLORS = [PHI_COLOR, "#3DDC84"];
const KIND_NAMES = ["φ", "θ"];
const LINE_WIDTH_PX = 3;
const WEDGE_OPACITY = 0.18;

const ARC = ["arcX0", "arcY0", "arcZ0", "arcX1", "arcY1", "arcZ1"] as const;
const WEDGE = ["wedgeX0", "wedgeY0", "wedgeZ0", "wedgeX1", "wedgeY1", "wedgeZ1", "wedgeX2", "wedgeY2", "wedgeZ2"] as const;

const kindRGB = KIND_COLORS.map((c) => new THREE.Color(c));

function rgb(kind: number): THREE.Color {
  return kindRGB[kind] ?? new THREE.Color(PHI_COLOR);
}

function f32(col: DataView | undefined, i: number): number {
  return col && col.byteLength >= i * 4 + 4 ? col.getFloat32(i * 4, true) : 0;
}

function u8(col: DataView | undefined, i: number): number {
  return col && col.byteLength > i ? col.getUint8(i) : 0;
}

function columns(row: number, names: readonly NodeValueName[]): (DataView | undefined)[] {
  return names.map((n) => nodeBytes(row, n));
}

function sameArray(a: Float32Array, b: Float32Array): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

function formatTicks(v: number): string {
  const r = Math.round(v * 10) / 10;
  return Number.isInteger(r) ? String(r) : r.toFixed(1);
}

const _label = new THREE.Vector3();

export function AngleArcs() {
  const { camera, gl, size } = useThree();

  const { lines, lineGeom, lineMat, wedgeGeom, held } = useMemo(() => {
    const lineGeom = new LineSegmentsGeometry();
    const lineMat = new LineMaterial({ vertexColors: true, linewidth: LINE_WIDTH_PX, worldUnits: false });
    const lines = new LineSegments2(lineGeom, lineMat);
    lines.frustumCulled = false;
    lines.raycast = () => {};
    const wedgeGeom = new THREE.BufferGeometry();
    return {
      lines,
      lineGeom,
      lineMat,
      wedgeGeom,
      held: {
        positions: new Float32Array(0),
        colors: new Float32Array(0),
        wedgePos: new Float32Array(0),
        wedgeCol: new Float32Array(0),
      },
    };
  }, []);

  useEffect(() => () => {
    lineGeom.dispose();
    lineMat.dispose();
    wedgeGeom.dispose();
    setArcLabels([]);
  }, [lineGeom, lineMat, wedgeGeom]);

  useFrame(() => {
    lineMat.resolution.set(size.width, size.height);
    const { nodes } = ownerCounts();

    const pos: number[] = [];
    const col: number[] = [];
    const wPos: number[] = [];
    const wCol: number[] = [];
    const labels: ArcLabel[] = [];
    const el = gl.domElement;
    const px = { width: Math.max(1, el.clientWidth), height: Math.max(1, el.clientHeight) };

    for (let row = 0; row < nodes; row++) {
      const a = columns(row, ARC);
      const ak = nodeBytes(row, "arcKind");
      const n = a[0] ? a[0].byteLength / 4 : 0;
      for (let i = 0; i < n; i++) {
        for (let c = 0; c < 6; c++) pos.push(f32(a[c], i));
        const k = rgb(u8(ak, i));
        col.push(k.r, k.g, k.b, k.r, k.g, k.b);
      }

      const w = columns(row, WEDGE);
      const wk = nodeBytes(row, "wedgeKind");
      const wn = w[0] ? w[0].byteLength / 4 : 0;
      for (let i = 0; i < wn; i++) {
        for (let c = 0; c < 9; c++) wPos.push(f32(w[c], i));
        const k = rgb(u8(wk, i));
        wCol.push(k.r, k.g, k.b, k.r, k.g, k.b, k.r, k.g, k.b);
      }

      const lx = nodeBytes(row, "arcLabelX");
      const ly = nodeBytes(row, "arcLabelY");
      const lz = nodeBytes(row, "arcLabelZ");
      const lk = nodeBytes(row, "arcLabelKind");
      const lt = nodeBytes(row, "arcLabelTicks");
      const ln = lx ? lx.byteLength / 4 : 0;
      for (let i = 0; i < ln; i++) {
        _label.set(f32(lx, i), f32(ly, i), f32(lz, i)).project(camera);
        if (_label.z > 1) continue;
        const p = ndcToPixel(_label.x, _label.y, px);
        const kind = u8(lk, i);
        labels.push({
          text: `${KIND_NAMES[kind] ?? "?"} ${formatTicks(f32(lt, i))}`,
          color: KIND_COLORS[kind] ?? PHI_COLOR,
          px: p.px,
          py: p.py,
        });
      }
    }

    const nextPos = Float32Array.from(pos);
    const nextCol = Float32Array.from(col);
    if (!sameArray(nextPos, held.positions) || !sameArray(nextCol, held.colors)) {
      held.positions = nextPos;
      held.colors = nextCol;
      lineGeom.setPositions(nextPos);
      lineGeom.setColors(nextCol);
      lines.visible = nextPos.length > 0;
    }

    const nextWPos = Float32Array.from(wPos);
    const nextWCol = Float32Array.from(wCol);
    if (!sameArray(nextWPos, held.wedgePos) || !sameArray(nextWCol, held.wedgeCol)) {
      held.wedgePos = nextWPos;
      held.wedgeCol = nextWCol;
      wedgeGeom.setAttribute("position", new THREE.BufferAttribute(nextWPos, 3));
      wedgeGeom.setAttribute("color", new THREE.BufferAttribute(nextWCol, 3));
    }

    setArcLabels(labels);
  });

  return (
    <>
      <primitive object={lines} />
      <mesh geometry={wedgeGeom} frustumCulled={false} raycast={() => null}>
        <meshBasicMaterial vertexColors transparent opacity={WEDGE_OPACITY} side={THREE.DoubleSide} depthWrite={false} />
      </mesh>
    </>
  );
}
