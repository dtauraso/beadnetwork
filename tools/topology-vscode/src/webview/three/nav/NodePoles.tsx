import React from "react";
import { type NavNode } from "./buffer-nav";
import { PolarFrame } from "./polar-frame";

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
