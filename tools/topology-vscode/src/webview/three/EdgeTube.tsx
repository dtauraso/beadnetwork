// EdgeTube.tsx — the cascade-link overlay, plus the EdgeTubes buffer-poll wrapper. The
// edge's own drawn LINE (the tube + arrowhead + pick/selection halo that used to trace every
// edge's segment, matching the JSON path's SingleEdgeTube) is REMOVED COMPLETELY, not
// replaced by a smaller stand-in: the source node's own chain of placeholder beads is the
// edge's visual now (docs/beads-are-the-edge.md, MODEL.md's wire-lifecycle section), so a
// separately-drawn line — visible or an invisible pick halo — duplicated that depiction.
// Consequence, recorded rather than silently patched over: an edge can no longer be
// raycast-picked or shown as Go-selected, because the only mesh that ever carried
// userData[BUFFER_EDGE_TAG] was this file's now-deleted halo. `raw-input.ts`'s `edgeOnly`
// pick and `scene-content.tsx`'s `pickBufferEdge` are consequently unreachable (nothing ever
// tags a hit), and `EdgeAccessor.selected`/`readEdgeSelected`/the Edge block's `Selected`
// column are dead on the TS side — left in place rather than torn out (a buffer-schema
// change is a bigger, separate decision; `readEdgeSelected` is listed in
// `tools/check-no-dead-buffer-column.sh`'s ALLOWED_DEAD with this same note). The Edge
// block's SEGMENT (SX..EZ) stays very much alive: `edge-stream-blocks.ts`'s `segment()` and
// `edgeCount` are still read for stream-capacity growth (`buffer-scene.tsx`) and by the
// `.probe` debug decoder (`buffer-log.ts`) — nothing about removing the RENDERED tube
// touches that. The cascade-link overlay (`LayoutLinkOverlay`, `EdgeTubes`'s `showCascade`
// branch) is a DIFFERENT feature — its own edge between two NODE CENTERS, never riding a
// bead edge — and is untouched. There is no PortInstances any more
// (docs/channels-not-ports.md — a port is never drawn).
//
// TIMING CONTRACT (why this file is imperative, not setState-driven):
// NodeInstances updates node meshes IMPERATIVELY inside its useFrame (setMatrixAt +
// instanceMatrix.needsUpdate), so a moved node lands on the SAME frame it is decoded. If the
// cascade-link overlay's segment coordinates flowed through React state (setSegs ->
// re-render -> useMemo rebuild), it would land ONE FRAME LATER than the nodes it connects —
// a render-side lag, not a data bug. So per-frame COORDINATES are pushed to each overlay
// slot via an imperative handle (EdgeHandle.update), updated in the same useFrame that reads
// the node/edge streams — never through state.
//
// What DOES stay in useState: the mounted cascade-link SLOT COUNT and the dim flag. Those
// change on link add/remove and the overlay toggle, never per drag-frame, so a one-frame
// commit latency on them is imperceptible. Holding them is buffer reflection (count/flags Go
// owns), not domain authority — no segment geometry is cached in state (check-no-webview-state).

import React, {
  useRef, useState, useEffect, forwardRef, useImperativeHandle,
} from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getViewBlocks } from "./view-blocks";
import { getNodeFrame, getLayoutLinks } from "./node-stream-blocks";
import {
  SHADING_PARAM_LAYOUT_LINK_COLOR,
  SHADING_PARAM_LAYOUT_LINK_EMISSIVE,
  SHADING_PARAM_LAYOUT_LINK_EMISSIVE_INTENSITY,
} from "../../schema/shading-params";
import {
  readLayoutLinkSrcNodeRow, readLayoutLinkDstNodeRow,
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius,
  readOverlayOverlaysVis, readOverlayCascadeLinks,
} from "../../schema/buffer-layout";
import { DIRECTION_ZERO_EPS } from "./buffer-scene-shared";

// Arrowhead cone dims for the cascade-link overlay (the pre-branch sizes) — the only
// arrowhead left in this file once the per-edge tube/arrow/halo is gone.
const DL_ARROWHEAD_LENGTH = 7;
const DL_ARROWHEAD_RADIUS = 3.5;

const LAYOUT_LINK_EMISSIVE_COLOR = new THREE.Color(SHADING_PARAM_LAYOUT_LINK_EMISSIVE);

interface EdgeSeg { sx: number; sy: number; sz: number; ex: number; ey: number; ez: number; }

// Imperative per-slot handle: the parent pushes this slot's current segment every frame,
// bypassing React state so the cascade-link overlay lands the same frame as the nodes it
// connects (see the timing contract at the top of this file).
interface EdgeHandle { update(seg: EdgeSeg): void }

