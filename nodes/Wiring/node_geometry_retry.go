// node_geometry_retry.go — a nodeGeometry's own outbound-message retry queue drain.
package Wiring

// flushPending retries every message in m.msg.pending in order, attempting a non-blocking
// send to its destination's inbox. A destination whose channel is momentarily full
// stays in the queue (retried next call) — and so does every LATER item addressed to
// that SAME destination, even if its own channel isn't full, so per-destination FIFO
// is preserved (a retained item is never overtaken by a newer one to the same
// destination). An item whose destination doesn't resolve (unknown id) is dropped,
// matching the old deliverMove no-op for an unknown id. Called only from m's own
// driving goroutine (sendMove, at enqueue time, and the driving loop, every cycle).
func (m *nodeGeometry) flushPending() {
	if len(m.msg.pending) == 0 || m.msg.resolveDest == nil {
		return
	}
	blocked := map[string]bool{}
	kept := m.msg.pending[:0]
	for _, item := range m.msg.pending {
		if blocked[item.destID] {
			kept = append(kept, item)
			continue
		}
		trySend, ok := m.msg.resolveDest(item.destID)
		if !ok {
			continue
		}
		if !trySend(item.msg) {
			blocked[item.destID] = true
			kept = append(kept, item)
		}
	}
	m.msg.pending = kept
}
