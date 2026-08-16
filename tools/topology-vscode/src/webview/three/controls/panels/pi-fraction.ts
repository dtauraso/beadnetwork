const MAX_DENOMINATOR = 64;
const RATIONAL_EPS = 1e-6;

export interface PiFraction {
  p: number;
  q: number;
}

export function piFractionOf(multiple: number): PiFraction | null {
  for (let q = 1; q <= MAX_DENOMINATOR; q++) {
    const p = Math.round(multiple * q);
    if (Math.abs(multiple * q - p) < RATIONAL_EPS) return { p, q };
  }
  return null;
}

export function formatPi(multiple: number): string {
  const f = piFractionOf(multiple);
  if (!f) return `${Math.round(multiple * 1000) / 1000}π`;
  if (f.p === 0) return "0";
  const num = f.p === 1 ? "π" : f.p === -1 ? "−π" : `${f.p}π`;
  return f.q === 1 ? num : `${num}/${f.q}`;
}

export function formatPiDraft(multiple: number): string {
  const f = piFractionOf(multiple);
  if (!f) return String(Math.round(multiple * 1000) / 1000);
  return f.q === 1 ? String(f.p) : `${f.p}/${f.q}`;
}

export function parsePiDraft(text: string): number {
  const s = text.trim();
  const slash = s.indexOf("/");
  if (slash < 0) return Number.parseFloat(s);
  const p = Number.parseFloat(s.slice(0, slash));
  const q = Number.parseFloat(s.slice(slash + 1));
  if (!Number.isFinite(p) || !Number.isFinite(q) || q === 0) return Number.NaN;
  return p / q;
}
