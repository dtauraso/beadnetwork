package nodegeom

// NearestTo picks the id in centers whose world position is closest to p, by squared
// distance — the ordering is the same as true distance and there is no reason to take a
// square root just to compare. Pulled out of moverRegistry.nearestNodeTo
// (nodes/Wiring/mover_registry.go): that method's only touch of moverRegistry's own state
// was building the id->center map it loops over here; the loop itself never read anything
// but its own locals.
func NearestTo(centers map[string]vec3, p vec3) (string, bool) {
	best, bestD2, found := "", 0.0, false
	for id, c := range centers {
		d := c.Sub(p)
		d2 := d.X*d.X + d.Y*d.Y + d.Z*d.Z
		if !found || d2 < bestD2 {
			best, bestD2, found = id, d2, true
		}
	}
	return best, found
}
