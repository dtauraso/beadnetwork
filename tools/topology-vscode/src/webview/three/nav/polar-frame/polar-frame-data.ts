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

// Every arc is a QUARTER, and each is named by the point it REACHES: both
// angles, in multiples of pi, on the same scale the axis labels use ("+Z
// phi-pi/2"). One number per angle — the far end fixes the near one, since
// every arc is pi/2 wide.
//
// An angle of ZERO is left out rather than written. "theta pi/2" is the arc
// reaching theta = pi/2 at phi = 0; there is no "phi 0" to read.
export const PHI_CIRCLES: { sx: number; sy: number; n: number; c: string; label: string }[] = [
  { sx: 1, sy: 1, n: 2, c: "#ff8c00", label: "φ π/2" },
  { sx: 1, sy: -1, n: 4, c: "#9370db", label: "φ π" },
  { sx: -1, sy: 1, n: 6, c: "#9acd32", label: "φ π/2 θ π" },
  { sx: -1, sy: -1, n: 8, c: "#cd853f", label: "φ π θ π" },
];
// The second theta meridian, in the yz plane at phi = pi/2 — perpendicular to
// both PHI_CIRCLES (the phi = 0 meridian, in xy) and THETA_CIRCLES (the
// equator, in xz). Theta sweeps along it exactly as it does along the first
// meridian, so it carries the same quarter arcs and the same +-theta labels.
//
// Quadrants are indexed by (sy, sz). Both +theta arcs take the colour the +theta
// arc near -x already has, so rising theta reads the same on this meridian as on
// the other one; the -theta pair keeps its own colours.
export const PHI_CIRCLES_THETA_HALF: { sy: number; sz: number; n: number; c: string; label: string }[] = [
  { sy: 1, sz: 1, n: 21, c: "#9acd32", label: "φ π/2 θ π/2" },
  { sy: 1, sz: -1, n: 22, c: "#9acd32", label: "φ π/2 θ 3π/2" },
  { sy: -1, sz: 1, n: 23, c: "#00ced1", label: "φ π θ π/2" },
  { sy: -1, sz: -1, n: 24, c: "#9370db", label: "φ π θ 3π/2" },
];

// The equator: theta is pi/2 the whole way round, and phi runs the full turn —
// pi/2 at +z, pi at -x, 3pi/2 at -z, 2pi back at +x. Theta is named here like
// anywhere else, because pi/2 is not zero.
export const THETA_CIRCLES: { sx: number; sz: number; n: number; c: string; label: string }[] = [
  { sx: 1, sz: 1, n: 11, c: "#00ced1", label: "φ π/2 θ π/2" },
  { sx: 1, sz: -1, n: 12, c: "#9370db", label: "φ π/2 θ 2π" },
  { sx: -1, sz: 1, n: 13, c: "#ff69b4", label: "φ π/2 θ π" },
  { sx: -1, sz: -1, n: 14, c: "#9acd32", label: "φ π/2 θ 3π/2" },
];
