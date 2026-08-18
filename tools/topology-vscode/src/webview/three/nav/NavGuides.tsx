import React, { useMemo, useState, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useOverlayFlags } from "../controls/flags/overlay-flags";
import { getNodeSections } from "../scene/nodes/node-sections";
import { getViewBlocks } from "../scene/view-blocks";
import {
  type NavNode, decodeNavNodes, sceneSphereFromColumns,
} from "./buffer-nav";
import { navSignature } from "./nav-signature";
import { SceneGuides } from "../../../../Scene/Guides/SceneGuides";
import { PolarFrame } from "../../../../Scene/Poles/PolarFrame";
import { SceneVectors } from "../../../../Scene/Vectors/SceneVectors";
import { NodePoles } from "../../../../Node/Poles/NodePoles";
import { NodePoleSphere } from "../../../../Node/Poles/NodePoleSphere";

export function NavGuides() {

  const bufFlags = useOverlayFlags();

  const g = bufFlags?.overlays ?? false;
  const showTori = g && !!bufFlags?.tori;
  const showScenePoles = g && !!bufFlags?.scenePoles;
  const showNodePoles = g && !!bufFlags?.nodePoles;
  const showNodePoleSphere = g && !!bufFlags?.nodePoleSphere;
  const allPoleSpheres = !!bufFlags?.allPoleSpheres;
  const showHandholds = g && !!bufFlags?.handholds;
  const showSceneVectors = g && !!bufFlags?.sceneVectors;

  const [navTick, setNavTick] = useState(0);
  const bufNavRef = useRef<NavNode[]>([]);
  const bufSigRef = useRef("");

  const sceneSphereRef = useRef<{ center: THREE.Vector3; radius: number }>({ center: new THREE.Vector3(), radius: 100 });
  useFrame(() => {

    if (!showTori && !showScenePoles && !showNodePoles && !showNodePoleSphere && !showHandholds && !showSceneVectors) return;
    const blocks = getViewBlocks();
    const decodedNode = getNodeSections();
    if (!decodedNode || !blocks) return;
    bufNavRef.current = decodeNavNodes();
    sceneSphereRef.current = sceneSphereFromColumns();
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
