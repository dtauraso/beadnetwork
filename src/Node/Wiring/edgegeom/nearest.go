package edgegeom

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
