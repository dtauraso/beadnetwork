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

export const PHI_CIRCLES: { sx: number; sy: number; n: number; c: string; label: string }[] = [
  { sx: 1, sy: 1, n: 2, c: "#ff8c00", label: "φ π/2" },
  { sx: 1, sy: -1, n: 4, c: "#9370db", label: "φ π" },
  { sx: -1, sy: 1, n: 6, c: "#9acd32", label: "φ π/2 θ π" },
  { sx: -1, sy: -1, n: 8, c: "#cd853f", label: "φ π θ π" },
];
export const PHI_CIRCLES_THETA_HALF: { sy: number; sz: number; n: number; c: string; label: string }[] = [
  { sy: 1, sz: 1, n: 21, c: "#9acd32", label: "φ π/2 θ π/2" },
  { sy: 1, sz: -1, n: 22, c: "#9acd32", label: "φ π/2 θ 3π/2" },
  { sy: -1, sz: 1, n: 23, c: "#00ced1", label: "φ π θ π/2" },
  { sy: -1, sz: -1, n: 24, c: "#9370db", label: "φ π θ 3π/2" },
];

export const THETA_CIRCLES: { sx: number; sz: number; n: number; c: string; label: string }[] = [
  { sx: 1, sz: 1, n: 11, c: "#00ced1", label: "φ π/2 θ π/2" },
  { sx: 1, sz: -1, n: 12, c: "#9370db", label: "φ π/2 θ 2π" },
  { sx: -1, sz: 1, n: 13, c: "#ff69b4", label: "φ π/2 θ π" },
  { sx: -1, sz: -1, n: 14, c: "#9acd32", label: "φ π/2 θ 3π/2" },
];
