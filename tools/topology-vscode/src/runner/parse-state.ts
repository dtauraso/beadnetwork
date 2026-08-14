export interface StreamParseState {
  stdoutBuf: string;

  viewBuf: Buffer;

  edgeBufs: Buffer[];

  nodeBufs: Buffer[];
  interiorBufs: Buffer[];

  beadBufs: Buffer[];
}

export function freshStreamState(edgeCount: number, nodeCount: number): StreamParseState {
  return {
    stdoutBuf: "",
    viewBuf: Buffer.alloc(0),
    edgeBufs: Array.from({ length: edgeCount }, () => Buffer.alloc(0)),
    nodeBufs: Array.from({ length: nodeCount }, () => Buffer.alloc(0)),
    interiorBufs: Array.from({ length: nodeCount }, () => Buffer.alloc(0)),
    beadBufs: Array.from({ length: nodeCount }, () => Buffer.alloc(0)),
  };
}
