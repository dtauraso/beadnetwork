import React from "react";
import * as THREE from "three";
import { PolarFrame } from "./PolarFrame";
import { overlayFlag } from "../../Overlay/overlay-flags";

export function ScenePoles({ center, scale }: {
  center: THREE.Vector3;
  scale: number;
}) {
  if (!overlayFlag("scenePoles")) return null;

  return <PolarFrame center={center} scale={scale} />;
}
