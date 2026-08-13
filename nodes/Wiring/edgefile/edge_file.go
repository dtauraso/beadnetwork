package edgefile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

type edgeFile struct {
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle"`
	Kind         string `json:"kind"`
	Label        string `json:"label"`

	// D — the vector from this edge's source to its target. See
	// nodes/Wiring/loadspec/edge_delta.go: A + D = B.
	DeltaPolarR     *float64 `json:"deltaPolarR,omitempty"`
	DeltaPolarPhi   *float64 `json:"deltaPolarPhi,omitempty"`
	DeltaPolarTheta *float64 `json:"deltaPolarTheta,omitempty"`
}

func edgeFilePath(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "edges", label+".json")
}

func edgesDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id, "edges")
}

// SourceHandleFor names the handle a NEW edge leaves srcPort through, and refuses when the
// port cannot carry another edge. A PortBroadcast handle is the port name plus the next free
// index ("ToNext0", "ToNext1") — BroadcastBaseName strips that digit back off on load, and a
// bare name would land in the same bucket without one. A PortOut carries exactly one edge:
// build_nodes.go binds labels[0] and drops the rest, so a second edge from an occupied
// PortOut is a file on disk that never becomes a wire. Refusing says so; writing it does not.
func SourceHandleFor(root, src, srcPort string, broadcast bool) (string, string, bool) {
	used := countHandlesOn(root, src, srcPort)
	if broadcast {
		return fmt.Sprintf("%s%d", srcPort, used), "", true
	}
	if used > 0 {
		return "", fmt.Sprintf("%s's %s is already connected, and it carries one edge", src, srcPort), false
	}
	return srcPort, "", true
}

// countHandlesOn counts the edges already leaving srcPort, matching both the bare port name
// and its indexed broadcast forms.
func countHandlesOn(root, src, srcPort string) int {
	entries, err := os.ReadDir(edgesDirPath(root, src))
	if err != nil {
		return 0
	}
	used := 0
	for _, e := range entries {
		if e.IsDir() { // path-resolution-ok: skipping a stray directory, not resolving a scene path
			continue
		}
		var ef edgeFile
		jsonpersist.ReadJSONBestEffort(filepath.Join(edgesDirPath(root, src), e.Name()), &ef)
		if handleIsOn(ef.SourceHandle, srcPort) {
			used++
		}
	}
	return used
}

func handleIsOn(handle, srcPort string) bool {
	if handle == srcPort {
		return true
	}
	rest, cut := strings.CutPrefix(handle, srcPort)
	if !cut || rest == "" {
		return false
	}
	for _, r := range rest {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func WriteEdgeFile(root, src, srcPort, target, targetPort string) error {
	label := src + "To" + target
	return jsonpersist.WriteJSONAtomic(edgeFilePath(root, src, label), edgeFile{
		SourceHandle: srcPort, Target: target, TargetHandle: targetPort, Kind: "chain", Label: label,
	})
}

// WriteEdgeDelta records D on an edge that already exists, leaving every other
// key as it was. Its ONLY caller is that edge's own edgeMover — the single
// writer of nodes/<source>/edges/<label>.json.
func WriteEdgeDelta(root, src, label string, d polar.Polar) error {
	path := edgeFilePath(root, src, label)
	var ef edgeFile
	if !jsonpersist.ReadJSONIfExists(path, &ef) {
		// No file means no edge to annotate. An edge is authored before it can
		// move, so this is a missing file, not a first write.
		return fmt.Errorf("WriteEdgeDelta: no edge file at %s", path)
	}
	ef.DeltaPolarR, ef.DeltaPolarPhi, ef.DeltaPolarTheta = &d.R, &d.Phi, &d.Theta
	return jsonpersist.WriteJSONAtomic(path, ef)
}

func RemoveEdgesTo(root, id string, nodeIDs []string) error {
	for _, n := range nodeIDs {
		entries, err := os.ReadDir(edgesDirPath(root, n))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() { // path-resolution-ok: skipping a stray directory, not resolving a scene path
				continue
			}
			path := filepath.Join(edgesDirPath(root, n), e.Name())
			var ef edgeFile
			jsonpersist.ReadJSONBestEffort(path, &ef)
			if ef.Target != id {
				continue
			}
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}
