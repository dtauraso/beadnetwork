// port_geom_emit.go — the port-geometry helper half of emit_geometry.go, split out as a
// pure move (no logic changes): partnerCenterFn/buildPartnerCenterFn, aimedPortPosDir,
// buildPortGeoms, effectiveRadius. See interior_stream.go for the interiorStream I/O type
// and bead_emit.go for the bead-emission helpers; builders.go keeps the reflection-driven
// port-manifest/node-construction half.

package Wiring

import (
	T "github.com/dtauraso/wirefold/Trace"
)

// partnerCenterFn returns the CURRENT world center of the single partner node connected
// to (port, isInput) via one edge — the aimed-port model's one input. ok is false for an
// edgeless port (no partner), which falls back to ring placement. Built once per node by
// newMoveDispatch (dynamic, atomic-snapshot-backed) or once per node at initial construction
// (static, straight off the loaded geoms) — see buildPartnerCenterFn.
type partnerCenterFn func(port string, isInput bool) (vec3, bool)

// buildPartnerCenterFn returns a partnerCenterFn for nodeID: it scans edgeEndpoints (the
// static edge-label → source/target/handle map) for the one edge touching (port, isInput) on
// nodeID and, if found, resolves the partner's current center via centerOf. This is the ONE
// place the (node,port,isInput) → partner-id lookup lives, shared by the static
// (construction-time) and dynamic (mover, atomic-snapshot) callers so both agree.
func buildPartnerCenterFn(nodeID string, edgeEndpoints map[string]EdgeEndpoints, centerOf func(id string) vec3) partnerCenterFn {
	return func(port string, isInput bool) (vec3, bool) {
		for _, ep := range edgeEndpoints {
			if !isInput && ep.Source == nodeID && ep.SourceHandle == port {
				return centerOf(ep.Target), true
			}
			if isInput && ep.Target == nodeID && ep.TargetHandle == port {
				return centerOf(ep.Source), true
			}
		}
		return vec3{}, false
	}
}

// aimedPortPosDir returns the port-position/direction closure used by BOTH the node's own
// live geometry re-derive (nodeMover.writeStreamFrame, node_mover.go) and the load-time row
// seed (newMoveDispatch's md.nodeSeeds, node_move.go) — the ONE place aimed-vs-static port
// placement is computed, so seed and live emit can never drift apart. partnerCenter may
// be nil (no edges known / test callers), in which case every port takes the ring-placement
// fallback.
func aimedPortPosDir(g nodeGeom, partnerCenter partnerCenterFn) func(name string, isInput bool) (vec3, vec3) {
	return func(name string, isInput bool) (vec3, vec3) {
		var pc vec3
		hasPartner := false
		if partnerCenter != nil {
			pc, hasPartner = partnerCenter(name, isInput)
		}
		pos := portWorldPosAimed(g, name, isInput, pc, hasPartner)
		if hasPartner {
			if dirVec := pc.sub(nodeWorldPos(g)); dirVec.length() >= portDegenerateEps {
				return pos, dirVec.normalize()
			}
		}
		dir, _ := portDir(g, name, isInput)
		return pos, dir
	}
}

// buildPortGeoms derives the full port-geometry slice (input ports then output ports) from g
// and a port-position/direction function. Shared by nodeMover.writeStreamFrame (live re-derive)
// and the load-time row seed so both agree on port order and values.
func buildPortGeoms(g nodeGeom, portPosDir func(name string, isInput bool) (pos, dir vec3)) []T.PortGeom {
	ports := make([]T.PortGeom, 0, len(g.Inputs)+len(g.Outputs))
	appendPort := func(name string, isInput bool) {
		pos, dir := portPosDir(name, isInput)
		ports = append(ports, T.PortGeom{
			Name: name, IsInput: isInput,
			PX: pos.X, PY: pos.Y, PZ: pos.Z,
			DX: dir.X, DY: dir.Y, DZ: dir.Z,
		})
	}
	for _, p := range g.Inputs {
		appendPort(p.Name, true)
	}
	for _, p := range g.Outputs {
		appendPort(p.Name, false)
	}
	return ports
}

// effectiveRadius returns the node's REACH radius (max distance to a surface child),
// falling back to nodeR for childless nodes (ReachR == 0) so the value stays sane.
// Used by nodeMover.writeStreamFrame (sphereR).
func effectiveRadius(g nodeGeom) float64 {
	if g.ReachR > 0 {
		return g.ReachR
	}
	return nodeR(g)
}
