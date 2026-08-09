// polar-frame.tsx — PolarFrame: the camera-independent pole-frame markers for ONE center
// (used for both the scene sphere and each node's own sphere). Purely decorative except for
// its octant-mode handholds: raycast disabled elsewhere, depthWrite false, transparent.

import React, { useMemo } from "react";
import * as THREE from "three";
import { AxisLabel } from "./axis-label";

// WORLD_UP — the axis PolarFrame is drawn poled at. A frame given its own pole rotates as
// a whole from this axis onto it (PolarFrame's quat), so +y stays the only pole any of the
// meshes below is written against.
const WORLD_UP = new THREE.Vector3(0, 1, 0);

// HANDHOLD_TERM_TAG — userData key stamped (value `true`) on the octant angle handhold
// meshes and the pole-crossing radius handholds to mark them as pickable handholds. This is
// a PRESENCE marker only — no numeric term crosses the TS→Go bridge (a "polar rule-builder"
// consumer of a per-handhold term was proposed but never built; Go's handhold-down branch
// only needs to know a handhold was grabbed, not which one).
export const HANDHOLD_TERM_TAG = "handholdTerm";

// The 8 octants of the polar sphere — a sign triple (±x,±y,±z), a distinct color, and a
// compact label. When octants={true} the θ/φ angle arcs are reflected (group scale) into
// each octant and colored from here, so every octant gets its own angle-arc pair.
const OCTANTS: { s: [number, number, number]; color: string; tag: string }[] = [
  { s: [1, 1, 1], color: "#ffffff", tag: "+x+y+z" },
  { s: [1, 1, -1], color: "#ff8c00", tag: "+x+y−z" },
  { s: [1, -1, 1], color: "#00ced1", tag: "+x−y+z" },
  { s: [1, -1, -1], color: "#9370db", tag: "+x−y−z" },
  { s: [-1, 1, 1], color: "#ff69b4", tag: "−x+y+z" },
  { s: [-1, 1, -1], color: "#9acd32", tag: "−x+y−z" },
  { s: [-1, -1, 1], color: "#00bfff", tag: "−x−y+z" },
  { s: [-1, -1, -1], color: "#cd853f", tag: "−x−y−z" },
];

// ── ARC NUMBER ↔ COLOR LEGEND ───────────────────────────────────────────────
// Each quarter-arc carries a unique number (θ arcs 1..8, φ arcs 9..16) drawn near
// it, colored by its octant (OCTANTS[i].color). θ# = i+1, φ# = i+9.
//
// Per-octant (number → octant → color):
//    #1 / #9   +x+y+z   white        #ffffff
//    #2 / #10  +x+y−z   orange       #ff8c00
//    #3 / #11  +x−y+z   teal         #00ced1
//    #4 / #12  +x−y−z   purple       #9370db
//    #5 / #13  −x+y+z   pink         #ff69b4
//    #6 / #14  −x+y−z   yellow-green  #9acd32
//    #7 / #15  −x−y+z   sky-blue     #00bfff
//    #8 / #16  −x−y−z   peru/tan     #cd853f
//
// Grouped by shared-position REGION (the two offset circles you see together —
// a→color1, b→color2 — so you can note just the numbers):
//   θ regions (X-Y plane):        φ regions (X-Z plane):
//     +x+y :  1 white  / 2 orange    +x+z :  9 white      / 11 teal
//     +x−y :  3 teal   / 4 purple    +x−z : 10 orange     / 12 purple
//     −x+y :  5 pink   / 6 yel-grn   −x+z : 13 pink       / 15 sky-blue
//     −x−y :  7 sky-blu/ 8 peru      −x−z : 14 yel-grn    / 16 peru
// ────────────────────────────────────────────────────────────────────────────

// User-chosen single circle per region (1 per θ/φ). Each: sign pair, its number, color.
const THETA_CIRCLES: { sx: number; sy: number; n: number; c: string }[] = [
  { sx: 1, sy: 1, n: 2, c: "#ff8c00" },
  { sx: 1, sy: -1, n: 4, c: "#9370db" },
  { sx: -1, sy: 1, n: 6, c: "#9acd32" },
  { sx: -1, sy: -1, n: 8, c: "#cd853f" },
];
const PHI_CIRCLES: { sx: number; sz: number; n: number; c: string }[] = [
  { sx: 1, sz: 1, n: 11, c: "#00ced1" },
  { sx: 1, sz: -1, n: 12, c: "#9370db" },
  { sx: -1, sz: 1, n: 13, c: "#ff69b4" },
  { sx: -1, sz: -1, n: 14, c: "#9acd32" },
];

