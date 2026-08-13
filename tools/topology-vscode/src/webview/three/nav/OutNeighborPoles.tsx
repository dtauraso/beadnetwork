import React from "react";
import * as THREE from "three";
import { type NodeOutPoles } from "../scene/nodes/node-out-poles";
import { PolarFrame } from "./polar-frame";

// One +x/+y/+z frame per OUTGOING neighbour, with +y along that neighbour's
// pole direction — n outgoing neighbours, n poles. Both the node centre and
// each direction come from the node's own stream frame; PolarFrame turns a
// pole vector into the frame's orientation.
export function OutNeighborPoles({ nodes }: { nodes: NodeOutPoles[] }) {
  return (
    <>
      {nodes.flatMap((n) =>
        n.poles.map((p, i) => (
          <PolarFrame
            key={`${n.row}-${i}`}
            center={new THREE.Vector3(n.cx, n.cy, n.cz)}
            scale={n.radius}
            pole={new THREE.Vector3(p.x, p.y, p.z)}
          />
        )),
      )}
    </>
  );
}
