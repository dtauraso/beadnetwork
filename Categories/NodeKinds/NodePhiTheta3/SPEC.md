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
moves no node. The start button beside that one turns each node's k = 1 partners by
`start_j + t_j` about the selected node (Local to global).

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

The card stops at its local numbers; showing them is conversion, and the conversion is the rotation
card (`docs/pair-node/math/framework/sphere-card-3-rotation.js`): a card change turns a node about
the selected node C. Each node starts at its saved index.

1. **Centre.** C is the centre of the node selected in the editor. When the selection changes, the
   dispatcher tells every card node (`CardPanel.Inboxes.SetCentre`); the selected node's geometry
   goroutine then sends its centre to the other two, and again whenever it moves.
2. **What turns, by how much.** When node n's link j has `k_j = 1` and changes, n turns node j by
   that change, in ticks turned to radians as spokes (`T/12s` index steps each): `start_j + t_j` on
   the start button, the difference when start or t is edited, and each round's direction sum
   (`local_arrival_j` after the steps minus the picked arrival). A 0 change sends nothing, so it
   moves nothing. n's geometry goroutine sends the turn to j's on the turn channel for the pair
   n → j — one buffered channel per ordered pair, made by this kind.
3. **The turn, in j's own geometry goroutine.** j measures its local vector about C once,
   `(r, φ, θ) = cart_to_polar(P − C)`, when C is set, and keeps it; a pointer drag of j, or C
   moving, measures it again at the next turn. A turn adds `a_φ` to φ and `a_θ` to θ — two separate
   rotations, so their order does not matter — and j goes to `C + polar_to_cart(r, φ, θ)`.
4. **Nothing is passed on.** j moves only itself; no other node is told, so there is no chain and
   nothing to loop. A turn with no node selected, or about j itself, is dropped with a `turn`
   breadcrumb.
5. **Saved.** j's new place is committed like a pointer drag, so its index (read from O, rounded)
   is saved and a reload shows it there.

The arrows drawn from each node are its `start_j` vectors (`k_j = 0` draws nothing), length
`start_j`'s r in node radii; they show the card's values and do not place any node.

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
