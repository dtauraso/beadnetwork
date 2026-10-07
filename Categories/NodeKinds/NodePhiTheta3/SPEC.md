# NodePhiTheta3

The pair φ, θ 3 tab's node, running the three-node phi theta card
(`docs/pair-node/math/framework/sphere-card-3.js`). The card is the source; this file and
`node.go` follow it. The card is local and polar-only: every number below is in ticks of
`s`, and nothing on it names a global centre.

## Shared — the three nodes together

| Name | Value |
|------|-------|
| s | the tick mark scale; on the card s = 1, in Go s is set from the s panel (default 30). It is the nodes' only: 12s must divide the scene's turn so a tick is a whole number of index steps, and the panel refuses any other s |
| 12s | 1 full turn φ, 1 full turn θ |
| 3s | 1 quarter turn φ, 1 quarter turn θ |
| P | the pole numbers, {0, 6s, 12s} |
| qt | the quarter turn mark between two pole numbers, qt ∈ {3s, 9s} |
| m | the jump constant, m ∈ ℤ⁺, m = 1 by default (set from the m panel) |
| L_j | 1 if a link reaches node j, 0 if not |
| pole_offset_jφ, pole_offset_jθ | ∈ {0, 1s, 2s, 3s}; the panel takes the multiple of s (0, 1, 2 or 3) |
| p | ∈ P |
| k_j | ∈ {0, 1} |
| ← | the node receives the vector on a link |
| → | the node sends the vector on a link |

## Each node's vectors

Node n holds, for each of its two partners j: `start_j = [φ, θ, r]`,
`pole_offset_j = [φ, θ]` and `k_j`; and `L = [L₁, L₂, L₃]`. Each one is an editable panel
in the tab, and each value is its own file: the default under the tracked
`nodes/<n>/data/state/` (`start-<j>-phi`, `pole-offset-<j>-theta`, `k-<j>`, `l-<j>`, …), and a
typed edit under the gitignored `nodes/<n>/drag/state/`, which replaces the default at load.
`s` and `m` are scene-wide and live in the gitignored `view/card-s.bin` and `view/card-m.bin`.

Rounds run while the speed slider is above 0. The step button beside it runs one round on
every node, so a paused card can be walked a round at a time. The reset button beside it puts
every node back at round 0, so the next round sends `start_j` again.

## dir_down(local_arrival, p, qt)

    pole       = p + pole_offset
    direction  = −m  if local_arrival ∈ (pole, qt), else 0      (φ and θ each)

## dir_up(local_arrival, qt, p)

    pole       = p − pole_offset
    direction  =  m  if local_arrival ∈ (qt, pole), else 0      (φ and θ each)

## pick_one(k₁, k₂, arrival₁, arrival₂)

    arrival = k₁·arrival₁ + k₂·arrival₂   if k₁ ⊕ k₂, else 0

## One node

1. `local_arrival_j ← link j` for each partner j. The first round sends `start_j` on link j;
   every later round sends `[k_j · local_arrival_j]`. A link with `L_j = 0` arrives as 0.
2. `local_arrival₁ = local_arrival₂ = pick_one(k₁, k₂, local_arrival₁, local_arrival₂)`.
3. For each link, with that link's `pole_offset_j`:
   `local_arrival_j = dir_down(·, 0, 3s) + dir_up(·, 3s, 6s) + dir_down(·, 6s, 9s) + dir_up(·, 9s, 12s)`.
   r passes through unchanged.
4. `[k_j · local_arrival_j] → link j`.

The exchange is six unbuffered channels, one per ordered pair. A round is four rendezvous
(send to both partners, receive from both), and the round's select stays willing to perform
any one still outstanding, so a cycle of three cannot deadlock.

## Local to global

The card stops at its local numbers; placing them is conversion. Each node's `start_j` is a
vector from its centre: φ and θ count spokes, each `T/12s` index steps (one spoke is 1/(12s) of
a turn, T is the whole turn in index steps, and T/12s is a whole number), and r is a length in
node radii, `r · node radius`, not rounded. It is drawn as an arrow from the node's centre to its
exact tip, so r = 1 reaches the ring.

Chaining: when node n's `k_j = 1`, node j starts at the exact tip of n's `start_j` — n's centre
plus that vector. n's geometry goroutine sends that tip to j's geometry goroutine on the
placement channel for the pair n → j — one buffered channel per ordered pair, made by this kind
beside the card's link channels, so every placement is delivered and none passes through the
dispatcher or a pointer-drag slot. j's centre is the exact tip, and j's own arrows start there, so
a chain's shape is its local vectors alone. The tip measured against the scene centre and rounded
to the scene's index is kept beside it for what counts in index steps — the vectors to edge
neighbours and the move they are told — and is never read back into the chain. n sends when n
starts, whenever n's start, k or s changes, and whenever n itself moves — so a move passes down a
chain of k links, node by node. When `k_j` goes back to 0, n sends j a release; once no partner
places j, j's centre is its index again. A pointer drag of j also makes its index its centre.

j keeps the latest tip from each partner sending to it. With one sender, j goes to that tip. With
two, j's own k picks, as pick_one does: j takes the tip from the partner its single `k = 1` names,
and stays put if its k names neither or both.

A placement moves j's centre only. j's own `start` vectors keep their own angles in the same axes
every node uses.

A card-placed position is not saved: it is the tip of the sender's `start_j`, so it is placed
again from the card when the scene loads. Only the card's values and pointer drags are saved.

The card checks k only within a node (k₁ ⊕ k₂), so the k links can loop across nodes (1 leads 2,
2 leads 1). Each move carries the nodes it has passed through, and a node never sends it on to
one already on that path, so a loop moves each node once and stops instead of running away.

## Description

One of three φ, θ nodes running the three-node card: receives a local arrival on each of its
two links, picks one with k, steps it with dir_down/dir_up, and sends it on.

## View

| Field | Value |
|-------|-------|
| kindId | 13 |
| kind | nodePhiTheta3 |
| bg | #e3f2fd |
| border | #1565c0 |
| text | #0d2b4a |
| minWidth | 70 |
| shape | rect |
| fill | #e3f2fd |
| stroke | #1565c0 |
| width | 70 |
| height | 60 |

## Ports

None. A port is where a bead line attaches, and nothing is placed on these edges: what
crosses is the card's vector, on the channel the edge allocates.

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|

## Runtime status

- Loader-registered: yes
- TSX render: present
