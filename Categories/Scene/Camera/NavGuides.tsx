import React, { useMemo, useState, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { overlayFlag, overlayFlagSignature } from "../../Overlay/overlay-flags";
import { ownerCounts } from "../owner-counts";
import {
  type NavNode, decodeNavNodes, sceneSphereFromColumns,
} from "./nav-nodes";
import { navSignature } from "./nav-signature";
import { SceneGuides } from "../Guides/SceneGuides";
import { PolarFrame } from "../Poles/PolarFrame";
import { SceneVectors } from "./SceneVectors";
import { NodePoles } from "../../Node/Poles/NodePoles";
import { NodePoleSphere } from "../../Node/Poles/NodePoleSphere";

export function NavGuides() {

  const g = overlayFlag("overlays");
  const showTori = g && overlayFlag("tori");
  const showScenePoles = g && overlayFlag("scenePoles");
  const showNodePoles = g && overlayFlag("nodePoles");
  const showNodePoleSphere = g && overlayFlag("nodePoleSphere");
  const allPoleSpheres = overlayFlag("allPoleSpheres");
  const showHandholds = g && overlayFlag("handholds");
  const showSceneVectors = g && overlayFlag("sceneVectors");

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
    if (!showTori && !showScenePoles && !showNodePoles && !showNodePoleSphere && !showHandholds && !showSceneVectors) return;
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

  if (navNodes.length < 1) return null;

  return (
    <>
      <SceneGuides
        center={cs.center}
        radius={radiusKey}
        tube={tubeKey / 10}
        showTori={showTori}
        showHandholds={showHandholds}
      />
      {showSceneVectors && <SceneVectors center={cs.center} nodes={navNodes} tube={tube * 0.35} />}
      {showScenePoles && <PolarFrame center={cs.center} scale={radiusKey} />}
      {showNodePoles && <NodePoles nodes={navNodes} />}
      {showNodePoleSphere && <NodePoleSphere nodes={navNodes} all={allPoleSpheres} />}
    </>
  );
}
