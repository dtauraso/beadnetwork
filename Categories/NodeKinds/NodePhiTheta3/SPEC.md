# NodePhiTheta3

The pair φ, θ 3 tab's node, running the three-node phi theta card
(`docs/pair-node/math/framework/sphere-card-3.js`). The card is the source; this file and
`node.go` follow it. The card is local and polar-only: every number below is in ticks of
`s`, and nothing on it names a global centre.

## Shared — the three nodes together

| Name | Value |
|------|-------|
| s | the tick mark scale; on the card s = 1, in Go s = the index steps per tick (default 30, set from the s panel) |
| 12s | 1 full turn φ, 1 full turn θ |
| 3s | 1 quarter turn φ, 1 quarter turn θ |
| P | the pole numbers, {0, 6s, 12s} |
| qt | the quarter turn mark between two pole numbers, qt ∈ {3s, 9s} |
| m | the jump constant, m ∈ ℤ⁺, m = 1 by default (set from the m panel) |
| L_j | 1 if a link reaches node j, 0 if not |
| pole_offset_jφ, pole_offset_jθ | ∈ {0, 1s, 2s, 3s} |
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
every node, so a paused card can be walked a round at a time.

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

## What the panels do on screen

- `s`: each node draws 12s tick lines from its centre to its ring. φ and θ in `start_j` count
  those spokes (one spoke is 1/(12s) of a turn).
- `start_j`: drawn as a light-violet arrow from node n's centre; r is in node radii, so r = 1
  reaches the ring.
- `k_j = 1`: node j is placed at the tip of n's `start_j`, and follows it when that tip moves,
  including when n is itself placed by another node. Setting it back to 0 leaves j where it is.
- When both partners have `k = 1` toward node j, j follows the one its own single `k = 1` names,
  and stays put if its k names neither or both.
- k links that loop (1 → 2 and 2 → 1) move each node once per change and stop.
- `m`, `pole_offset_j`, `L`: used in the card's rounds only; they do not move nodes.
- Every panel value is saved. A position a node was placed at by `k` is not saved; it is placed
  again from the card when the scene loads.

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
