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
| start_jφ, start_jθ | a multiple of s; the panel takes the multiple, so 12 on the panel is a full turn whatever s is. r is not an angle and is not scaled |
| t_jφ, t_jθ | ∈ {0, …, s − 1}, the ticks past start_j; the angle is start_j + t_j ticks. The panel shows `start +t` with ▲▼ arrows that step one tick, carrying into start at s. Changing s sets every t to 0. Each is its own state file, `start-<j>-phi-tick`, `start-<j>-theta-tick` |
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

Rounds run while the speed slider is above 0. Each node waits one clock cycle before its first
round, and that cycle reads the saved speed (sent before the nodes start), so a scene loaded at
speed 0 runs no round until the slider or step says so. The step button beside it runs one round on
every node, so a paused card can be walked a round at a time. The reset button beside it puts
every node back at round 0, so the next round sends `start_j` again; it redraws the arrows and
moves no node. The start button beside that one places each node's k = 1 partners where its
vectors say (Local to global).

## dir_down(local_arrival, p, qt)

    pole       = p + pole_offset
    direction  = −m  if local_arrival ∈ (pole, qt), else 0      (φ and θ each)

## dir_up(local_arrival, qt, p)

    pole       = p − pole_offset
    direction  =  m  if local_arrival ∈ (qt, pole), else 0      (φ and θ each)

## pick_one(k₁, k₂, arrival₁, arrival₂)

    arrival = k₁·arrival₁ + k₂·arrival₂   if k₁ ⊕ k₂, else 0

## One node

Everything in a round is φ and θ only, as on the card: `local_arrival_j`, what crosses a link,
pick_one and the directions. r is never in a round; it is read from `start_j` only where the
vector is converted (Local to global).

1. `local_arrival_j ← link j` for each partner j. The first round sends `[k_j · start_jφ, k_j · start_jθ]` on
   link j; every later round sends `[k_j · local_arrival_j]`, so a link with `k_j = 0` always
   carries 0. A link with `L_j = 0` arrives as 0.
2. `local_arrival₁ = local_arrival₂ = pick_one(k₁, k₂, local_arrival₁, local_arrival₂)`. Each
   k here is the k of the path the arrival came along: the sender's `k_j` for that link, sent
   with the angles. k = 1 marks the path's direction, so a node keeps what its predecessor on
   the path sent it, not what comes back the other way.
3. For each link, with that link's `pole_offset_j`:
   `local_arrival_j = local_arrival_j + dir_down(·, 0, 3s) + dir_up(·, 3s, 6s) + dir_down(·, 6s, 9s) + dir_up(·, 9s, 12s)`.
   The directions step the arrival.
4. `[k_j · local_arrival_j] → link j`.

The exchange is six unbuffered channels, one per ordered pair. A round is four rendezvous
(send to both partners, receive from both), and the round's select stays willing to perform
any one still outstanding, so a cycle of three cannot deadlock.

## Local to global

The card stops at its local numbers; showing them is conversion. For each pair separately: when
node a's `k_b = 1`, node b is where the a→b vector says — a's centre plus that vector.

1. **The a→b vector.** φ and θ count spokes, each `T/12s` index steps, and r is `start_b`'s r in
   node radii. Its angles are `start_b + t_b` before the first round and after a reset, then after
   each round the `k_b · local_arrival_b` a sent, unless that is 0 — a 0 is no change, so the
   vector keeps the last one that was not 0. It is drawn as the arrow from a.
2. **When b is placed.** a's geometry goroutine sends b the tip `C_a + polar_to_cart(a→b)` at
   load, on the start button, when a's start, t, k or s is edited, and after a round that changed
   the vector. Reset redraws the arrow and moves nothing. The tip goes on the tip channel for the
   pair a → b — one buffered channel per ordered pair, made by this kind.
3. **One hop.** b goes to the tip and passes nothing on: each pair is placed on its own, so there is
   no chain and no loop to stop. If a moves later, b follows only when a next sends.
4. **Saved.** b's new place is committed like a pointer drag, so its index (read from O, rounded)
   is saved and a reload starts from it.

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