function sameSeg(a: EdgeSeg, b: EdgeSeg): boolean {
  return a.sx === b.sx && a.sy === b.sy && a.sz === b.sz
    && a.ex === b.ex && a.ey === b.ey && a.ez === b.ez;
}

/**
 * Builds an arrow descriptor: a cone whose apex sits at `apex`, pointing in `dir`
 * (normalized, toward the apex). ConeGeometry apex is at +Y; we rotate +Y onto `dir`.
 * center places the cone so its apex lands at `apex`.
 */
function buildArrow(apex: THREE.Vector3, dir: THREE.Vector3, height: number): {
  center: THREE.Vector3; q: THREE.Quaternion;
} {
  const q = new THREE.Quaternion().setFromUnitVectors(new THREE.Vector3(0, 1, 0), dir);
  const center = apex.clone().addScaledVector(dir, -height / 2);
  return { center, q };
}

// One cascade-link pair's cyan bidirectional overlay: thin tube (radius 1.0) + an
// outward-pointing arrowhead at each end. Mirrors the pre-removal DoubleEdgeOverlay. This
// is its OWN edge between the two NODES' CENTERS (Node block CX/CY/CZ, re-streamed on
// every move) — it does NOT reference or ride along any bead edge, so it can never be
// coupled to (or dimmed/tinted by) the bead edge's own selection/opacity state. One
// uniform color/opacity for every cascade link, always.
//
// Same timing contract as EdgeTube: the segment is pushed imperatively (update), not a prop,
// so a link overlay tracks its dragged endpoints on the same frame as the nodes it connects.
const LayoutLinkOverlay = forwardRef<EdgeHandle, object>(
  function LayoutLinkOverlay(_props, ref) {
    const lineMeshRef = useRef<THREE.Mesh>(null);
    const arrowStartRef = useRef<THREE.Mesh>(null);
    const arrowEndRef = useRef<THREE.Mesh>(null);
    const lastSeg = useRef<EdgeSeg | null>(null);
    const geoRef = useRef<THREE.TubeGeometry | null>(null);

    useImperativeHandle(ref, () => ({
      update(seg: EdgeSeg) {
        if (lastSeg.current && sameSeg(lastSeg.current, seg)) return;
        lastSeg.current = seg;

        const start = new THREE.Vector3(seg.sx, seg.sy, seg.sz);
        const end = new THREE.Vector3(seg.ex, seg.ey, seg.ez);
        const curve = new THREE.LineCurve3(start, end);
        // Pre-branch cascade-link sizes: thin tube (radius 1.0) + an arrowhead at each end.
        const lineGeo = new THREE.TubeGeometry(curve, 1, 1.0, 6, false);
        if (geoRef.current) geoRef.current.dispose();
        geoRef.current = lineGeo;
        if (lineMeshRef.current) lineMeshRef.current.geometry = lineGeo;

        const dir = end.clone().sub(start);
        const as = arrowStartRef.current;
        const ae = arrowEndRef.current;
        if (dir.length() >= DIRECTION_ZERO_EPS) {
          const dirNorm = dir.clone().normalize();
          if (as) {
            const { center, q } = buildArrow(start, dirNorm.clone().negate(), DL_ARROWHEAD_LENGTH);
            as.position.set(center.x, center.y, center.z);
            as.quaternion.set(q.x, q.y, q.z, q.w);
            as.visible = true;
          }
          if (ae) {
            const { center, q } = buildArrow(end, dirNorm, DL_ARROWHEAD_LENGTH);
            ae.position.set(center.x, center.y, center.z);
            ae.quaternion.set(q.x, q.y, q.z, q.w);
            ae.visible = true;
          }
        } else {
          if (as) as.visible = false;
          if (ae) ae.visible = false;
        }
      },
    }), []);

    useEffect(() => () => { if (geoRef.current) geoRef.current.dispose(); }, []);

    const coneMesh = (r: React.Ref<THREE.Mesh>) => (
      <mesh ref={r} raycast={() => null} frustumCulled={false} visible={false}>
        <coneGeometry args={[DL_ARROWHEAD_RADIUS, DL_ARROWHEAD_LENGTH, 16]} />
        <meshStandardMaterial
          color={SHADING_PARAM_LAYOUT_LINK_COLOR}
          emissive={LAYOUT_LINK_EMISSIVE_COLOR}
          emissiveIntensity={SHADING_PARAM_LAYOUT_LINK_EMISSIVE_INTENSITY}
        />
      </mesh>
    );

    return (
      <>
        <mesh ref={lineMeshRef} raycast={() => null} frustumCulled={false}>
          <meshStandardMaterial
            color={SHADING_PARAM_LAYOUT_LINK_COLOR}
            emissive={LAYOUT_LINK_EMISSIVE_COLOR}
            emissiveIntensity={SHADING_PARAM_LAYOUT_LINK_EMISSIVE_INTENSITY}
          />
        </mesh>
        {coneMesh(arrowStartRef)}
        {coneMesh(arrowEndRef)}
      </>
    );
  },
);

