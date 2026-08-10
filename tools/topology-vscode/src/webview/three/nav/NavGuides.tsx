// NavGuides.tsx — decorative 3D navigation overlays: the polar-sphere tori, the scene/node/
// selected-sphere pole frames (drawn by PolarFrame), and the grab handholds.
// Purely decorative: raycast disabled, depthWrite false, transparent.

import React, { useMemo, useState, useEffect, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useOverlayFlags } from "../controls/flags/overlay-flags";
import { getNodeFrame } from "../scene/node-stream-blocks";
import { getViewBlocks } from "../scene/view-blocks";
import {
  type NavNode, decodeNavNodes, sceneSphereFromSnapshot,
} from "./buffer-nav";
import { navSignature } from "./nav-signature";
import { PolarFrame } from "./polar-frame";

// NavGuides — decorative 3D navigation overlays (the polar-sphere tori, pole
// frames, and θ/φ arcs). Rendered directly as the combined export; there is no
// pass-through wrapper.
export function NavGuides() {
  // Overlay flags are Go-owned and streamed into the buffer's Overlay columns. useOverlayFlags
  // subscribes to snapshot arrivals so a flip re-renders even when the node-position
  // navSignature is unchanged. null until the first snapshot lands (nothing to draw yet).
  const bufFlags = useOverlayFlags();

  // "Overlays" master gate (Go-owned): when false, ALL polar guides are suppressed (the
  // toolbar also hides their individual buttons). It does NOT touch each guide's own
  // Go-owned visibility, so reactivating restores every guide to its prior on/off state.
  const g = bufFlags?.overlays ?? false;
  const showTori = g && !!bufFlags?.tori;
  const showScenePoles = g && !!bufFlags?.scenePoles;
  const showNodePoles = g && !!bufFlags?.nodePoles;
  const showSelPoles = g && !!bufFlags?.selSpherePoles;
  const showHandholds = g && !!bufFlags?.handholds;

  // ── Buffer-driven nav sampling ───────────────────────────────────────────────
  // The overlay geometry derives from the binary buffer (Go-owned node centers/radii/sphereR
  // + Go-owned selection column). Sample the latest snapshot each frame and bump a render tick
  // only when the coarse signature changes, so tori/frames rebuild on real position/selection
  // changes (a drag) — not every frame.
  const [navTick, setNavTick] = useState(0);
  const bufNavRef = useRef<NavNode[]>([]);
  const bufSigRef = useRef("");
  // Scene sphere: Go-owned, established once at load and never moved (see
  // sceneSphereFromSnapshot) — sampled alongside navNodes but not part of navSignature
  // since it is constant after the first snapshot.
  const sceneSphereRef = useRef<{ center: THREE.Vector3; radius: number }>({ center: new THREE.Vector3(), radius: 100 });
  useFrame(() => {
    // Visibility gate FIRST: if none of the guides this component renders is on, skip the
    // per-node decode/allocate work entirely (decodeNavNodes/sceneSphereFromSnapshot/
    // navSignature all allocate per node, per frame). Mirrors the exact flag set the JSX
    // below gates on, read early instead of only at render time.
    if (!showTori && !showScenePoles && !showNodePoles && !showSelPoles && !showHandholds) return;
    const blocks = getViewBlocks();
    const decodedNode = getNodeFrame();
    if (!decodedNode || !blocks) return;
    bufNavRef.current = decodeNavNodes(decodedNode);
    sceneSphereRef.current = sceneSphereFromSnapshot(blocks);
    const sig = navSignature(bufNavRef.current);
    if (sig !== bufSigRef.current) {
      bufSigRef.current = sig;
      setNavTick((t) => t + 1);
    }
  });

  // Node records that drive every guide below. Memoized so downstream guide computations
  // recompute only when the node data actually changes (navTick bumps on a real change).
  const navNodes = useMemo<NavNode[]>(
    () => bufNavRef.current,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [navTick],
  );

  // Latched selection: Go-owned LatchedSel column (see Buffer/layout.go; set by the
  // affected node's own nodeMover). Selection only DECIDES which sphere the sel-highlight frames; it
  // does not have to stay selected to keep the frame shown. So DEselecting the node
  // (clicking empty space) leaves the latched sphere framed — only selecting a different
  // node replaces it. The sel toggle still gates visibility. This is read-only reflection
  // of Go's own latch state — NavGuides authors nothing.
  const latchedSel = navNodes.find((n) => n.latchedSel)?.row ?? null;

  // WORLD-FIXED scene sphere (Go-owned, established once at load — see
  // sceneSphereFromSnapshot), so it zooms WITH the diagram. Tube thickness matches the node
  // spheres' tori (scene-content SphereRing: max(0.5, nodeRadius·0.08)).
  const cs = sceneSphereRef.current;
  const tube = navNodes.length > 0 ? Math.max(0.5, navNodes[0]!.radius * 0.08) : 1;
  // Build geometry ONLY when the sphere actually changes (rounded radius/tube), not on
  // every render — rebuilding each frame under node-geometry churn made the tori flicker
  // and effectively disappear.
  const radiusKey = Math.round(cs.radius);
  const tubeKey = Math.round(tube * 10);
  const { geoA, geoB } = useMemo(
    () => ({
      geoA: new THREE.TorusGeometry(radiusKey, tubeKey / 10, 12, 96),
      geoB: new THREE.TorusGeometry(radiusKey, tubeKey / 10, 12, 96),
    }),
    [radiusKey, tubeKey],
  );
  // Dispose the outgoing GPU geometries when the memo rebuilds (radius/tube change) or
  // on unmount. React runs this cleanup for the PREVIOUS geoA/geoB before creating the
  // next pair, so the still-mounted current pair is never double-disposed. NavGuides
  // re-renders on every node-geometry stream event (incl. drags); without this the
  // replaced TorusGeometry buffers leak.
  useEffect(() => {
    return () => {
      geoA.dispose();
      geoB.dispose();
    };
  }, [geoA, geoB]);
  const rotB = useMemo(() => new THREE.Euler(Math.PI / 2, 0, 0), []);

  // Handholds: 4 grab points per torus, 90° apart. Grabbing one starts a CONSTRAINED
  // rotation — the first two cursor points lock the rotation disk (see
  // interaction-handlers.ts "handhold-rotating"). These are the only PICKABLE part of
  // the nav overlay (the tori stay raycast-disabled); each carries userData.handhold.
  // Placed in the torus's own local frame: geoA lies in XY (handholds at z=0), geoB is
  // the same ring under rotB, so its handhold group shares that rotation.
  const hhAngles = [0, Math.PI / 2, Math.PI, (3 * Math.PI) / 2];
  const hhRadius = Math.max(radiusKey * 0.04, 3); // grabbable, scales with the sphere
  const handholds = (rotation?: THREE.Euler) => (
    <group rotation={rotation}>
      {hhAngles.map((a) => (
        <mesh key={a} position={[radiusKey * Math.cos(a), radiusKey * Math.sin(a), 0]} userData={{ handhold: true }}>
          <sphereGeometry args={[hhRadius, 16, 16]} />
          <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} transparent opacity={0.9} />
        </mesh>
      ))}
    </group>
  );

  if (navNodes.length < 1) return null;

  // Selected-sphere poles (separate, additional feature — gated by selSpherePolesVisible,
  // independent of the per-node poles below). The LATCHED selection decides which node's
  // sphere to frame (persists through deselect), and we draw THAT node's own sphere pole
  // frame at full SPHERE scale (its Go-streamed sphereR). Every node has a sphere, so this
  // works for leaf nodes (3, 5) too — no parent remapping. Never selected ⇒ no frame.
  const sphereCenters = latchedSel !== null ? navNodes.filter((n) => n.row === latchedSel) : [];

  // WORLD-FIXED tori: the pole is the diagram's own top axis (world Y), so the horizontal torus
  // (geoB, normal world Y) is the diagram's equator — the polar frame is anchored to the
  // diagram, not the camera.
  const pos: [number, number, number] = [cs.center.x, cs.center.y, cs.center.z];
  return (
    <>
      <group position={pos}>
        {showTori && (
          <>
            <mesh geometry={geoA} raycast={() => null}>
              <meshBasicMaterial color="#cc8844" transparent opacity={0.4} depthWrite={false} />
            </mesh>
            <mesh geometry={geoB} rotation={rotB} raycast={() => null}>
              <meshBasicMaterial color="#cc8844" transparent opacity={0.4} depthWrite={false} />
            </mesh>
          </>
        )}
        {/* Grab handholds (4 per torus, 90° apart) — the pickable part of the overlay. Gated by both overlaysVisible (master) and handholdsVisible (per-overlay). */}
        {showHandholds && handholds()}
        {showHandholds && handholds(rotB)}
      </group>
      {/* Scene pole frame at the content-sphere center. */}
      {showScenePoles && <PolarFrame center={cs.center} scale={radiusKey} />}
      {/* Per-node pole frames — one PolarFrame per node, gated behind nodePolesVisible. */}
      {/* NODE 1 ONLY carries its streamed pole today (row 0 — ROW ID = NODE ID - 1). The
          pole is not special to node 1: every node streams its own (Buffer/layout.go
          PoleTheta/PolePhi), and dropping this row check is all it takes to honour the rest.
          They are ALL held at world +y, so every node's frame reads the same way. */}
      {showNodePoles && navNodes.map((node) => (
        <PolarFrame
          key={node.row}
          center={node.center}
          scale={node.radius}
          tag={`(${node.label})`}
          /* No pole override: EVERY node's frame is held at world +y, node 1 included.
             Row 0 alone used to be rotated onto its own streamed pole, from when node 1 was
             the only vector-connected node — which left one frame in the scene tilted
             differently from all the others and reading as a property of that node rather
             than of the feature. Go still streams every node's pole (Buffer/layout.go's
             PoleTheta/PolePhi); passing it here for all of them is the other way to make
             them agree, and is what to do when the frames should follow their nodes. */
        />
      ))}
      {/* Selected-sphere poles (additional feature) — the center(s) of the sphere(s) the
          SELECTED node sits on, drawn at SPHERE scale. Independent of the per-node poles. */}
      {showSelPoles && sphereCenters.map((center) => (
        <PolarFrame
          key={`sel-${center.row}`}
          center={center.center}
          scale={center.sphereR ?? center.radius}
          tag={`(${center.label})`}
          octants
        />
      ))}
    </>
  );
}
