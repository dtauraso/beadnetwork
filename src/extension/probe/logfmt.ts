export type LogFields = Record<string, string | number | boolean | undefined | null>;

function encodeValue(v: string | number | boolean): string {
  const s = String(v);
  if (s === "") return `""`;
  if (!/[\s"=]/.test(s)) return s;
  return `"${s.replace(/\\/g, "\\\\").replace(/"/g, '\\"').replace(/\n/g, "\\n")}"`;
}

export function logfmt(fields: LogFields): string {
  const parts: string[] = [];
  for (const [k, v] of Object.entries(fields)) {
    if (v === undefined || v === null) continue;
    parts.push(`${k}=${encodeValue(v)}`);
  }
  return parts.join(" ");
}

export function fieldOf(line: string, key: string): string | undefined {
  const m = new RegExp(`(?:^| )${key}=("(?:[^"\\\\]|\\\\.)*"|[^ ]*)`).exec(line);
  if (!m) return undefined;
  const raw = m[1];
  if (raw === undefined) return undefined;
  if (!raw.startsWith(`"`)) return raw;
  return raw.slice(1, -1).replace(/\\n/g, "\n").replace(/\\"/g, `"`).replace(/\\\\/g, "\\");
}
