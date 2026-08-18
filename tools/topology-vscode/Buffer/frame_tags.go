// carries a `//frametag:ts=NAME` marker comment is emitted verbatim as `export const NAME`

package Buffer

//frametag:ts=BUF_BLOCK_TAG_VIEW
const BufBlockTagView byte = 4

//frametag:ts=BUF_VIEW_FRAME_HEADER_SIZE
const BufViewFrameHeaderSize = 8

//frametag:ts=BUF_BLOCK_TAG_EDGE_STREAM
const BufBlockTagEdgeStream byte = 5

//frametag:ts=BUF_EDGE_STREAM_FRAME_HEADER_SIZE
const BufEdgeStreamFrameHeaderSize = 8

//frametag:ts=BUF_BLOCK_TAG_NODE_STREAM
const BufBlockTagNodeStream byte = 6

//frametag:ts=BUF_BLOCK_TAG_INTERIOR_STREAM
const BufBlockTagInteriorStream byte = 7

//frametag:ts=BUF_NODE_STREAM_FRAME_HEADER_SIZE
const BufNodeStreamFrameHeaderSize = 16

//frametag:ts=BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE
const BufInteriorStreamFrameHeaderSize = 4

//frametag:ts=BUF_BLOCK_TAG_BEAD_STREAM
const BufBlockTagBeadStream byte = 8

//frametag:ts=BUF_BEAD_STREAM_FRAME_HEADER_SIZE
const BufBeadStreamFrameHeaderSize = 12
