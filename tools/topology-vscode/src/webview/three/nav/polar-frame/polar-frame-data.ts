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

// Every arc is a QUARTER, so each spans pi/2 and its name is the interval it
// covers, in multiples of pi — the same scale the axis labels already use
// ("+Z phi-pi/2"). The two theta meridians cover the same theta twice over, so
// each name carries the phi of the meridian it lies on.
export const THETA_CIRCLES: { sx: number; sy: number; n: number; c: string; label: string }[] = [
  { sx: 1, sy: 1, n: 2, c: "#ff8c00", label: "θ 0→π/2 φ0" },
  { sx: 1, sy: -1, n: 4, c: "#9370db", label: "θ π/2→π φ0" },
  { sx: -1, sy: 1, n: 6, c: "#9acd32", label: "θ 0→π/2 φπ" },
  { sx: -1, sy: -1, n: 8, c: "#cd853f", label: "θ π/2→π φπ" },
];
// The second theta meridian, in the yz plane at phi = pi/2 — perpendicular to
// both THETA_CIRCLES (the phi = 0 meridian, in xy) and PHI_CIRCLES (the
// equator, in xz). Theta sweeps along it exactly as it does along the first
// meridian, so it carries the same quarter arcs and the same +-theta labels.
//
// Quadrants are indexed by (sy, sz). Both +theta arcs take the colour the +theta
// arc near -x already has, so rising theta reads the same on this meridian as on
// the other one; the -theta pair keeps its own colours.
export const THETA_CIRCLES_PHI_HALF: { sy: number; sz: number; n: number; c: string; label: string }[] = [
  { sy: 1, sz: 1, n: 21, c: "#9acd32", label: "θ 0→π/2 φπ/2" },
  { sy: 1, sz: -1, n: 22, c: "#9acd32", label: "θ 0→π/2 φ3π/2" },
  { sy: -1, sz: 1, n: 23, c: "#00ced1", label: "θ π/2→π φπ/2" },
  { sy: -1, sz: -1, n: 24, c: "#9370db", label: "θ π/2→π φ3π/2" },
];

// The equator, so phi runs the whole turn: 0 at +x, pi/2 at +z, pi at -x,
// 3pi/2 at -z.
export const PHI_CIRCLES: { sx: number; sz: number; n: number; c: string; label: string }[] = [
  { sx: 1, sz: 1, n: 11, c: "#00ced1", label: "φ 0→π/2" },
  { sx: 1, sz: -1, n: 12, c: "#9370db", label: "φ 3π/2→2π" },
  { sx: -1, sz: 1, n: 13, c: "#ff69b4", label: "φ π/2→π" },
  { sx: -1, sz: -1, n: 14, c: "#9acd32", label: "φ π→3π/2" },
];
