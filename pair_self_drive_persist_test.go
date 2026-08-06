package main

// pair_self_drive_persist_test.go — the persistence exception (docs/testing-shape.md):
// bytes on disk, through a REAL reload, driving the ACTUAL production path (a real
// LoadTopology, a real MoveDispatch.Start, the real Node1/Node2 Update goroutines, and
// the real editor->Go binary bridge — W.RunStdinReader decoding a framed edit record
// exactly as stdin_reader.go's own doc comment describes) rather than a bare mover
// literal or an in-package short-circuit. task/pair-node-owns-itself removed the
// separate nodeMover goroutine for a pair node (Node1/Node2); this pins that its
// position STILL reaches disk and STILL reloads correctly now that the node's own
// Update goroutine is the sole driver of that state.
//
// This lives in package main (not nodes/Wiring) for the same reason
// kind_registry_parity_test.go does: main is the only package that imports every node
// kind (kinds_generated.go's blank imports), so it is the only place Node1/Node2 are
// actually registered and LoadTopology can build a real pair.
import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring"

	T "github.com/dtauraso/wirefold/Trace"
)

// tiltVectorEntityIndex finds "tiltVector"'s position in the updateKinds= token of the
// real InputLayoutFingerprint, rather than hardcoding it — the fingerprint is the single
// source of truth both languages derive their enum order from (input_codec.go).
func tiltVectorEntityIndex(t *testing.T) byte {
	t.Helper()
	fp := Wiring.InputLayoutFingerprint
	const marker = "updateKinds="
	i := strings.Index(fp, marker)
	if i < 0 {
		t.Fatalf("InputLayoutFingerprint has no %q token: %s", marker, fp)
	}
	rest := fp[i+len(marker):]
	if sp := strings.IndexByte(rest, ' '); sp >= 0 {
		rest = rest[:sp]
	}
	for idx, name := range strings.Split(rest, ",") {
		if name == "tiltVector" {
			return byte(idx)
		}
	}
	t.Fatalf("InputLayoutFingerprint updateKinds= has no tiltVector token: %s", fp)
	return 0
}

// frameTiltVectorTheta builds one framed [len:u32-LE][record] editor->Go bridge record
// for kind=="tiltVector" attr=="theta", dirUp=="up" on the given buffer ROW — the exact
// byte layout decodeInputRecord's "tiltVector"/inTiltVectorAttrTheta case documents:
// [22][entityIdx][4][row][1].
func frameTiltVectorTheta(t *testing.T, row byte) []byte {
	t.Helper()
	const inKindEditUpdate = 22
	const inTiltVectorAttrTheta = 4
	rec := []byte{inKindEditUpdate, tiltVectorEntityIndex(t), inTiltVectorAttrTheta, row, 1}
	frame := make([]byte, 4+len(rec))
	binary.LittleEndian.PutUint32(frame[:4], uint32(len(rec)))
	copy(frame[4:], rec)
	return frame
}

// writePairTree lays down a minimal real Node1/Node2 pair (one bead edge each
// direction, the shape task/tilt-sets-pair-distance's model requires — see
// Wiring.repositionForTiltIndex's own doc comment) so LoadTopology builds the real
// graph, real nodeMovers, and real Node1/Node2 kinds each with a real self-drive claim.
func writePairTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, body string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	mk("nodes/1/meta.json", `{"id":"1","type":"Node1","r":100,"scenePolarR":125.44,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":-0.6981317007977318}`)
	mk("nodes/2/meta.json", `{"id":"2","type":"Node2","r":100,"scenePolarR":125.44,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":0.6981317007977318}`)
	mk("nodes/1/edges/e1.json", `{"label":"e1","kind":"data","sourceHandle":"Out","target":"2","targetHandle":"In"}`)
	mk("nodes/2/edges/e2.json", `{"label":"e2","kind":"data","sourceHandle":"Out","target":"1","targetHandle":"In"}`)
	return root
}

func readNode2Position(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "nodes", "2", "position.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read nodes/2/position.json: %v", err)
	}
	return string(b)
}

