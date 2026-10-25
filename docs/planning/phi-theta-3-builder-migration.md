---
branch: task/phi-theta-3-builder-migration
---

# The a→b builder in the pair φ, θ 3 tab

## Target

The tab runs what `docs/pair-node/phi-theta-3/builder/` runs, as a network: each node is its
own goroutine holding its own float place (`base/at-*.bin`) and its own out-links, and a
change to one node reaches another only as a message between them.

It reverses one invariant of `Categories/NodeKinds/NodePhiTheta3/SPEC.md`: "nothing crosses
between nodes, and no node sets another node's position". A picked link's source now SENDS
its vector and the target places ITSELF from it. No node writes another node's place.

## The model

| builder | the tab |
|---|---|
| frame `ρ, T_φ, T_θ, c_r`, O | `constants/rho.bin`, `max-index-phi/theta.bin`, `constant-r.bin`; O is the scene centre |
| `s` | `view/card-s.bin`, a node's own `base/s.bin` over it |
| node `{name, s, at}` | node `n`: `base/at-{r,phi,theta}.bin`, `base/s.bin` |
| edge `{from, to, start, delta, increment, lock, locking}` | `nodes/<from>/links/<from>To<to>/`, owned and written by `from` only |
| `picked` | the link's `picked.bin`; the TARGET grants a pick (at most one picked link into a node) |
| `send` / `take` | `Put{P, D}` from source to target; the target sits at `gridAt(P + D)` |
| `moved` → `remeasure` | `Moved{P}` to every node; each remeasures ITS OWN links touching the mover |
| `moveAll` chains | roots send `Put` down picked links; each target forwards after it sits |
| `moveAll` rings | a `Trail` goes round the picked links; the node it returns to has the whole ring and places only itself |
| sequences | `sequences/<k>/nodes/…`, the open one named by `view/sequence.bin` |

The pure state machine is `Categories/Node/Links` (`Peer.Handle(msg) → messages`). The kind
wraps it in its goroutine and channels; nothing else steps it.

## UI

Each piece is laid out and hit-tested in Go and drawn from one block file by one TS drawer
(`Categories/Chrome/Panels/Sheet`): rows of text, fields with ± arrows, and buttons.

| builder | the tab | side |
|---|---|---|
| sequence picker, add / delete sequence | **Sequences** pill dropdown | left |
| tree page | **Tree** pill | left |
| node to place, method cards, ◀ step ▶ | **Place** pill | left |
| setup `ρ, T_φ, T_θ, c_r, s` | **Frame** pill | right |
| nodes table: `s, I_r, I_φ, I_θ`, add / remove | **Node places** pill | right |
| edges: start, Δ ±, increment, pick, locks, all | **Links** pill | right |

The left pills stack below the left panels; the right ones join the existing pill stack.

## Order — one commit each

1. Links: source-owned links, the frame, `Put`/`Moved`; a link's start edit and its Δ ± move its
   target. The Links pill (start, Δ ±). The old card (start/end/L, m/cw/ccw, the move toward
   end) goes, since links replace it.
2. Remeasure: a node that moves makes every link touching it remeasure, and the Links pill shows it.
3. Picks: the pick edit, granted by the target; pick / unpick, increments and locks on the Links pill.
4. The all row: every picked link steps; chains cascade from roots; a ring places itself.
5. The Frame and Node places pills; adding and removing nodes and links.
6. Import `sequences.js` into `sequences/<k>/` and load the open one.
7. The left pill stack and the Sequences pill dropdown: open, add, delete.
8. The Place pill with the two method cards.
9. The Tree pill.

## Verification

- A probe (`.probe/linkcheck`, not committed) drives `Links.Peer`s with the edit list of
  `docs/pair-node/phi-theta-3/tests/tests.js`, delivering messages in shuffled order, and
  compares node places with `tests-want.js` (`T_WANT`, `T_DIRS`) to 1e-9.
- Rings and chains: the same probe against the builder's `moveAll`, run in node, on the four
  saved sequences.
- `scripts/stop-checks.sh` empty after every commit; then the editor.

## Risks

- Messages arrive in any order. A node remeasures on every `Moved` it hears, so the last
  remeasure uses both ends' final places; the probe's shuffled delivery is what checks it.
- An all-step's `Put` can reach a node before that node has stepped its own Δ; each all-step
  carries a round number and a node holds messages from a round it has not reached.
