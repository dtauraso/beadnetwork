// splitJsonlLines is the pure newline-framing step for stdout: given the carried-over
// partial buffer and a freshly-arrived chunk, it returns every COMPLETE (newline-
// terminated) line and the trailing partial `rest` to carry into the next call. A line
// split across two chunks is reassembled (its bytes accumulate in `rest` until the
// newline arrives); multiple lines in one chunk all come out; a trailing partial is
// buffered. handleStdout owns per-line dispatch; this owns only the framing.
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

// MAX_FRAME_BYTES bounds a single framed-binary record read off ANY dedicated stream fd
// (view/edge/node/interior). MUST match Go's `maxFrameBytes` in nodes/Wiring/stdin_reader.go
// — this is the SAME [len:u32-LE][payload] protocol, just read in the opposite direction
// (Go emits, TS decodes here), and a corrupt/hostile length must be rejected the same way
// on both ends or a bound that only fires going one way is not a bound on the protocol at
// all. Parity is enforced by tools/bridge/check-frame-bytes-parity.sh. Without this bound, a
// corrupt length makes splitFrames' carried-over `rest` grow forever waiting for bytes that
// will never complete the frame — unbounded memory, silently.
export const MAX_FRAME_BYTES = 1 << 20;

// splitFrames is the pure length-prefix framing step shared by every dedicated stream fd
// (view/edge/node/interior): given the carried-over partial Buffer and a freshly-arrived
// binary chunk, it returns every COMPLETE frame payload (as an ArrayBuffer, ready to
// transfer zero-copy) and the trailing partial `rest` to carry into the next call. Frames
// are [len:u32-LE][payload] with NO tag byte (the fd position identifies the stream — see
// Buffer/stream_fds.go). A frame split across two chunks is reassembled; multiple frames in
// one chunk all come out; a trailing partial (len header not yet complete, or payload bytes
// not yet complete) is buffered. Each handleXFd method owns dispatch; this owns only the
// framing.
//
// A decoded length exceeding MAX_FRAME_BYTES is reported via `error` (mirroring Go's
// stdin_reader.go: log and stop, never throw across an event-emitter callback where it
// would be swallowed) and framing STOPS immediately — `rest` is returned as whatever bytes
// preceded the bad header, and no further bytes from `chunk` are scanned. The caller
// (handleXFd) is responsible for ceasing to feed that fd's chunks into splitFrames again;
// see the `deadStreams` field.
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
    // Slice out the payload and copy into a standalone ArrayBuffer (detached from
    // the Node.js Buffer pool so it can be transferred zero-copy to the webview).
    const payload = rest.slice(4, needed);
    const ab = payload.buffer.slice(payload.byteOffset, payload.byteOffset + payload.byteLength);
    frames.push(ab);
    rest = rest.slice(needed);
  }
  return { frames, rest };
}
