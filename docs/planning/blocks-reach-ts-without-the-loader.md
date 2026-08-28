# Blocks reach TS without VS Code's resource loader

Intent, not status. Delete this file when the change lands.

## The target

Go keeps writing block files exactly as it does now. The webview stops **fetching** them.
The extension host reads each block file from disk and sends its bytes to the webview; the
webview receives instead of asking.

What does NOT change:

- Go writes the same files, one writer per file, no lock, same paths, same layout.
- `*_values.go`, the generators, the block layout, `value_file.go` — untouched.
- The TS→Go direction (framed binary records on stdin, fire-and-forget) — untouched.
- A block is still the current value. Nothing becomes a stream of deltas.

What changes:

- `leaf-values.ts` and `row-leaf-values.ts` stop calling `fetch` and take bytes handed to
  them instead. The `LeafValues`/`RowLeafValues` shape their callers use stays identical.
- The extension host gains a reader: it watches or polls the block files itself and posts
  `{ path, bytes }` to the webview.
- `localResourceRoots` stops mattering for data. The bundle is already inlined in the html.

## Why

VS Code's webview resource loader stops serving this panel. Measured, on 2026-08-27:

- Every request from our panel — the 1.2MB bundle and two 60-byte scene files alike — is
  issued and then neither served nor refused. No error event ever fires; `readyState` stays
  `loading`.
- A 20-line probe extension opened a panel in the SAME window at the SAME time and its
  fetch returned `HTTP 200 in 9ms`. The loader is healthy; it is our panel it will not serve.
- Not the roots, the options, the CSP, the nonce, the symlinked path, the query string, the
  service worker, or the host's event loop — each tested and eliminated, several twice.
- Inlining the bundle into the html boots the editor: `bundle-eval → before-render →
  ready-sent`, and it takes pointer input. It then sits on "starting the network…" because
  the readers still fetch, and those fetches hang.

So the boot no longer needs the loader and the data still does. This change removes the
last dependency on it.

The traffic is also worth removing on its own: 15 leaf readers and 4 row readers issue
~260 requests/s after the cadence cut (~730 before), forever, whether or not a byte changed.
Each one costs a service-worker hop, a file read, and a buffer in VS Code's main process —
the process whose growth we spent an evening chasing.

## What breaks, and in what order

1. **`leaf-values.ts` × 20 and `row-leaf-values.ts` × 16.** These are deliberate copies, one
   per concern. Every one needs the same edit. The guard is that they stay identical: a copy
   left on `fetch` will silently keep polling and appear to work until the loader stops
   serving that panel too.
2. **The host reader.** New code in `Start/extension/`. It needs the scene root, the list of
   block paths (from each concern's `paths/block.bin`, the same file the readers read today),
   and a cadence. This is the one genuinely new thing.
3. **Row blocks.** `view/nodes/{row}/node.bin` and friends are templated by row. The host has
   to know which rows exist — `counts/nodes.bin` already answers that, and the row space is
   the largest id, not a live count.
4. **`webview-options.ts`.** Roots can shrink to nothing once no resource is fetched. Do this
   LAST and separately; it is the one step that cannot be tested by reading.
5. **The scene switch.** `view/scene/selected.bin` lives at the anchor and decides which scene
   loads. The host reads it today only to build the html; it becomes a block like any other.

## Verification

- `bash scripts/stop-checks.sh` empty, at every step.
- The editor draws the pair scene: two nodes, an edge, chrome. This is the check that has
  been missing all along — the Go side is verified headless, the render half is not.
- `.probe/ts.log` shows `bundle-eval → before-render → ready-sent` and then node draws.
- `scripts/watch-file-growth.sh` with the panel open: VS Code's main process flat, and
  request volume nil rather than 260/s.
- Drag a node and see the block file change on disk and the drawing follow it. Blocks are
  still files, so `xxd` on a block file is still the ground truth.

## Risks

**This reverses a pinned invariant.** CLAUDE.md's bridge surface says Go→TS is the block
files "and NOTHING ELSE — no frames, no pipes, no host→webview message at all (trace events
were the last to stream)". This change adds a host→webview message. The invariant has to be
rewritten in the same commit, or the guard and the prose disagree.

**It reintroduces the failure mode that deleted the frames.** From `f0b0c5cef`: "an fd/row
mismatch could silently undraw an entire entity class". A push path can deliver the right
bytes to the wrong row and draw something plausible and wrong, where a fetch of a wrong path
is a visible 404. Mitigation, and it must be in the first commit rather than added later:
**every message carries the block's own path**, and the receiver rejects bytes whose path it
did not ask for. A wrong address then fails loudly instead of drawing.

**The host becomes a reader of files it does not own.** Persistence doctrine is "one owner
writes inside `nodes/<id>/`: that node". The host would only READ, which breaks no writer
rule, but `check-persist-write-ownership.sh` should be read before assuming that is fine.

**Cadence moves into the host.** Today each reader picks its own; after this, the host's loop
decides, and a slow host stalls every block at once rather than one reader. The pointer
target needs frame cadence and everything else does not — that distinction has to survive the
move, or the editor feels laggy and the traffic comes back.

**It may not fix it.** The loader is one consumer of whatever is actually wrong. If the
panel is being starved by something broader, removing our fetches removes the symptom we can
see and the editor works — which is the point — but the cause stays unknown. Say so plainly
rather than declaring the bug solved.

## What is deliberately not in scope

- Changing what a block contains, or how Go writes one.
- Trace files. `scripts/readtrace` reads them from disk and is not part of the render path.
- The TS→Go direction.
- Any attempt to fix VS Code. If the probe extension reproduces the hang on a clean profile,
  that is a bug report with a repro, filed separately.
