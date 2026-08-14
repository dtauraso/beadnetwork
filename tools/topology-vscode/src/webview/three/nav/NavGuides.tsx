import React, { useMemo, useState, useEffect, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useOverlayFlags } from "../controls/flags/overlay-flags";
import { getNodeFrame } from "../scene/nodes/node-frame-aggregate";
import { getViewBlocks } from "../scene/view-blocks";
import {
  type NavNode, decodeNavNodes, sceneSphereFromSnapshot,
} from "./buffer-nav";
import { navSignature } from "./nav-signature";
import { PolarFrame } from "./polar-frame";
import { SceneVectors } from "./SceneVectors";
import { NodePoles } from "./NodePoles";

export function NavGuides() {

  const bufFlags = useOverlayFlags();

  const g = bufFlags?.overlays ?? false;
  const showTori = g && !!bufFlags?.tori;
  const showScenePoles = g && !!bufFlags?.scenePoles;
  const showNodePoles = g && !!bufFlags?.nodePoles;
  const showSelPoles = g && !!bufFlags?.selSpherePoles;
  const showHandholds = g && !!bufFlags?.handholds;
  const showSceneVectors = g && !!bufFlags?.sceneVectors;

  const [navTick, setNavTick] = useState(0);
  const bufNavRef = useRef<NavNode[]>([]);
  const bufSigRef = useRef("");

  const sceneSphereRef = useRef<{ center: THREE.Vector3; radius: number }>({ center: new THREE.Vector3(), radius: 100 });
  useFrame(() => {

    if (!showTori && !showScenePoles && !showNodePoles && !showSelPoles && !showHandholds && !showSceneVectors) return;
    const blocks = getViewBlocks();
    const decodedNode = getNodeFrame();
    if (!decodedNode || !blocks) return;
    bufNavRef.current = decodeNavNodes(decodedNode);
    sceneSphereRef.current = sceneSphereFromSnapshot(blocks);
    const sig = navSignature(bufNavRef.current);
    if (sig !== bufSigRef.current) {
      bufSigRef.current = sig;
      setNavTick((t) => t + 1);
    }
  });

  const navNodes = useMemo<NavNode[]>(
    () => bufNavRef.current,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [navTick],
  );


  const latchedSel = navNodes.find((n) => n.latchedSel)?.row ?? null;

  const cs = sceneSphereRef.current;
  const tube = navNodes.length > 0 ? Math.max(0.5, navNodes[0]!.radius * 0.08) : 1;

  const radiusKey = Math.round(cs.radius);
  const tubeKey = Math.round(tube * 10);
  const { geoA, geoB } = useMemo(
    () => ({
      geoA: new THREE.TorusGeometry(radiusKey, tubeKey / 10, 12, 96),
      geoB: new THREE.TorusGeometry(radiusKey, tubeKey / 10, 12, 96),
    }),
    [radiusKey, tubeKey],
  );

  useEffect(() => {
    return () => {
      geoA.dispose();
      geoB.dispose();
    };
  }, [geoA, geoB]);
  const rotB = useMemo(() => new THREE.Euler(Math.PI / 2, 0, 0), []);

  const hhAngles = [0, Math.PI / 2, Math.PI, (3 * Math.PI) / 2];
  const hhRadius = Math.max(radiusKey * 0.04, 3); 
  const handholds = (rotation?: THREE.Euler) => (
    <group rotation={rotation}>
      {hhAngles.map((a) => (
        <mesh key={a} position={[radiusKey * Math.cos(a), radiusKey * Math.sin(a), 0]} userData={{ handhold: true }}>
          <sphereGeometry args={[hhRadius, 16, 16]} />
          <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} transparent opacity={0.9} />
        </mesh>
      ))}
    </group>
  );

  if (navNodes.length < 1) return null;

  const sphereCenters = latchedSel !== null ? navNodes.filter((n) => n.row === latchedSel) : [];

  const pos: [number, number, number] = [cs.center.x, cs.center.y, cs.center.z];
  return (
    <>
      <group position={pos}>
        {showTori && (
          <>
            <mesh geometry={geoA} raycast={() => null}>
              <meshBasicMaterial color="#cc8844" transparent opacity={0.4} depthWrite={false} />
            </mesh>
            <mesh geometry={geoB} rotation={rotB} raycast={() => null}>
              <meshBasicMaterial color="#cc8844" transparent opacity={0.4} depthWrite={false} />
            </mesh>
          </>
        )}
        {}
        {showHandholds && handholds()}
        {showHandholds && handholds(rotB)}
      </group>
      {}
      {showSceneVectors && <SceneVectors center={cs.center} nodes={navNodes} tube={tube * 0.35} />}
      {}
      {showScenePoles && <PolarFrame center={cs.center} scale={radiusKey} />}
      {}
      {}
      {}
      {showNodePoles && <NodePoles nodes={navNodes} />}
      {}
      {showSelPoles && sphereCenters.map((center) => (
        <PolarFrame
          key={`sel-${center.row}`}
          center={center.center}
          scale={center.sphereR ?? center.radius}
          tag={`(${center.label})`}
          octants
        />
      ))}
    </>
  );
}
