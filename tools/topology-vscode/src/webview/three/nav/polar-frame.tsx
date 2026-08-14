import React, { useMemo } from "react";
import * as THREE from "three";
import { computePolarFrameGeometry } from "./polar-frame/polar-frame-geometry";
import { PolarAxisArrows } from "./polar-frame/PolarAxisArrows";
import { PolarArcs } from "./polar-frame/PolarArcs";
import { PolarAxisLabels } from "./polar-frame/PolarAxisLabels";

const WORLD_UP = new THREE.Vector3(0, 1, 0);

export function PolarFrame({ center, scale, tag, pole }: {
  center: THREE.Vector3; scale: number; tag?: string; pole?: THREE.Vector3;
}) {
  const { poleLen, poleRadius, coneH, coneBaseR, arcR, arcTube, arcMid } =
    computePolarFrameGeometry(scale);
  const sfx = tag ? ` ${tag}` : "";

  const quat = useMemo(() => {
    const q = new THREE.Quaternion();
    if (pole) q.setFromUnitVectors(WORLD_UP, pole.clone().normalize());
    return q;
  }, [pole]);
  return (
    <group position={[center.x, center.y, center.z]} quaternion={quat}>
      <PolarAxisArrows poleLen={poleLen} poleRadius={poleRadius} coneH={coneH} coneBaseR={coneBaseR} />
      <PolarArcs arcR={arcR} arcTube={arcTube} />
      <PolarAxisLabels poleLen={poleLen} coneH={coneH} arcMid={arcMid} sfx={sfx} />
    </group>
  );
}
