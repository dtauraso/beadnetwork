# Node polar rules panel

A left panel shows, per node, the polar constraint rules that decide where it may sit
relative to its holders — and lets the user edit them. Rule edits are **tracked authored
content** in `topology/nodes/<id>/meta.json`, deliberately unlike drag output
(gitignored `position.json` / `*.geom.json`).

## The rule that exists

`polar.OrbitRule` (`nodes/Wiring/geom/polar/orbit_angles.go:5`) is the only positional
constraint. `TrimToOrbitRule` (`nodes/Wiring/nodedrag/node_drag.go:53`) applies it in the
node's OWN goroutine, to the delta from each **in-neighbour**:

- `r` held unconditionally whenever a rule exists (`out.R = have.R`) — no field disables it,
- `φ` frozen iff `Phi != nil` (the stored value is decorative; only nil-ness is read),
- `θ` clamped to `[−MaxTheta, +MaxTheta]`.

The deleted rule-builder's grammar (`(center→a, θ) = (center→b, −φ)`, `∈ ◯ torus`,
commit `9e247aea1`) described an equality between two nodes' components. Nothing of that
shape survives, so the panel does NOT resurrect it.

## Syntax

One block per node, one sub-block per in-neighbour, one line per component; every component
prints every time, so "unconstrained" is a word rather than an absent line.

```
node 3 · PulseLeft
  orbits 2          r   fixed 54.3
                    φ   locked
                    θ   ∈ [−90°, +90°]
```

`φ` toggles, `θ` is typed in degrees with ENTER to commit, `r` is read-only because it is
not separately representable. Degrees on screen, radians on disk.

## Ripple

1. **Buffer** (`Buffer/bufschema/`, `BufLayoutVersion` 46→47)
   - `layout_node.go`: `OrbitRLocked u8`, `OrbitPhiLocked u8`, `OrbitThetaMax f32` (<0 = free).
   - `layout_edge.go`: `DstNodeRow i32` — the panel derives in-neighbours from edges, and
     `r` from the two node centers. Without it a target row is not recoverable from a frame.
   - `layout_panel.go`: `NodeRules u8`.
2. **Node frame** — `nodeframe.NodeFrameInput` carries the three orbit fields;
   `writeStreamFrame` fills them from `m.OrbitRule()`. Edge frame fills `DstNodeRow`.
3. **Panel flag** — `viewstate`: `PanelState.NodeRulesOpen`, `PanelToggles["nodeRules"]`,
   `ViewPanelFlags.NodeRules`; persists to `topology/view/panels.json` like its siblings.
4. **TS → Go edit** — a new addressed entity kind `node` on the existing `update` op (NOT a
   new op): attrs `orbitPhi` (toggle) and `orbitMaxTheta` (`X` = degrees, `Num` = node row).
   Parity: `input_codec.go` fingerprint, `input-layout-gen.ts`, `messages.ts`,
   `handle-message.ts`, `dispatch_edit.go`'s `EDIT_UPDATE_KINDS` fence.
5. **Ownership** — the stdin goroutine must NOT touch node state. The handler posts a
   `movemsg` orbit-rule message to the node's inbox (`MoveDispatch.Inboxes`); the node's own
   goroutine applies it, persists, and re-emits its frame.
6. **Persistence** — a meta.json writer that read-modify-writes only the `orbit` key.
   `nodefiles.WriteNewNodeFiles` currently rewrites meta.json as `{id,type}` and would
   clobber a rule; it must preserve `orbit`.
7. **Generated** — `go run ./tools/gen-node-defs` after 1 and 4, or the kind does not exist
   in the binary (`check-generated.sh`).
8. **Webview** — `NodeRulesPanel.tsx`, a read-only buffer decode via `useSyncExternalStore`
   (no store, no TS domain state — `check-no-webview-state.sh`), fire-and-forget writes
   (`check-no-await-on-bridge.sh`), mount + CSS, and a toggle row alongside the existing
   node panel rows.

## Verification

`bash scripts/stop-checks.sh` clean (EMPTY stdout), then drive the editor: open the panel,
confirm node 3 reads `φ locked` / `θ ∈ [−90°, +90°]` matching its meta.json, toggle φ and
retype θ, confirm meta.json changed on disk and a drag now obeys the new bound.

Delete this doc when the change lands.
