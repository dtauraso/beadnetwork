import React, { useMemo, useState, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { overlayFlag, overlayFlagSignature } from "../View/Flags/overlay-flags";
import { ownerCounts } from "../owner-counts";
import {
  type NavNode, decodeNavNodes, sceneSphereFromColumns,
} from "../../Node/nav-nodes";
import { navSignature } from "../../Node/nav-nodes";
import { Tori } from "../Tori/Tori";
import { Handholds } from "../Handholds/Handholds";
import { ScenePoles } from "../Poles/ScenePoles";
import { SceneVectors } from "../SceneVectors/SceneVectors";
import { NodePoles } from "../../Node/Poles/NodePoles";
import { NodePoleSphere } from "../../Node/Poles/NodePoleSphere";

const GUIDE_FLAGS = [
  "tori", "handholds", "sceneVectors", "scenePoles", "nodePoles", "nodePoleSphere",
] as const;

export function SceneGuides() {
  const guidesOn = overlayFlag("overlays");
  const anyGuideVisible = guidesOn && GUIDE_FLAGS.some((f) => overlayFlag(f));

  const [navTick, setNavTick] = useState(0);
  const navRef = useRef<NavNode[]>([]);
  const navSigRef = useRef("");
  const flagSigRef = useRef(overlayFlagSignature());

  const sceneSphereRef = useRef<{ center: THREE.Vector3; radius: number }>({ center: new THREE.Vector3(), radius: 100 });
  useFrame(() => {
    const flagSig = overlayFlagSignature();
    if (flagSig !== flagSigRef.current) {
      flagSigRef.current = flagSig;
      setNavTick((t) => t + 1);
    }
    if (!anyGuideVisible) return;
    if (ownerCounts().nodes <= 0) return;
    navRef.current = decodeNavNodes();
    sceneSphereRef.current = sceneSphereFromColumns();
    const sig = navSignature(navRef.current);
    if (sig !== navSigRef.current) {
      navSigRef.current = sig;
      setNavTick((t) => t + 1);
    }
  });

  const navNodes = useMemo<NavNode[]>(
    () => navRef.current,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [navTick],
  );

  const cs = sceneSphereRef.current;
  const tube = navNodes.length > 0 ? navNodes[0]!.radius : 1;

  const radiusKey = Math.round(cs.radius);
  const tubeKey = Math.round(tube * 10);

  if (!guidesOn) return null;
  if (navNodes.length < 1) return null;

  return (
    <>
      <Tori center={cs.center} radius={radiusKey} tube={tubeKey / 10} />
      <Handholds center={cs.center} radius={radiusKey} />
      <SceneVectors center={cs.center} nodes={navNodes} tube={tube * 0.35} />
      <ScenePoles center={cs.center} scale={radiusKey} />
      <NodePoles nodes={navNodes} />
      <NodePoleSphere nodes={navNodes} />
    </>
  );
}
