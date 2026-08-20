export function splitLengthPrefixed(
  carry: Buffer, chunk: Buffer,
): { values: Buffer[]; rest: Buffer } {
  let buf = carry.length === 0 ? chunk : Buffer.concat([carry, chunk]);
  const values: Buffer[] = [];

  for (;;) {
    if (buf.length < 4) break;
    const len = buf.readUInt32LE(0);
    if (buf.length < 4 + len) break;
    values.push(buf.subarray(4, 4 + len));
    buf = buf.subarray(4 + len);
  }
  return { values, rest: buf };
}
