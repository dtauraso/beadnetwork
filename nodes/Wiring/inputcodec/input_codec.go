// input_codec.go — BINARY decode of the editor→Go input stream.
//
// The TS→Go bridge is a purely BINARY buffer, symmetric with the Go→TS content
// buffer streamed on fd 3. The webview builds a binary RECORD per message and the
// extension host writes each record FRAMED as [len:u32-LE][record] to Go's stdin.
// This file decodes one record (kind byte + fixed numeric fields + length-prefixed
// UTF-8 sections) back into the SAME StdinMsg the old newline-JSON path produced,
// so Wiring's applyEdit / HandleRawInput dispatch is UNCHANGED — only the wire
// decode differs.
//
// The record layout is defined ONCE here. The TS side
// (tools/topology-vscode/src/schema/input-layout-gen.ts) is GENERATED from this file's
// InputLayoutFingerprint by tools/gen-node-defs, so it cannot hand-drift from Go.
//
// Numbers are little-endian (matching the content buffer's little-endian encoding). Enum discriminators
// (event kind, hit kind, update entity kind, update attr, overlay flag) are u8 indices
// into the shared orderings. There is NO JSON on the wire: every record is fully numeric.
// The live editor→Go traffic is raw-input, overlays toggle (numeric flag-id), and
// the bare `save` COMMAND (Go persists its OWN authoritative scene state).
// edit-create and edit-delete record kinds were REMOVED end-to-end — no live TS
// sender ever emitted them, and their only trigger (a port-drop gesture calling
// PacedWire.Restore()) unconditionally tore down a live wire's in-flight beads. Their
// kind bytes (20, 21) are left as GAPS below, never renumbered. Only edit-update
// remains.
//
// Kind 3 was inKindResend (removed: the ext host now caches the last stream frame and
// replays it on webview "ready" instead of asking Go to re-emit geometry — see
// runCommand.ts's BuildAndRunRunner.lastSnapshot/getLastSnapshot). Left as an intentional
// GAP rather than renumbered, so no other kind's wire value moves.
//
// "Defined ONCE here" is the FILE SET this header opens, not this one file: the layout PIN
// (InputLayoutFingerprint, the kind bytes, the attr indices, the enum orderings) is
// input_fingerprint.go — the file tools/gen-node-defs reads; the byte-reading primitives
// are nodes/Wiring/recread; the raw-input record's flat field run is raw_input_decode.go; and the
// addressed edit's entity→attribute decoders are edit_update_decode.go. What is left in
// THIS file is the one thing that has to know all four: the top-level KIND BYTE dispatch.

package inputcodec

import "github.com/dtauraso/wirefold/nodes/Wiring/recread"

// DecodeInputRecord decodes one deframed record body (WITHOUT the [len] frame) into a
// StdinMsg. ok=false means the record was malformed/unknown and must be ignored
// (forward-compatible; mirrors the old json.Unmarshal-error `continue`).
func DecodeInputRecord(rec []byte) (StdinMsg, bool) {
	if len(rec) == 0 {
		return StdinMsg{}, false
	}
	r := recread.NewReader(rec, 1)
	switch rec[0] {
	case InKindSave:
		return StdinMsg{Type: "save"}, true
	case InKindRawInput:
		ev, ok := decodeRawInput(r)
		if !ok {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "raw-input", Event: &ev}, true
	case InKindEditUpdate:
		return decodeEditUpdate(r)
	}
	return StdinMsg{}, false
}
