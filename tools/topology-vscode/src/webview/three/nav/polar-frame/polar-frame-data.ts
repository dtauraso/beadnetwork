export const HANDHOLD_TERM_TAG = "handholdTerm";

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

export const THETA_CIRCLES: { sx: number; sy: number; n: number; c: string }[] = [
  { sx: 1, sy: 1, n: 2, c: "#ff8c00" },
  { sx: 1, sy: -1, n: 4, c: "#9370db" },
  { sx: -1, sy: 1, n: 6, c: "#9acd32" },
  { sx: -1, sy: -1, n: 8, c: "#cd853f" },
];
// The second theta meridian, in the yz plane at phi = pi/2 — perpendicular to
// both THETA_CIRCLES (the phi = 0 meridian, in xy) and PHI_CIRCLES (the
// equator, in xz). Theta sweeps along it exactly as it does along the first
// meridian, so it carries the same quarter arcs and the same +-theta labels.
//
// Quadrants are indexed by (sy, sz) and coloured with the +x octants, the way
// the xy circle is coloured with the -z ones.
export const THETA_CIRCLES_PHI_HALF: { sy: number; sz: number; n: number; c: string }[] = [
  { sy: 1, sz: 1, n: 21, c: "#ffffff" },
  { sy: 1, sz: -1, n: 22, c: "#ff8c00" },
  { sy: -1, sz: 1, n: 23, c: "#00ced1" },
  { sy: -1, sz: -1, n: 24, c: "#9370db" },
];

export const PHI_CIRCLES: { sx: number; sz: number; n: number; c: string }[] = [
  { sx: 1, sz: 1, n: 11, c: "#00ced1" },
  { sx: 1, sz: -1, n: 12, c: "#9370db" },
  { sx: -1, sz: 1, n: 13, c: "#ff69b4" },
  { sx: -1, sz: -1, n: 14, c: "#9acd32" },
];
