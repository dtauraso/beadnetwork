import { useEffect, useMemo } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { LineSegments2 } from "three/examples/jsm/lines/LineSegments2.js";
import { LineSegmentsGeometry } from "three/examples/jsm/lines/LineSegmentsGeometry.js";
import { LineMaterial } from "three/examples/jsm/lines/LineMaterial.js";
import { ownerCounts } from "../../Scene/owner-counts";
import { nodeBytes } from "../node-leaves";
import type { NodeValueName } from "../node-values-gen";

const PHI_COLOR = "#FF9F1C";
const KIND_COLORS = [PHI_COLOR, "#3DDC84", "#F5E663", "#4DA3FF"];
const LINE_WIDTH_PX = 2;

const ARC = ["arcX0", "arcY0", "arcZ0", "arcX1", "arcY1", "arcZ1"] as const;

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

export function AngleArcs() {
  const { size } = useThree();

  const { lines, lineGeom, lineMat, held } = useMemo(() => {
    const lineGeom = new LineSegmentsGeometry();
    const lineMat = new LineMaterial({ vertexColors: true, linewidth: LINE_WIDTH_PX, worldUnits: false });
    const lines = new LineSegments2(lineGeom, lineMat);
    lines.frustumCulled = false;
    lines.raycast = () => {};
    return {
      lines,
      lineGeom,
      lineMat,
      held: { positions: new Float32Array(0), colors: new Float32Array(0) },
    };
  }, []);

  useEffect(() => () => {
    lineGeom.dispose();
    lineMat.dispose();
  }, [lineGeom, lineMat]);

  useFrame(() => {
    lineMat.resolution.set(size.width, size.height);
    const { nodes } = ownerCounts();

    const pos: number[] = [];
    const col: number[] = [];

    for (let row = 0; row < nodes; row++) {
      const a = columns(row, ARC);
      const ak = nodeBytes(row, "arcKind");
      const n = a[0] ? a[0].byteLength / 4 : 0;
      for (let i = 0; i < n; i++) {
        for (let c = 0; c < 6; c++) pos.push(f32(a[c], i));
        const k = rgb(u8(ak, i));
        col.push(k.r, k.g, k.b, k.r, k.g, k.b);
      }
    }

    const nextPos = Float32Array.from(pos);
    const nextCol = Float32Array.from(col);
    if (!sameArray(nextPos, held.positions) || !sameArray(nextCol, held.colors)) {
      held.positions = nextPos;
      held.colors = nextCol;
      lines.visible = nextPos.length > 0;
      if (lines.visible) {
        lineGeom.setPositions(nextPos);
        lineGeom.setColors(nextCol);
      }
    }
  });

  return <primitive object={lines} />;
}
