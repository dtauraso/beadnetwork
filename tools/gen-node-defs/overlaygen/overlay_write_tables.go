package overlaygen

import (
	"bufio"
	"fmt"
)

func writeOverlayDefaultConstructor(w *bufio.Writer, flags []overlayFlag) {
	fmt.Fprintln(w, `// DefaultOverlayState is the startup overlay snapshot used by Wiring's newMoveDispatch.`)
	fmt.Fprintln(w, `func DefaultOverlayState() OverlayState {`)
	fmt.Fprintln(w, "\treturn OverlayState{")
	for _, f := range flags {
		if f.defaultOn {
			fmt.Fprintf(w, "\t\t%s: true,\n", exportField(f.field))
		}
	}
	fmt.Fprintln(w, "\t}")
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
}

func writeOverlayTogglesMap(w *bufio.Writer, flags []overlayFlag) {
	fmt.Fprintln(w, `// OverlayToggles maps an overlay FLAG name (the attr="toggle" wire name) to the`)
	fmt.Fprintln(w, `// OverlayState method that flips it.`)
	fmt.Fprintln(w, `//`)
	fmt.Fprintln(w, `// OVERLAY_TOGGLES_START`)
	fmt.Fprintln(w, `var OverlayToggles = map[string]func(*OverlayState, *T.Trace){`)
	for _, f := range flags {
		fmt.Fprintf(w, "\t%q: (*OverlayState).Toggle%s,\n", f.flag, f.method)
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `// OVERLAY_TOGGLES_END`)
	fmt.Fprintln(w)
}

func writeOverlayBreadcrumbTables(w *bufio.Writer, flags []overlayFlag) {
	fmt.Fprintln(w, `// OverlayFlagBreadcrumbScope names the overlay flags whose Toggle also logs a`)
	fmt.Fprintln(w, `// structured "pole-toggle-go" debug breadcrumb, and the scope text that breadcrumb`)
	fmt.Fprintln(w, `// carries. Flags absent here emit no such breadcrumb.`)
	fmt.Fprintln(w, `//`)
	fmt.Fprintln(w, `// OVERLAY_BREADCRUMB_SCOPES_START`)
	fmt.Fprintln(w, `var OverlayFlagBreadcrumbScope = map[string]string{`)
	for _, f := range flags {
		if f.breadcrumb != "" {
			fmt.Fprintf(w, "\t%q: %q,\n", f.flag, f.breadcrumb)
		}
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `// OVERLAY_BREADCRUMB_SCOPES_END`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `// OverlayFlagValue reads the post-toggle bool for the flags named in`)
	fmt.Fprintln(w, `// OverlayFlagBreadcrumbScope (same key set).`)
	fmt.Fprintln(w, `var OverlayFlagValue = map[string]func(*OverlayState) bool{`)
	for _, f := range flags {
		if f.breadcrumb != "" {
			fmt.Fprintf(w, "\t%q: func(o *OverlayState) bool { return o.%s },\n", f.flag, exportField(f.field))
		}
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
}

func writeOverlayTraceKindMap(w *bufio.Writer, flags []overlayFlag) {
	fmt.Fprintln(w, `// OverlayFlagTraceKind maps the wire FLAG name (same keys as OverlayToggles) to its`)
	fmt.Fprintln(w, `// Trace.Kind* string, so Wiring's applyUpdate case can hand EmitViewFrame the ONE event`)
	fmt.Fprintln(w, `// that flag's toggle logged (matching the per-toggle tr.X(bool) call).`)
	fmt.Fprintln(w, `// Referencing T.Kind<Method> by name means a flag missing its Trace kind constant is a`)
	fmt.Fprintln(w, `// Go compile error here, not a silent no-op — see writeOverlayGen in`)
	fmt.Fprintln(w, `// tools/gen-node-defs/overlay_gen.go.`)
	fmt.Fprintln(w, `//`)
	fmt.Fprintln(w, `// OVERLAY_TRACE_KINDS_START`)
	fmt.Fprintln(w, `var OverlayFlagTraceKind = map[string]string{`)
	for _, f := range flags {
		fmt.Fprintf(w, "\t%q: T.Kind%s,\n", f.flag, f.method)
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `// OVERLAY_TRACE_KINDS_END`)
	fmt.Fprintln(w)
}