// No `capacity` (edge-slot count) parameter any more: it sized the now-deleted per-edge tube
// pool. `buffer-scene.tsx`'s own edge-stream-capacity growth bookkeeping is unrelated to
// rendering (it still reads `getEdgeStreamAccessor().edgeCount` for segment/label decode
// capacity) and needs no matching edit.
export function EdgeTubes({ layoutLinkCapacity }: { layoutLinkCapacity: number }) {
  const [showCascade, setShowCascade] = useState(false);
  // Mounted layout-link slot count — low-frequency (a link is added/removed, or the
  // overlay toggles) — not per-frame.
  const [linkCount, setLinkCount] = useState(0);

  // Imperative handles to every mounted cascade-link slot — this is the per-frame coordinate
  // channel that replaces the old setLinkSegs state (see the timing contract at top of file).
  const linkHandles = useRef<(EdgeHandle | null)[]>([]);

  useFrame(() => {
    const blocks = getViewBlocks();
    const decodedNode = getNodeFrame();
    if (!decodedNode || !blocks) return;
    // Layout-link overlay pairs: aggregated from the per-node dedicated streams' own
    // outbound layout-links (see getLayoutLinks' doc comment,
    // memory/feedback_no_single_writer_bridge.md).
    const { layoutLinkCount, layoutLinkView } = getLayoutLinks();
    const { overlayView } = blocks;
    // LayoutLink's SrcNodeRow/DstNodeRow resolve against the NODE frame's Node block — both
    // frames are built from the same stable seed-order row tables in the same emit call, so they
    // share the same stable node-row order (see frame_tags.go's BufBlockTagNode comment).
    const { nodeView } = decodedNode;

    // Cascade-link overlay: Go-streamed pairs (LayoutLink block, sourced from each node's
    // OWN cascade-edges.json — see node_mover.go's cascadeEdges doc comment). This is its
    // OWN edge between the two NODES' CENTERS — it never references or rides along a bead
    // edge, so it can never be coupled to (or dimmed/tinted by) the bead edge's own
    // selection/opacity state.
    // Both overlay flags (0/1 columns) must be set. Coerce each side explicitly with `> 0`.
    const cascade = readOverlayOverlaysVis(overlayView) > 0 && readOverlayCascadeLinks(overlayView) > 0;
    if (cascade !== showCascade) setShowCascade(cascade);

    // Clamp with the layout-link's OWN capacity — cascade links are independent of the
    // (now-deleted) per-edge tube pool.
    const linkN = Math.min(layoutLinkCount, layoutLinkCapacity);
    if (linkN !== linkCount) setLinkCount(linkN);

    for (let i = 0; i < linkN; i++) {
      const srcRow = readLayoutLinkSrcNodeRow(layoutLinkView, i);
      const dstRow = readLayoutLinkDstNodeRow(layoutLinkView, i);
      // Center-to-center would drive the tube THROUGH each node sphere; pull each
      // endpoint back to its node's SURFACE by offsetting inward along the link
      // direction by that node's radius, so cascade links terminate at the bodies
      // (not their centers) and don't pile up/darken inside the nodes.
      const scx = readNodeCX(nodeView, srcRow), scy = readNodeCY(nodeView, srcRow), scz = readNodeCZ(nodeView, srcRow);
      const dcx = readNodeCX(nodeView, dstRow), dcy = readNodeCY(nodeView, dstRow), dcz = readNodeCZ(nodeView, dstRow);
      const rSrc = readNodeRadius(nodeView, srcRow), rDst = readNodeRadius(nodeView, dstRow);
      let ux = dcx - scx, uy = dcy - scy, uz = dcz - scz;
      const len = Math.hypot(ux, uy, uz);
      if (len > 1e-6) { ux /= len; uy /= len; uz /= len; } else { ux = uy = uz = 0; }
      const seg: EdgeSeg = {
        sx: scx + ux * rSrc, sy: scy + uy * rSrc, sz: scz + uz * rSrc,
        ex: dcx - ux * rDst, ey: dcy - uy * rDst, ez: dcz - uz * rDst,
      };
      linkHandles.current[i]?.update(seg);
    }
  });

  return (
    <>
      {showCascade && Array.from({ length: linkCount }, (_, i) => (
        <LayoutLinkOverlay
          key={`layout-link-row-${i}`}
          ref={(h) => { linkHandles.current[i] = h; }}
        />
      ))}
    </>
  );
}
