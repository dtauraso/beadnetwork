import React from "react";
import { type NavNode } from "../nav-nodes";
import { PolarFrame } from "../../Scene/Poles/PolarFrame";
import { overlayFlag } from "../../Overlay/overlay-flags";

export function NodePoles({ nodes }: { nodes: NavNode[] }) {
  if (!overlayFlag("nodePoles")) return null;

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
