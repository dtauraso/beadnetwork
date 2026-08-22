package nodedefs

import (
	"fmt"
	"strings"

	"github.com/dtauraso/wirefold/NodeKinds/gen/kindscan"
)

func buildDef(v kindscan.ViewDef, ports []kindscan.Port) string {
	targets := filterPorts(ports, "in")
	sources := filterPorts(ports, "out")

	var fields []string
	fields = append(fields, fmt.Sprintf(`bg: "%s"`, v.Bg))
	fields = append(fields, fmt.Sprintf(`border: "%s"`, v.Border))
	fields = append(fields, fmt.Sprintf(`text: "%s"`, v.Text))
	if v.MinWidth != "" {
		fields = append(fields, fmt.Sprintf(`minWidth: %s`, v.MinWidth))
	}
	if v.Shape != "" {
		fields = append(fields, fmt.Sprintf(`shape: "%s"`, v.Shape))
	}
	if v.Fill != "" {
		fields = append(fields, fmt.Sprintf(`fill: "%s"`, v.Fill))
	}
	if v.Stroke != "" {
		fields = append(fields, fmt.Sprintf(`stroke: "%s"`, v.Stroke))
	}
	if v.Width != "" {
		fields = append(fields, fmt.Sprintf(`width: %s`, v.Width))
	}
	if v.Height != "" {
		fields = append(fields, fmt.Sprintf(`height: %s`, v.Height))
	}
	if v.Desc != "" {

		fields = append(fields, fmt.Sprintf(`desc: %q`, v.Desc))
	}

	if len(targets) > 0 {
		fields = append(fields, fmt.Sprintf(`inputs: [%s]`, joinPortsTyped(targets)))
	}
	if len(sources) > 0 {
		fields = append(fields, fmt.Sprintf(`outputs: [%s]`, joinPortsTyped(sources)))
	}
	return "{ " + strings.Join(fields, ", ") + " }"
}

func filterPorts(ports []kindscan.Port, dir string) []kindscan.Port {
	var out []kindscan.Port
	for _, p := range ports {
		if p.Direction == dir {
			out = append(out, p)
		}
	}
	return out
}

func joinPortsTyped(ports []kindscan.Port) string {
	var parts []string
	for _, p := range ports {
		ek := p.EdgeKind
		if ek == "" {
			ek = "chain"
		}
		if p.IsMulti {
			parts = append(parts, fmt.Sprintf(`{ name: "%s", kind: "%s", isMulti: true }`, p.ID, ek))
		} else {
			parts = append(parts, fmt.Sprintf(`{ name: "%s", kind: "%s" }`, p.ID, ek))
		}
	}
	return strings.Join(parts, ", ")
}
