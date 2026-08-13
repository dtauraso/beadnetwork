import React from "react";
import * as THREE from "three";
import { type NavNode } from "./buffer-nav";
import { type NodeOutPoles } from "../scene/nodes/node-out-poles";
import { PolarFrame } from "./polar-frame";

// Every pole frame a node carries, in one place:
//
//   - ONE world-up frame, +y up, labelled with the node
//   - ONE per OUTWARD NEIGHBOUR, +y along that edge's stored path vector
//
// The up frame is not a fallback for having no neighbours — it is drawn for
// every node, alongside however many neighbour frames it has.
export function NodePoles({ nodes, outPoles }: {
  nodes: NavNode[];
  outPoles: NodeOutPoles[];
}) {
  const polesByRow = new Map(outPoles.map((p) => [p.row, p.poles]));

  return (
    <>
      {nodes.map((node) => (
        <React.Fragment key={node.row}>
          <PolarFrame
            center={node.center}
            scale={node.radius}
            tag={`(${node.label})`}
          />
          {(polesByRow.get(node.row) ?? []).map((p, i) => (
            <PolarFrame
              key={i}
              center={node.center}
              scale={node.radius}
              pole={new THREE.Vector3(p.x, p.y, p.z)}
            />
          ))}
        </React.Fragment>
      ))}
    </>
  );
}
