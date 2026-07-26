// dead_end_edges.go — the static dead-end-edge set that makes delta-forward propagation
// a TREE instead of a graph with cycles, so no runtime visit-tracking/once-per-drag guard
// is needed at all (see moveMsgKindDeltaForward's doc comment, move_msg.go, and
// nodeMover.forwardDelta, node_mover.go).
//
// computeDeadEndEdges runs ONCE at load (newMoveDispatch), deterministically: build the
// undirected node-adjacency graph from the edge set, then walk a spanning tree by BFS
// starting at the lexicographically-smallest node id, visiting each node's neighbors in
// SORTED id order. Every edge that is NOT part of that spanning tree is a dead-end edge —
// a cycle-closing edge. Forwarding never crosses a dead-end edge; every node stays
// reachable regardless, because a dead-end edge is by construction NOT needed for
// reachability (removing only non-tree edges leaves a spanning tree, which reaches every
// node the original graph did).
//
// The result is a write-once set (map[string]bool keyed by an unordered node-id pair) —
// read-only after construction, safe to read from any mover goroutine with no lock, the
// same "load-time-constant table" shape as md.gs (geom_seeds.go) and edgeIDs.
package Wiring

import "sort"

// deadEndKey builds the unordered-pair key used by computeDeadEndEdges/isDeadEndEdge —
// same pair, same key regardless of which endpoint is named first.
func deadEndKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

// computeDeadEndEdges builds the dead-end edge set from the load-time edge endpoints. See
// the file doc comment for the algorithm. Deterministic given the same edgeEndpoints.
func computeDeadEndEdges(edgeEndpoints map[string]EdgeEndpoints) map[string]bool {
	adj := map[string][]string{}
	nodeSet := map[string]bool{}
	for _, ep := range edgeEndpoints {
		nodeSet[ep.Source] = true
		nodeSet[ep.Target] = true
		adj[ep.Source] = append(adj[ep.Source], ep.Target)
		adj[ep.Target] = append(adj[ep.Target], ep.Source)
	}
	for id := range adj {
		sort.Strings(adj[id])
	}
	nodeIDs := make([]string, 0, len(nodeSet))
	for id := range nodeSet {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)

	treeEdges := map[string]bool{}
	visited := map[string]bool{}
	for _, start := range nodeIDs {
		if visited[start] {
			continue
		}
		visited[start] = true
		queue := []string{start}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, next := range adj[cur] {
				if visited[next] {
					continue
				}
				visited[next] = true
				treeEdges[deadEndKey(cur, next)] = true
				queue = append(queue, next)
			}
		}
	}

	deadEnds := map[string]bool{}
	for _, ep := range edgeEndpoints {
		key := deadEndKey(ep.Source, ep.Target)
		if !treeEdges[key] {
			deadEnds[key] = true
		}
	}
	return deadEnds
}

// isDeadEndEdge reports whether the undirected edge between a and b is one of the static
// dead-end (non-spanning-tree) edges computed once at load. Read-only after construction —
// safe from any mover goroutine, no lock.
func (md *MoveDispatch) isDeadEndEdge(a, b string) bool {
	return md.deadEndEdges[deadEndKey(a, b)]
}
