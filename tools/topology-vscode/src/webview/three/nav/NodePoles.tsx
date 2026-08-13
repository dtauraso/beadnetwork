import React from "react";
import { type NavNode } from "./buffer-nav";
import { PolarFrame } from "./polar-frame";

// The world-up pole frame every node carries, +y up, labelled with the node.
//
// It used to be drawn alongside one frame per OUTWARD NEIGHBOUR, aimed along
// that edge's stored path vector. Those paths were a node's cache of where
// its neighbours were, and they went when the centre broadcast that kept them
// current went; a node no longer holds a direction toward anything.
export function NodePoles({ nodes }: { nodes: NavNode[] }) {
  return (
    <>
      {nodes.map((node) => (
        <React.Fragment key={node.row}>
          <PolarFrame
            center={node.center}
            scale={node.radius}
            tag={`(${node.label})`}
          />
        </React.Fragment>
      ))}
    </>
  );
}
