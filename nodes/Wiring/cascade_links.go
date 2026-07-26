// cascade_links.go — the static CASCADE-LINK set: the full undirected node-adjacency
// graph MINUS a few cycle-closing links, so that plain "forward to my cascade-link
// neighbors, one hop, excluding the sender, concurrently" (nodeMover.forwardDelta) is
// loop-free BY CONSTRUCTION — no runtime visit-tracking, no once-per-drag guard, no
// forwarding-tree traversal needed at all.
//
// computeCascadeLinks runs ONCE at load (newMoveDispatch), deterministically: build the
// undirected node-adjacency graph from the edge set, then walk a spanning tree by BFS
// starting at the lexicographically-smallest node id, visiting each node's neighbors in
// SORTED id order. Every edge that IS part of that spanning walk is a cascade link; every
// edge that is NOT (the cycle-closing edges) is dropped. The kept (cascade-link) set is
// what's stored and exposed — every node stays reachable via cascade links alone, because
// a spanning walk reaches every node the original graph did.
//
// The result is a write-once set (map[string]bool keyed by an unordered node-id pair) —
// read-only after construction, safe to read from any mover goroutine with no lock, the
// same "load-time-constant table" shape as md.gs (geom_seeds.go) and edgeIDs.
package Wiring

import "sort"

// cascadeLinkKey builds the unordered-pair key used by computeCascadeLinks/isCascadeLink —
// same pair, same key regardless of which endpoint is named first.
func cascadeLinkKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

// computeCascadeLinks builds the cascade-link edge set from the load-time edge
// endpoints. See the file doc comment for the algorithm. Deterministic given the same
// edgeEndpoints.
func computeCascadeLinks(edgeEndpoints map[string]EdgeEndpoints) map[string]bool {
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

	cascadeLinks := map[string]bool{}
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
				cascadeLinks[cascadeLinkKey(cur, next)] = true
				queue = append(queue, next)
			}
		}
	}
	return cascadeLinks
}

// isCascadeLink reports whether the undirected edge between a and b is one of the
// static cascade links computed once at load. Read-only after construction — safe
// from any mover goroutine, no lock.
func (md *MoveDispatch) isCascadeLink(a, b string) bool {
	return md.cascadeLinks[cascadeLinkKey(a, b)]
}
