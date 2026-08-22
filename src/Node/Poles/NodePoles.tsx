import React from "react";
import { type NavNode } from "../../Camera/nav-nodes";
import { PolarFrame } from "../../Scene/Poles/PolarFrame";

export function NodePoles({ nodes }: { nodes: NavNode[] }) {
  return (
    <>
      {nodes.map((node) => (
        <React.Fragment key={node.row}>
          <PolarFrame
            center={node.center}
            scale={node.radius}
            tag={`(${node.label})`}
            pole={node.pole}
          />
        </React.Fragment>
      ))}
    </>
  );
}
