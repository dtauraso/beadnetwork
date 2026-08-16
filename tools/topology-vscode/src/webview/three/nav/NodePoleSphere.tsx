import React from "react";
import { type NavNode } from "./buffer-nav";
import { PolarFrame } from "./polar-frame";

export function NodePoleSphere({ nodes }: { nodes: NavNode[] }) {
  return (
    <>
      {nodes.map((node) => (
        <React.Fragment key={node.row}>
          <PolarFrame
            center={node.center}
            scale={node.poleRingR}
            tag={`(${node.label})`}
            pole={node.pole}
            octants
          />
        </React.Fragment>
      ))}
    </>
  );
}
