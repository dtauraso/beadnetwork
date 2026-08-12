// port_geometry.go — Go-OWNED port-to-port segment geometry.
//
// Go owns geometry in this architecture: it computes each node/port world position
// and streams the result into the content buffer; the TS webview renders from that
// buffer and computes no geometry of its own (guard: check-ts-computes-no-geometry.sh).
// This file is NOT a mirror of a TS counterpart — the former TS port-geometry helpers
// were removed when Go took over, and geometry-helpers.ts now holds only screen-coord
// conversion (ndcToPixel/pixelToNDC) for input picking, nothing this file reproduces.
//
// Go must compute a pulse's travel budget from the SAME segment the
// bead is drawn on: a straight line from the source OUTPUT port's sphere-surface
// point to the target INPUT port's sphere-surface point. NodeWorldPos, NodeRadius,
// portDir and portWorldPos here feed arcLengthBetweenPorts
// (loader.go / stdin_reader.go), which returns the chord length.
//
// Inputs the geometry needs, per node:
//   - kind        → width/height via KindDims (generated from SPEC.md View)
//   - center      → world center (from meta.json x/y/z or origin fallback)
//   - port lists  → inputs/outputs with optional side + slot (from the spec node;
//                   falls back to registry ports with default sides when absent)
//
// Every magic number is pulled from CurveParam* constants in curve_params.go.
//
// A node's OWN geometry (identity, polar position, sphere radius) lives in
// node_geom.go; the interior-bead 2x2 slot grid lives in interior_slot_geometry.go.
// This file keeps the seam it is named for: the geometry BETWEEN two nodes — the edge
// segment itself and the live center-to-center distance/direction measurement it and
// chain_beads.go both read from. The mutual-pair parallel-chain offset is its own
// concern, split out to parallel_chain_offset.go; ring-axis derivation is in
// ring_axis.go.

package nodegeom

// EdgeSegment is the straight world segment the renderer draws for an edge: NODE SURFACE
// TO NODE SURFACE along the centre-to-centre line (docs/bead-model/channels-not-ports.md — a port is
// a load-time channel-binding ROLE now, never a place, so it contributes no geometry to
// this segment at all). start = the source node's center, moved out to its own
// NodeTorusOuterR toward the target; end = the target's center, moved out to ITS
// NodeTorusOuterR toward the source. These are the SAME two surface points
// chain_beads.go anchors bead 0 and the last bead to (docs/bead-model/bead-lattice.md "Placement":
// "Bead 0's torus is tangent to the source node's torus... bead N-1's torus is tangent to
// the target node's torus, EXACTLY") — this is deliberate, not incidental: the edge
// segment and the bead chain must measure between the identical two points, which is
// exactly the invariant the old port-radius offset broke (the chain measured node-torus
// to node-torus while the port sat proud of/inside that surface, so the first and last
// bead were off by the port's own radius while interior spacing stayed correct).
func EdgeSegment(src, tgt NodeGeom) wireSegment {
	srcCenter := NodeWorldPos(src)
	tgtCenter := NodeWorldPos(tgt)
	dir := tgtCenter.Sub(srcCenter)
	if dir.Length() < 1e-9 {
		// Degenerate (coincident centers, e.g. a not-yet-positioned node): fall back to the
		// bare centers rather than dividing by a near-zero length.
		return wireSegment{Start: srcCenter, End: tgtCenter}
	}
	unit := dir.Normalize()
	start := srcCenter.Add(unit.Scale(NodeTorusOuterR(src.Kind)))
	end := tgtCenter.Sub(unit.Scale(NodeTorusOuterR(tgt.Kind)))
	return wireSegment{Start: start, End: end}
}

// EdgeCenterDistAndDir returns the LIVE center-to-center distance BETWEEN two nodes AND
// the live unit direction from selfCenter toward targetCenter, from their live cartesian
// world centers — ONE measurement of the edge, not two. selfCenter/targetCenter must be
// NodeWorldPos of each node, the SAME function EdgeSegment (above) and every emitGeometry
// call use, so this reads the identical value the renderer draws the node at — not the
// SOURCE node's stored, quantized LocalPolar (lp.QuantIR*StepR and its QuantITheta/
// QuantIPhi bearing), which is an integer-step APPROXIMATION of both this distance and
// this direction (1-degree angular cells), which can drift from the live geometry between
// drags. chain_beads.go reads the LIVE value (EdgeStepCount's `dist`) rather than the
// stored one, so a bearing residue (a chain aimed up to half a degree off, from the
// 1-degree stored cell) can never reappear independent of the length: distance and
// direction are returned from the SAME Length()/Sub() call, so a caller cannot read a
// length from one measurement and a bearing from another.
//
// ok is false only when the centers are degenerate (coincident, e.g. a
// not-yet-positioned node with HasPos false) — direction is undefined at zero
// separation, and the caller falls back to the stored quantized bearing/distance rather
// than dividing by a near-zero length.
//
// This one Length()/Normalize() pair is deliberately NOT in chain_beads.go: that file is
// guarded against math.Sqrt/Vec3.Length/Normalize
// (tools/network/beads/check-no-sqrt-in-chain-beads.sh, "index arithmetic, trig only at the
// polar2cart boundary" — memory/feedback_abc_times_constant_not_rederive.md).
// chainBeads calls this helper and receives only the resulting scalar distance and unit
// vector; the sqrt itself lives here, in the file that already computes EdgeSegment the
// same way.
func EdgeCenterDistAndDir(selfCenter, targetCenter vec3) (dist float64, unitDir vec3, ok bool) {
	delta := targetCenter.Sub(selfCenter)
	length := delta.Length()
	if length < 1e-9 {
		return 0, vec3{}, false
	}
	return length, delta.Normalize(), true
}
