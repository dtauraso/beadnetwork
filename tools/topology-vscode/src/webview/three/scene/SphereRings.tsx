




import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "./node-stream-blocks";
import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius, readNodeSelected, readNodeSphereR,
  readNodeVRX, readNodeVRY, readNodeVRZ, readNodeFRX, readNodeFRY, readNodeFRZ,
  readOverlayReachSphere,
} from "../../../schema/buffer-layout";
import { overlayOn } from "../controls/flags/overlay-flags";
import { NODE_SPHERE_RADIUS, NORMAL_DEGENERATE_EPS, SPHERE_RING_MIN_RADIUS, nodeRowColors } from "./buffer-scene-shared";








const SPHERE_RING_EMISSIVE_INTENSITY = 0.25;
const SPHERE_RING_OPACITY = 0.55;
const SPHERE_RING_TUBE_RATIO = 0.08; 
const SPHERE_RING_TUBE_MIN = 0.5;
const SPHERE_RING_RADIAL_SEGMENTS = 12;
const SPHERE_RING_TUBULAR_SEGMENTS = 96;
const _sphereRingDefaultNormal = new THREE.Vector3(0, 0, 1); 

interface OwnerRing {
  row: number; 
  cx: number; cy: number; cz: number;
  R: number; tube: number;
  vrx: number; vry: number; vrz: number;
  frx: number; fry: number; frz: number;
  color: string;
}



function SphereRingBuf({ ring }: { ring: OwnerRing }) {
  const { geo, vrQ, frQ } = useMemo(() => {
    const _geo = new THREE.TorusGeometry(ring.R, ring.tube, SPHERE_RING_RADIAL_SEGMENTS, SPHERE_RING_TUBULAR_SEGMENTS);
    const vrN = new THREE.Vector3(ring.vrx, ring.vry, ring.vrz);
    if (vrN.lengthSq() < NORMAL_DEGENERATE_EPS) vrN.set(0, 0, 1); else vrN.normalize();
    const frN = new THREE.Vector3(ring.frx, ring.fry, ring.frz);
    if (frN.lengthSq() < NORMAL_DEGENERATE_EPS) frN.set(1, 0, 0); else frN.normalize();
    return {
      geo: _geo,
      vrQ: new THREE.Quaternion().setFromUnitVectors(_sphereRingDefaultNormal, vrN),
      frQ: new THREE.Quaternion().setFromUnitVectors(_sphereRingDefaultNormal, frN),
    };
  }, [ring.R, ring.tube, ring.vrx, ring.vry, ring.vrz, ring.frx, ring.fry, ring.frz]);


  useEffect(() => () => { geo.dispose(); }, [geo]);

  return (
    <group position={[ring.cx, ring.cy, ring.cz]}>
      <mesh geometry={geo} quaternion={[vrQ.x, vrQ.y, vrQ.z, vrQ.w]} raycast={() => null} frustumCulled={false}>
        <meshStandardMaterial
          color={ring.color}
          emissive={ring.color}
          emissiveIntensity={SPHERE_RING_EMISSIVE_INTENSITY}
          transparent
          opacity={SPHERE_RING_OPACITY}
          depthWrite={false}
        />
      </mesh>
      <mesh geometry={geo} quaternion={[frQ.x, frQ.y, frQ.z, frQ.w]} raycast={() => null} frustumCulled={false}>
        <meshStandardMaterial
          color={ring.color}
          emissive={ring.color}
          emissiveIntensity={SPHERE_RING_EMISSIVE_INTENSITY}
          transparent
          opacity={SPHERE_RING_OPACITY}
          depthWrite={false}
        />
      </mesh>
    </group>
  );
}

function sameRings(a: OwnerRing[], b: OwnerRing[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const x = a[i]!;
    const y = b[i]!;
    if (
      x.row !== y.row ||
      x.cx !== y.cx || x.cy !== y.cy || x.cz !== y.cz ||
      x.R !== y.R || x.tube !== y.tube ||
      x.vrx !== y.vrx || x.vry !== y.vry || x.vrz !== y.vrz ||
      x.frx !== y.frx || x.fry !== y.fry || x.frz !== y.frz ||
      x.color !== y.color
    ) {
      return false;
    }
  }
  return true;
}

export function SphereRings() {
  const [rings, setRings] = useState<OwnerRing[]>([]);
  const prevRef = useRef<OwnerRing[]>([]);

  useFrame(() => {
    const decoded = getNodeFrame();
    const next: OwnerRing[] = [];



    if (decoded && overlayOn(readOverlayReachSphere)) {
      const { nodeCount, nodeView } = decoded;


      let selectedRow = -1;
      for (let i = 0; i < nodeCount; i++) {
        if (readNodeSelected(nodeView, i)) { selectedRow = i; break; }
      }

      if (selectedRow >= 0) {
        const row = selectedRow;

        const radius = readNodeRadius(nodeView, row) || NODE_SPHERE_RADIUS;
        const R = readNodeSphereR(nodeView, row) || radius;
        if (R >= SPHERE_RING_MIN_RADIUS) {
          const tube = Math.max(SPHERE_RING_TUBE_MIN, radius * SPHERE_RING_TUBE_RATIO);
          const ring: OwnerRing = {
            row,
            cx: readNodeCX(nodeView, row), cy: readNodeCY(nodeView, row), cz: readNodeCZ(nodeView, row),
            R, tube,
            vrx: readNodeVRX(nodeView, row), vry: readNodeVRY(nodeView, row), vrz: readNodeVRZ(nodeView, row),
            frx: readNodeFRX(nodeView, row), fry: readNodeFRY(nodeView, row), frz: readNodeFRZ(nodeView, row),
            color: nodeRowColors(nodeView, row).stroke,
          };
          next.push(ring);
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
        <SphereRingBuf key={ring.row} ring={ring} />
      ))}
    </>
  );
}
