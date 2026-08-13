// carries a `//frametag:ts=NAME` marker comment is emitted verbatim as `export const NAME`

package Buffer

//frametag:ts=BUF_BLOCK_TAG_VIEW
const BufBlockTagView byte = 4

// [tick:u32][layoutFingerprintHash:u32] — the hash lets the decoder refuse a
// frame whose buffer layout it was not built for. See BufLayoutFingerprintHash.
//
//frametag:ts=BUF_VIEW_FRAME_HEADER_SIZE
const BufViewFrameHeaderSize = 8

//frametag:ts=BUF_BLOCK_TAG_EDGE_STREAM
const BufBlockTagEdgeStream byte = 5

// tick, then the label byte count, then how many beads are in flight.
//
//frametag:ts=BUF_EDGE_STREAM_FRAME_HEADER_SIZE
const BufEdgeStreamFrameHeaderSize = 12

//frametag:ts=BUF_BLOCK_TAG_NODE_STREAM
const BufBlockTagNodeStream byte = 6

//frametag:ts=BUF_BLOCK_TAG_INTERIOR_STREAM
const BufBlockTagInteriorStream byte = 7

// tick, then the label byte count. Beads and poles left the node frame with
// the neighbour positions they were laid out against.
//
//frametag:ts=BUF_NODE_STREAM_FRAME_HEADER_SIZE
const BufNodeStreamFrameHeaderSize = 8

//frametag:ts=BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE
const BufInteriorStreamFrameHeaderSize = 4
