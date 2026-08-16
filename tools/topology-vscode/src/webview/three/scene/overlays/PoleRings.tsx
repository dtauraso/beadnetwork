import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "../nodes/node-frame-aggregate";
import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeSelected, readNodePoleRingR,
  readNodeVRX, readNodeVRY, readNodeVRZ,
  readNodeFRX, readNodeFRY, readNodeFRZ,
  readNodeSRX, readNodeSRY, readNodeSRZ,
  readOverlayNodePoles,
} from "../../../../schema/buffer-layout/buffer-layout";
import { overlayOn } from "../../controls/flags/overlay-flags";
import { nodeRowColors } from "../buffer-scene-shared";
import {
  SHADING_PARAM_POLE_RING_TUBE_RATIO,
  SHADING_PARAM_POLE_RING_OPACITY,
  SHADING_PARAM_POLE_RING_EMISSIVE_INTENSITY,
} from "../../../../schema/buffer-layout/shading-params";

const RING_RADIAL_SEGMENTS = 12;
const RING_TUBULAR_SEGMENTS = 96;
const DEGENERATE_EPS = 1e-12;
const RING_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);

interface PoleRing {
  row: number;
  cx: number; cy: number; cz: number;
  R: number;
  normals: [number, number, number][];
  color: string;
}

function normalsEqual(a: [number, number, number][], b: [number, number, number][]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const x = a[i]!;
    const y = b[i]!;
    if (x[0] !== y[0] || x[1] !== y[1] || x[2] !== y[2]) return false;
  }
  return true;
}

function PoleRingMesh({ ring }: { ring: PoleRing }) {
  const { geo, quats } = useMemo(() => {
    const tube = ring.R * SHADING_PARAM_POLE_RING_TUBE_RATIO;
    const _geo = new THREE.TorusGeometry(ring.R, tube, RING_RADIAL_SEGMENTS, RING_TUBULAR_SEGMENTS);
    const _quats = ring.normals.map(([nx, ny, nz]) => {
      const n = new THREE.Vector3(nx, ny, nz);
      if (n.lengthSq() < DEGENERATE_EPS) n.copy(RING_DEFAULT_NORMAL);
      else n.normalize();
      return new THREE.Quaternion().setFromUnitVectors(RING_DEFAULT_NORMAL, n);
    });
    return { geo: _geo, quats: _quats };
  }, [ring.R, ring.normals]);

  useEffect(() => () => { geo.dispose(); }, [geo]);

  return (
    <group position={[ring.cx, ring.cy, ring.cz]}>
      {quats.map((q, i) => (
        <mesh
          key={i}
          geometry={geo}
          quaternion={[q.x, q.y, q.z, q.w]}
          raycast={() => null}
          frustumCulled={false}
        >
          <meshStandardMaterial
            color={ring.color}
            emissive={ring.color}
            emissiveIntensity={SHADING_PARAM_POLE_RING_EMISSIVE_INTENSITY}
            transparent
            opacity={SHADING_PARAM_POLE_RING_OPACITY}
            depthWrite={false}
          />
        </mesh>
      ))}
    </group>
  );
}

function sameRings(a: PoleRing[], b: PoleRing[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const x = a[i]!;
    const y = b[i]!;
    if (
      x.row !== y.row ||
      x.cx !== y.cx || x.cy !== y.cy || x.cz !== y.cz ||
      x.R !== y.R || x.color !== y.color ||
      !normalsEqual(x.normals, y.normals)
    ) {
      return false;
    }
  }
  return true;
}

export function PoleRings() {
  const [rings, setRings] = useState<PoleRing[]>([]);
  const prevRef = useRef<PoleRing[]>([]);

  useFrame(() => {
    const decoded = getNodeFrame();
    const next: PoleRing[] = [];

    if (decoded && overlayOn(readOverlayNodePoles)) {
      const { nodeCount, nodeView } = decoded;

      let selectedRow = -1;
      for (let i = 0; i < nodeCount; i++) {
        if (readNodeSelected(nodeView, i)) { selectedRow = i; break; }
      }

      if (selectedRow >= 0) {
        const row = selectedRow;
        const R = readNodePoleRingR(nodeView, row);
        if (R > 0) {
          next.push({
            row,
            cx: readNodeCX(nodeView, row),
            cy: readNodeCY(nodeView, row),
            cz: readNodeCZ(nodeView, row),
            R,
            normals: [
              [readNodeVRX(nodeView, row), readNodeVRY(nodeView, row), readNodeVRZ(nodeView, row)],
              [readNodeFRX(nodeView, row), readNodeFRY(nodeView, row), readNodeFRZ(nodeView, row)],
              [readNodeSRX(nodeView, row), readNodeSRY(nodeView, row), readNodeSRZ(nodeView, row)],
            ],
            color: nodeRowColors(nodeView, row).stroke,
          });
        }
      }
    }

    if (!sameRings(prevRef.current, next)) {
      prevRef.current = next;
      setRings(next);
    }
  });

  return (
    <>
      {rings.map((ring) => (
        <PoleRingMesh key={ring.row} ring={ring} />
      ))}
    </>
  );
}
