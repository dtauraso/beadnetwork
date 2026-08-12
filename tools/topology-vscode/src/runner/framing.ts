





export function splitJsonlLines(buf: string, chunk: string): { lines: string[]; rest: string } {
  let rest = buf + chunk;
  const lines: string[] = [];
  let nl: number;
  while ((nl = rest.indexOf("\n")) !== -1) {
    lines.push(rest.slice(0, nl));
    rest = rest.slice(nl + 1);
  }
  return { lines, rest };
}









export const MAX_FRAME_BYTES = 1 << 20;

















export function splitFrames(buf: Buffer, chunk: Buffer): { frames: ArrayBuffer[]; rest: Buffer; error?: string } {
  let rest = buf.length > 0 ? Buffer.concat([buf, chunk]) : chunk;
  const frames: ArrayBuffer[] = [];
  while (rest.length >= 4) {
    const frameLen = rest.readUInt32LE(0);
    if (frameLen > MAX_FRAME_BYTES) {
      return { frames, rest, error: `bad frame length ${frameLen} (max ${MAX_FRAME_BYTES}); stopping stream` };
    }
    const needed = 4 + frameLen;
    if (rest.length < needed) break;


    const payload = rest.slice(4, needed);
    const ab = payload.buffer.slice(payload.byteOffset, payload.byteOffset + payload.byteLength);
    frames.push(ab);
    rest = rest.slice(needed);
  }
  return { frames, rest };
}