// PolarFrame — the camera-independent pole-frame markers for ONE center: the three
// axis sticks (+y pole green, +x φ0 red, +z φ90 blue) plus the θ (magenta) and φ
// (yellow) angle arcs, all anchored at `center` with the pole = world +y. `scale`
// sizes the frame (≈ the radius it should reach). `tag` suffixes the axis labels so
// the scene frame and a node's frame are distinguishable. Decorative (raycast off),
// not affected by the scene-tori toggle. Same drawing for every center, so node 2's
// frame matches the scene's exactly.
export function PolarFrame({ center, scale, tag, octants, pole }: {
  center: THREE.Vector3; scale: number; tag?: string; octants?: boolean; pole?: THREE.Vector3;
}) {
  const radiusKey = Math.max(Math.round(scale), 1);
  const poleLen = radiusKey * 1.3;
  const poleRadius = Math.max(radiusKey * 0.01, 1);
  const coneH = radiusKey * 0.12;
  const coneBaseR = radiusKey * 0.05;
  const arcR = poleLen * 0.68;
  const arcTube = Math.max(radiusKey * 0.012, 1.2);
  const arcMid = arcR * 1.12 * Math.SQRT1_2;
  const hhR = Math.max(radiusKey * 0.04, 3);   // handhold sphere radius (matches the tori handholds)
  const arcHH = arcR * Math.SQRT1_2;           // a quarter-arc's midpoint radius (45° in its plane)
  const sfx = tag ? ` ${tag}` : "";
  // The frame is DRAWN poled at +y throughout (every mesh below is placed in that frame);
  // an alternate pole is applied as one rotation of the whole group, so there is exactly
  // one place a pole can differ and no per-mesh axis to get wrong. Omitted pole = world +y,
  // i.e. identity — the frames that don't carry one are byte-identical to before.
  const quat = useMemo(() => {
    const q = new THREE.Quaternion();
    if (pole) q.setFromUnitVectors(WORLD_UP, pole.clone().normalize());
    return q;
  }, [pole]);
  return (
    <group position={[center.x, center.y, center.z]} quaternion={quat}>
      {/* +Y pole (green). */}
      <mesh position={[0, poleLen / 2, 0]} raycast={() => null}>
        <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
        <meshBasicMaterial color="#22dd55" depthWrite={false} />
      </mesh>
      <mesh position={[0, poleLen + coneH / 2, 0]} raycast={() => null}>
        <coneGeometry args={[coneBaseR, coneH, 12]} />
        <meshBasicMaterial color="#22dd55" depthWrite={false} />
      </mesh>
      {/* +X equatorial reference (φ=0, red). */}
      <mesh position={[poleLen / 2, 0, 0]} rotation={[0, 0, -Math.PI / 2]} raycast={() => null}>
        <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
        <meshBasicMaterial color="#dd3333" depthWrite={false} />
      </mesh>
      <mesh position={[poleLen + coneH / 2, 0, 0]} rotation={[0, 0, -Math.PI / 2]} raycast={() => null}>
        <coneGeometry args={[coneBaseR, coneH, 12]} />
        <meshBasicMaterial color="#dd3333" depthWrite={false} />
      </mesh>
      {/* +Z equatorial reference (φ=90°, blue). */}
      <mesh position={[0, 0, poleLen / 2]} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
        <meshBasicMaterial color="#3366dd" depthWrite={false} />
      </mesh>
      <mesh position={[0, 0, poleLen + coneH / 2]} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <coneGeometry args={[coneBaseR, coneH, 12]} />
        <meshBasicMaterial color="#3366dd" depthWrite={false} />
      </mesh>
      {/* Negative spokes (octant mode): the other halves of each axis (−Y/−X/−Z), so the
          full ±X ±Y ±Z cross frames all 8 octants. Same colors as the positive halves. */}
      {octants && (<>
        <mesh position={[0, -poleLen / 2, 0]} raycast={() => null}>
          <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
          <meshBasicMaterial color="#22dd55" depthWrite={false} />
        </mesh>
        <mesh position={[0, -(poleLen + coneH / 2), 0]} rotation={[Math.PI, 0, 0]} raycast={() => null}>
          <coneGeometry args={[coneBaseR, coneH, 12]} />
          <meshBasicMaterial color="#22dd55" depthWrite={false} />
        </mesh>
        <mesh position={[-poleLen / 2, 0, 0]} rotation={[0, 0, Math.PI / 2]} raycast={() => null}>
          <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
          <meshBasicMaterial color="#dd3333" depthWrite={false} />
        </mesh>
        <mesh position={[-(poleLen + coneH / 2), 0, 0]} rotation={[0, 0, Math.PI / 2]} raycast={() => null}>
          <coneGeometry args={[coneBaseR, coneH, 12]} />
          <meshBasicMaterial color="#dd3333" depthWrite={false} />
        </mesh>
        <mesh position={[0, 0, -poleLen / 2]} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
          <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
          <meshBasicMaterial color="#3366dd" depthWrite={false} />
        </mesh>
        <mesh position={[0, 0, -(poleLen + coneH / 2)]} rotation={[-Math.PI / 2, 0, 0]} raycast={() => null}>
          <coneGeometry args={[coneBaseR, coneH, 12]} />
          <meshBasicMaterial color="#3366dd" depthWrite={false} />
        </mesh>
      </>)}
      {!octants && (<>
      {/* θ angle arc (magenta): quarter-sweep from +Y pole to +X, X-Y meridian plane. */}
      <mesh raycast={() => null}>
        <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
        <meshBasicMaterial color="#dd33cc" depthWrite={false} />
      </mesh>
      {/* φ angle arc (yellow): quarter-sweep in equatorial X-Z plane, +X (φ0)→+Z (φ90). */}
      <mesh rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
        <meshBasicMaterial color="#dddd22" depthWrite={false} />
      </mesh>
      </>)}
      {octants && THETA_CIRCLES.map((t) => (
        <group key={`tc-${t.n}`} scale={[t.sx, t.sy, 1]}>
          <mesh raycast={() => null}>
            <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
            <meshBasicMaterial color={t.c} depthWrite={false} />
          </mesh>
        </group>
      ))}
      {octants && PHI_CIRCLES.map((p) => (
        <group key={`pc-${p.n}`} scale={[p.sx, 1, p.sz]}>
          <mesh rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
            <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
            <meshBasicMaterial color={p.c} depthWrite={false} />
          </mesh>
        </group>
      ))}
      {/* Labels — billboard sprites, always face the camera. */}
      <AxisLabel text={`+Y pole${sfx}`} color="#22dd55" position={[0, poleLen + coneH * 2, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+X φ0${sfx}`} color="#dd3333" position={[poleLen + coneH * 2, 0, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+Z φπ/2${sfx}`} color="#3366dd" position={[0, 0, poleLen + coneH * 2]} size={poleLen * 0.12} />
      {octants && (<>
        <AxisLabel text={`−Y${sfx}`} color="#22dd55" position={[0, -(poleLen + coneH * 2), 0]} size={poleLen * 0.12} />
        <AxisLabel text={`−X φπ${sfx}`} color="#dd3333" position={[-(poleLen + coneH * 2), 0, 0]} size={poleLen * 0.12} />
        <AxisLabel text={`−Z φ3π/2${sfx}`} color="#3366dd" position={[0, 0, -(poleLen + coneH * 2)]} size={poleLen * 0.12} />
      </>)}
      {!octants && (<>
      <AxisLabel text="θ" color="#dd33cc" position={[arcMid, arcMid, 0]} size={poleLen * 0.14} />
      <AxisLabel text="φ" color="#dddd22" position={[arcMid, 0, arcMid]} size={poleLen * 0.14} />
      </>)}
      {octants && THETA_CIRCLES.map((t) => (
        <AxisLabel key={`tl-${t.n}`} text={`${t.sy > 0 ? "+" : "−"}θ`} color={t.c} position={[t.sx * arcMid, t.sy * arcMid, 0]} size={poleLen * 0.11} />
      ))}
      {octants && PHI_CIRCLES.map((p) => (
        <AxisLabel key={`pl-${p.n}`} text={`${p.sz > 0 ? "+" : "−"}φ`} color={p.c} position={[p.sx * arcMid, 0, p.sz * arcMid]} size={poleLen * 0.11} />
      ))}
      {octants && (<>
        {/* Radius (r) handholds: the six pole-crossing grab-spheres (±arcR on each axis).
            All pickable and stamped as handholds (presence marker only). */}
        {([[arcR, 0, 0], [-arcR, 0, 0], [0, arcR, 0], [0, -arcR, 0], [0, 0, arcR], [0, 0, -arcR]] as [number, number, number][]).map((p, i) => (
          <mesh key={`hhp-${i}`} position={p} userData={{ [HANDHOLD_TERM_TAG]: true }}>
            <sphereGeometry args={[hhR, 12, 12]} />
            <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
          </mesh>
        ))}
        {/* θ/φ angle handholds: pickable, stamped as handholds (presence marker only). */}
        {THETA_CIRCLES.map((t) => (
          <mesh
            key={`th-${t.n}`}
            position={[t.sx * arcHH, t.sy * arcHH, 0]}
            userData={{ [HANDHOLD_TERM_TAG]: true }}
          >
            <sphereGeometry args={[hhR, 12, 12]} />
            <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
          </mesh>
        ))}
        {PHI_CIRCLES.map((p) => (
          <mesh
            key={`ph-${p.n}`}
            position={[p.sx * arcHH, 0, p.sz * arcHH]}
            userData={{ [HANDHOLD_TERM_TAG]: true }}
          >
            <sphereGeometry args={[hhR, 12, 12]} />
            <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
          </mesh>
        ))}
      </>)}
    </group>
  );
}
