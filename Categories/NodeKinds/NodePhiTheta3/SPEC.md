# NodePhiTheta3

The pair φ, θ 3 tab's node, running the three-node phi theta cards
(`docs/pair-node/phi-theta-3/`): the local card (`card-3-2/sphere-card-3-2.js`) for what a node holds
and how its start moves, and the each-pair card (`sphere-card-3-rotation.js`) for where a node is put,
in the a→b card's order — local first, global last (`atob/`). The cards are the source; this file and
the Go follow them. Every angle below is in ticks of `s`.

## Shared — the three nodes together

| Name | Value |
|------|-------|
| s | the tick mark scale; s = 30 by default, set from the panel. 12s must divide the scene's turn, and the panel refuses any other s |
| 12s | 1 full turn φ, 1 full turn θ |
| m | the jump constant, m ≥ 1, m = 1 by default |
| cw, ccw | ∈ {0, 1}; cw = 1, ccw = 0 by default. A move needs exactly one of them |
| L_j | 1 if the node has an outgoing link to node j, 0 if not |
| links | [L_j, L_k], j, k the other two nodes |
| start | [start_φ, start_θ, start_r]; start_φ, start_θ ∈ {0, …, 11}, a multiple of s on the panel; start_r in node radii |
| end | [end_φ, end_θ] ∈ {0, …, 11}, a multiple of s on the panel |

s, m, cw and ccw are scene-wide (`view/card-s.bin`, `card-m.bin`, `card-cw.bin`, `card-ccw.bin`).
Each node's start, end and links are its own state files: the default under the tracked
`nodes/<n>/data/state/` (`start-phi`, `end-theta`, `l-<j>`, …), and a typed edit under the gitignored
`nodes/<n>/drag/state/`, which replaces the default at load.

## A node's start is its own vector

A node's start is the vector from the node that links to it — its centre of rotation — to itself.
The node holds the current angles in ticks, `cur`, which begin at `start · s`.

## Move — start toward end

On every round (speed above 0, or the step button), φ and θ each:

    gap_cw  = (end + 12s − cur) mod 12s
    gap_ccw = (cur + 12s − end) mod 12s
    step    = min(m·s, cw·gap_cw + ccw·gap_ccw)
    cur     = (cur + cw·step + ccw·(12s − step)) mod 12s      if cw ⊕ ccw, else unchanged

Reset puts `cur` back at `start · s`. An edit to start does the same.

## What a node tells its partners

Whenever its links or its vector change, a node tells both partners `[links, start_r, cur]` on a
one-slot, latest-wins channel per ordered pair, and wakes them. Nothing else crosses between nodes.

## Local to global — where a node is put

A node h with `L_j = 1` is a centre of rotation for node j, and places it.

1. **Local.** j's vector, from what j told h: `d = polar_to_cart(start_r·ρ, φ·2π/12s, θ·2π/12s)`,
   ρ the node radius.
2. **One centre** (only h links to j). **Global, last:** `C_j = C_h + d`.
3. **Two centres** (h and g both link to j). j turns about both: its point is on the sphere shared by
   h's and g's spheres, each of radius `R = start_r·ρ`. With `D = C_g − C_h`:
   the shared centre is `C_h + D/2`, its radius `√(R² − |D|²/4)`, and
   **global, last:** `C_j = C_h + D/2 + polar_to_cart(√(R² − |D|²/4), φ, θ)`. Each holder works this
   from its own centre; when `|D| > 2R` the spheres do not meet and j is not moved.
4. The tip goes on the tip channel for the pair h → j, and h draws an arrow to each tip it sends.

A holder re-sends only when what it would send changes: a partner's vector or links, its own links,
s, or the start button. j does not pass its move on, and a holder moving does not move j until the
holder next sends.

## Description

One of three φ, θ nodes running the three-node cards: moves its start toward its end, tells its
partners, and places each node it links to around itself.

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
