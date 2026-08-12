// polar-frame-data.ts — the octant/circle legend tables PolarFrame draws from, split out of
// polar-frame.tsx. Pure static data: no hook, no three.js object, no computation — just the
// sign triples, numbers, and colors the octant-mode rendering indexes into.

// HANDHOLD_TERM_TAG — userData key stamped (value `true`) on the octant angle handhold
// meshes and the pole-crossing radius handholds to mark them as pickable handholds. This is
// a PRESENCE marker only — no numeric term crosses the TS→Go bridge (a "polar rule-builder"
// consumer of a per-handhold term was proposed but never built; Go's handhold-down branch
// only needs to know a handhold was grabbed, not which one).
export const HANDHOLD_TERM_TAG = "handholdTerm";

// The 8 octants of the polar sphere — a sign triple (±x,±y,±z), a distinct color, and a
// compact label. When octants={true} the θ/φ angle arcs are reflected (group scale) into
// each octant and colored from here, so every octant gets its own angle-arc pair.
export const OCTANTS: { s: [number, number, number]; color: string; tag: string }[] = [
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
export const THETA_CIRCLES: { sx: number; sy: number; n: number; c: string }[] = [
  { sx: 1, sy: 1, n: 2, c: "#ff8c00" },
  { sx: 1, sy: -1, n: 4, c: "#9370db" },
  { sx: -1, sy: 1, n: 6, c: "#9acd32" },
  { sx: -1, sy: -1, n: 8, c: "#cd853f" },
];
export const PHI_CIRCLES: { sx: number; sz: number; n: number; c: string }[] = [
  { sx: 1, sz: 1, n: 11, c: "#00ced1" },
  { sx: 1, sz: -1, n: 12, c: "#9370db" },
  { sx: -1, sz: 1, n: 13, c: "#ff69b4" },
  { sx: -1, sz: -1, n: 14, c: "#9acd32" },
];
