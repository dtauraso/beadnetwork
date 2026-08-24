import React from "react";
import { type NavNode } from "../nav-nodes";
import { PolarFrame } from "../../Scene/Poles/PolarFrame";
import { overlayFlag } from "../../Scene/View/Flags/overlay-flags";

export function NodePoleSphere({ nodes }: { nodes: NavNode[] }) {
  if (!overlayFlag("nodePoleSphere")) return null;
  const all = overlayFlag("allPoleSpheres");

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
