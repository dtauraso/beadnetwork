package kindscan

// parsePortsFromSpec reads nodes/<Kind>/SPEC.md and returns ports derived from
// the Ports table (Name + Direction columns). Used as a fallback when AST
// parsing discovers 0 ports — e.g. when all ports live in an embedded struct
// from another package that the AST walker cannot follow.
func parsePortsFromSpec(pkgDir string) []Port {
	lines, err := readSpecMDLines(pkgDir)
	if err != nil {
		return nil
	}
	tableLines := sectionLines(lines, "Ports")
	if tableLines == nil {
		return nil
	}
	headers, rows := parseMDTable(tableLines)
	nameIdx := indexOf(headers, "Name")
	dirIdx := indexOf(headers, "Direction")
	if nameIdx == -1 || dirIdx == -1 {
		return nil
	}
	var ports []Port
	for _, row := range rows {
		if nameIdx >= len(row) || dirIdx >= len(row) {
			continue
		}
		name := row[nameIdx]
		dir := row[dirIdx]
		if name == "" || (dir != "in" && dir != "out") {
			continue
		}
		ports = append(ports, Port{ID: name, Direction: dir})
	}
	return ports
}
