---
paths:
  - "Categories/Node/**/*.go"
  - "**/trace_event.go"
  - "**/trace_log.go"
  - "scripts/probe-merge.sh"
---

# Debugging the Go layer (probe breadcrumbs)

Go-side runtime debugging goes through the **DEBUG BREADCRUMB channel** — the Go analogue
of the webview's `postLog`. Call `tr.Breadcrumb(label, node, port, value string)` at a debug
site: it is a structured `Kind==KindBreadcrumb` row (`Categories/Node/trace_event.go`) that the
EMITTING goroutine writes ITSELF, as a fixed-width binary record appended to the file
belonging to the item the event is about:

    topology/view/nodes/<row>/trace.bin           the node
    topology/view/nodes/<row>/interior-trace.bin  its interior
    topology/view/nodes/<row>/beads-trace.bin     its beads
    topology/view/edges/<row>/trace.bin           the edge
    topology/view/trace.bin                       the VIEW owner

One writer per file — the goroutine that owns that item — so the append needs no lock,
the same rule the block files follow. **These events do NOT cross the Go→TS seam.** They
used to ride an events section on every stream frame, which the ext host decoded and
re-encoded as text; that whole path is gone, along with `buffer-log.ts` and the TS event
decoders. A stream frame now carries only a tick.

Read them with `scripts/probe-merge.sh --debug`, which decodes every `trace.bin` at READ
time via `scripts/readtrace` and filters to `kind=="breadcrumb" && debug==true` —
separate from genuine stderr errors (`.probe/go-errors.log`, still plain text). To read one
item on its own: `go run ./scripts/readtrace topology/view/nodes/3/trace.bin`.

Do NOT scatter `fmt.Fprintf(os.Stderr, ...)` for diagnosis; use a breadcrumb. Keep it
SPARSE — it is a debug tool for control events, not a per-tick firehose (see the log-flood
lesson). It is a cheap no-op when no scene root is set (headless tests).

## What the trace setting gates

The `beadnetwork.probe.trace` setting (default **off**) gates the non-breadcrumb bulk of the
trace files plus `ts.log` — the per-tick firehose (recv/send/edge-bead/node-geometry/etc.)
that once grew `go-edge.log` past a gigabyte. It does NOT gate breadcrumb rows: the ext host passes the setting to the Go
child as `BEADNETWORK_PROBE_TRACE` (`Start/extension/runner/lifecycle/process-lifecycle.ts`),
each of the four trace-writing packages reads it once at startup into its own
package-level `traceEnabled` var, and their `Append`/`appendTrace` skips
every non-breadcrumb event when it is off, so `scripts/probe-merge.sh --debug` and the
always-on error logs (`go-errors.log`/`ts-errors.log`/`handler-error-last.log`) work out of
the box on a fresh install with no setting change. Turn the setting on only when you need
the full per-tick trace, not just breadcrumbs.

## The five event kinds

`recv`, `fire`, `send`, `arrive`, `breadcrumb` — that is all of them
(`Categories/Node/trace_event.go`). There is no second, source-level gate: `KindEdgeBead`,
its `edgeBeadTraceEnabled` bool and the `BEADNETWORK_EDGE_BEAD_TRACE` env var were deleted
in `033ef6f99` along with the other 19 kinds that duplicated a column they already rode.
The bead block file that actually renders beads reads no trace flag and is unaffected by
any of this.

On editor hang/decouple/compound symptoms, read the `.probe` ERROR logs first
(`memory/feedback/debugging/probe-logs/feedback_runner_errors_probe_first.md`). For intermittent UI bugs, add cheap
runtime breadcrumbs + a repro before theorizing
(`memory/feedback/debugging/investigation-method/feedback_runtime_breadcrumbs_beat_static_analysis.md`).
