package Node

import (
	"path/filepath"
)

const (
	FileType       = "type.bin"
	FileGate       = "gate.bin"
	FileIndexPhi   = "index-phi.bin"
	FileIndexTheta = "index-theta.bin"
	FileIndexR     = "index-r.bin"
	FileAtPhi      = "at-phi.bin"
	FileAtTheta    = "at-theta.bin"
	FileAtR        = "at-r.bin"
	FileS          = "s.bin"
	FileRefPhi     = "ref-phi.bin"
	FileRefTheta   = "ref-theta.bin"
	FileRefS       = "ref-s.bin"
	FileTiltIdx    = "top-tilt-vector-phi-idx.bin"
)

func BaseDir(nodeDir string) string { return filepath.Join(nodeDir, "base") }
