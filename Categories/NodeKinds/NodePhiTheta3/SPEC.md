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

The node holds the current angles in ticks, `cur`, which begin at `start · s`.

## Move — start toward end

On every round (speed above 0, or the step button), φ and θ each:

    gap_cw  = (end + 12s − cur) mod 12s
    gap_ccw = (cur + 12s − end) mod 12s
    step    = min(m·s, cw·gap_cw + ccw·gap_ccw)
    cur     = (cur + cw·step + ccw·(12s − step)) mod 12s      if cw ⊕ ccw, else unchanged

Reset puts `cur` back at `start · s`. An edit to start does the same.

## Local to global — shown in the visual

Nothing crosses between nodes, and no node sets another node's position. The global cards only show
each node's own local data.

1. **The node** is drawn at its own centre from its scene index:
   `C_a = O + polar_to_cart(I_r·c_r, I_φ·c_φ, I_θ·c_θ)` (the scene card).
2. **Each a→b vector** (the each-pair card), for each b with a's `L_b = 1`, from a's own start:
   `(R, φ, θ) = (start_r·ρ, cur_φ·2π/12s, cur_θ·2π/12s)`, ρ the node radius,
   `d = polar_to_cart(R, φ, θ)`, drawn as an arrow from `C_a` to `C_a + d`, and b goes to its tip:
   a copy of b is drawn at `C_b = C_a + d`. Each pair on its own — b linked by two nodes is drawn at
   both tips, and b's own centre stays at its index.

a redraws its vectors when its start moves, an edit changes its start, links or s, or the start button.

## Description

One of three φ, θ nodes running the three-node cards: moves its start toward its end, and draws its
a→b vector from its own centre for each node it links to.

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
