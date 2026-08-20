// A probe log line is flat key=value pairs, space separated — every record the
// decoder emits is scalars only, so there is nothing for a document format to
// describe. Values containing a space, a quote or an equals sign are quoted.
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

// fieldOf reads one key back out of a line — the only read the pipeline needs,
// for routing a line to the error log by its label.
export function fieldOf(line: string, key: string): string | undefined {
  const m = new RegExp(`(?:^| )${key}=("(?:[^"\\\\]|\\\\.)*"|[^ ]*)`).exec(line);
  if (!m) return undefined;
  const raw = m[1];
  if (raw === undefined) return undefined;
  if (!raw.startsWith(`"`)) return raw;
  return raw.slice(1, -1).replace(/\\n/g, "\n").replace(/\\"/g, `"`).replace(/\\\\/g, "\\");
}
