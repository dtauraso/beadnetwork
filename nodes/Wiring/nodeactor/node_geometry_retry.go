package nodeactor

func (m *NodeGeometry) flushPending() {
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
