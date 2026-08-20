package kindscan

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
		if name == "" {
			continue
		}
		multi := dir == "broadcast"
		if multi {
			dir = "out"
		}
		if dir != "in" && dir != "out" {
			continue
		}
		ports = append(ports, Port{ID: name, Direction: dir, IsMulti: multi})
	}
	return ports
}
