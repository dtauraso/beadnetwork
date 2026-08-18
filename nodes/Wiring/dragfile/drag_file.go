package dragfile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

const (
	FileIndexPhi   = "index-phi.json"
	FileIndexTheta = "index-theta.json"
	FileIndexR     = "index-r.json"
	FileTiltIdx    = "top-tilt-vector-phi-idx.json"
)

func DragDir(root, id string) string {
	return filepath.Join(root, "nodes", id, "drag")
}

type JSON struct {
	IndexPhi   int
	IndexTheta int
	IndexR     int

	TopTiltVectorPhiIdx int32
}

func Write(root, id string, j JSON) error {
	dir := DragDir(root, id)
	for name, value := range map[string]int{
		FileIndexPhi:   j.IndexPhi,
		FileIndexTheta: j.IndexTheta,
		FileIndexR:     j.IndexR,
		FileTiltIdx:    int(j.TopTiltVectorPhiIdx),
	} {
		if err := jsonpersist.WriteJSONAtomic(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}

func Read(root, id string) (JSON, bool) {
	dir := DragDir(root, id)
	var j JSON
	found := false
	read := func(name string, dst *int) {
		if jsonpersist.ReadJSONIfExists(filepath.Join(dir, name), dst) {
			found = true
		}
	}
	read(FileIndexPhi, &j.IndexPhi)
	read(FileIndexTheta, &j.IndexTheta)
	read(FileIndexR, &j.IndexR)

	var tilt int
	read(FileTiltIdx, &tilt)
	j.TopTiltVectorPhiIdx = int32(tilt)

	if !found {
		return JSON{}, false
	}
	return j, true
}