// TestPairNodeSelfDrivePersistsThroughRealReload drives the real production path — a
// real LoadTopology, a real MoveDispatch.Start, the real Node1/Node2 Update goroutines
// (no separate nodeMover goroutine exists for either, per task/pair-node-owns-itself:
// mr.start skips a selfDriven node), and the real editor->Go binary bridge
// (W.RunStdinReader decoding a framed tiltVector edit record) — and confirms:
//  1. Neither pair node has a separate mover goroutine (NodeSelfDriven == true for both).
//  2. Node2's own goroutine writes its own position.json in response to the edit (its
//     r-index follows the tilt index, repositionForTiltIndex's model).
//  3. A FRESH LoadTopology of the same root loads that exact persisted offset back — a
//     real reload, not a re-read of the same in-memory struct.
func TestPairNodeSelfDrivePersistsThroughRealReload(t *testing.T) {
	root := writePairTree(t)
	tr := T.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodes, slotReg, md, speedSinks, err := Wiring.LoadTopology(ctx, root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 built nodes, got %d", len(nodes))
	}
	md.EnableEditPersist(root)

	for _, id := range []string{"1", "2"} {
		if !md.NodeSelfDriven(id) {
			t.Fatalf("node %q is not self-driven — task/pair-node-owns-itself requires every PAIR node to own itself, not a separate nodeMover goroutine", id)
		}
	}

	moverWG := md.Start(ctx) // launches 0 mover goroutines here: both pair nodes are self-driven

	var nodeWG sync.WaitGroup
	nodeWG.Add(len(nodes))
	for _, n := range nodes {
		n := n
		go func() {
			defer nodeWG.Done()
			n.Update(ctx)
		}()
	}

	// The real editor->Go bridge: a framed binary record on stdin, decoded and
	// dispatched by the same RunStdinReader the extension host's spawned process reads
	// from (stdin_reader.go's own doc comment). Row 1 == node "2" (ROW ID = NODE ID - 1).
	pr, pw := io.Pipe()
	var stdinWG sync.WaitGroup
	stdinWG.Add(1)
	go func() {
		defer stdinWG.Done()
		Wiring.RunStdinReader(ctx, pr, slotReg, md, tr, speedSinks)
	}()

	before := readNode2Position(t, root)

	if _, err := pw.Write(frameTiltVectorTheta(t, 1)); err != nil {
		t.Fatalf("write tiltVector edit record: %v", err)
	}

	// Poll (bounded, not a fixed sleep-as-barrier) until Node2's own goroutine has
	// written a NEW position.json — the observable effect of its own commit, which
	// (per this package's own persistence writers) happens synchronously with no
	// debounce, so this only ever waits on scheduling, never on a timer this test owns.
	deadline := time.Now().Add(3 * time.Second)
	var after string
	for time.Now().Before(deadline) {
		after = readNode2Position(t, root)
		if after != "" && after != before {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if after == "" || after == before {
		t.Fatalf("node 2's own position.json did not change after a tilt edit: before=%q after=%q", before, after)
	}

	var got struct {
		QuantITheta int `json:"quantITheta"`
		QuantIPhi   int `json:"quantIPhi"`
		QuantIR     int `json:"quantIR"`
	}
	if err := json.Unmarshal([]byte(after), &got); err != nil {
		t.Fatalf("nodes/2/position.json is not valid JSON: %v (%s)", err, after)
	}

	_ = pw.Close()
	cancel()
	moverWG.Wait()
	nodeWG.Wait()
	stdinWG.Wait()

	// A FRESH load of the SAME root — the real reload, not a re-read of md's own live
	// struct — must land on the SAME quantized offset repositionForTiltIndex just wrote.
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	_, _, md2, _, err := Wiring.LoadTopology(ctx2, root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("reload LoadTopology: %v", err)
	}
	iTheta, iPhi, iR, ok := md2.NodeQuantOffset("2")
	if !ok {
		t.Fatal("reload: no nodeMover for node 2")
	}
	if iTheta != got.QuantITheta || iPhi != got.QuantIPhi || iR != got.QuantIR {
		t.Fatalf("reload did not restore the persisted offset: file=(%d,%d,%d) reloaded=(%d,%d,%d)",
			got.QuantITheta, got.QuantIPhi, got.QuantIR, iTheta, iPhi, iR)
	}
}
