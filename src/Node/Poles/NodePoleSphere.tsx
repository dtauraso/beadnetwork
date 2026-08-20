import React from "react";
import { type NavNode } from "../../Camera/buffer-nav";
import { PolarFrame } from "../../Scene/Poles/PolarFrame";

export function NodePoleSphere({ nodes, all }: { nodes: NavNode[]; all: boolean }) {
  return (
    <>
      {nodes.filter((n) => all || n.selected || n.latchedSel).map((node) => (
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
