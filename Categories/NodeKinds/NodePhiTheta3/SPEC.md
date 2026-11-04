# NodePhiTheta3

The pair φ, θ 3 tab's node: each node holds its own float place and the links into it, and a
change to one node reaches another only as a message between them. The node's logic is
`Categories/Node/Links`'s `Peer`; this kind is its goroutine.

## The frame — shared by every node

| Name | Value |
|------|-------|
| O | the scene centre |
| T_φ, T_θ | `constants/max-index-phi.bin`, `max-index-theta.bin`: index steps a turn |
| c_r | `constants/constant-r.bin`: the length of one radial index step |
| ρ | `constants/rho.bin`: the length of one unit of a link's `r` (the node radius when absent) |
| s | `view/s.bin`: ticks a twelfth of a turn, for a node with no `s` of its own |
| Δt | `view/tick-ms.bin`: ms one human tick lasts (default 16); a change reaches every node as `SetTickMs`, with no reload |

A place `I = (I_r, I_φ, I_θ)` is a float index: `C = O + polar_to_cart(I_r·c_r, I_φ·2π/T_φ, I_θ·2π/T_θ)`.

## A node — its own state

| Name | File |
|------|------|
| place `at` | `base/at-{r,phi,theta}.bin` (f64), replaced by `drag/at-*.bin` once the node has moved |
| `s` | `base/s.bin`, replaced by `drag/s.bin` |
| the links into it | `links/<j>To<me>/` (`source.bin` names `j`), replaced as a whole by `drag/links/` once one is edited |

## A link a→b — its angles stored at b

`start (φ, θ, r)`, `Δ (φ, θ, r)`, `increment n (φ, θ, r)` — what one press of Δφ, Δθ or Δr moves it by, whole ticks for an angle and a whole number above 0 for r —; Δ and increment are integers, start is measured and real —, `lock (φ, θ) ∈ {−1, 0, 1}`, `locking`, `picked`,
held by b, the target. Angles are in ticks of b's `s`; r + Δr is never below 0, and Δr has no lock and
is not stepped by StepAll.

Everything below counts **ticks**: a turn is `12s` ticks, a half turn `6s`, and one press moves
one tick. `1` and `−1` are one tick, not one `s`; at `s = 1` the two read the same.

An angle (start, and start + Δ) is stored and shown as a turn in `0..12s`: `12s` is the same
point as 0 and is written 0. There are no negative angles and no direction change: past 6s an
angle keeps going the same way round the circle (at `s = 1`, 6 + 1 = 7). A file holding an
older signed start (`−k`, `1 ≤ k ≤ 6s`) is read as `6s + k` (at `s = 1`, −1 → 7, −6 → 0). Δ is a
term: the count of ticks pressed, with the sign it was pressed.

A press is a rotation: one tick up the `0..12s` circle for +, one tick down for −, the same way
for every link, with no cases and no ends. The displayed angle is the sum `start + Δ` wrapped
into `0..12s` (at `s = 1`: 4 + −4 = 0, 4 + −5 = 11, 11 + 1 = 0). A press of increment n adds n
to Δ in one go. Start and Δ are never rewritten, and every press has an exact inverse. A start
± press is one tick the same way, and a typed start is wrapped into `0..12s`. A recovered start
is the measured angle turned back Δ ticks (`measured − Δ`, wrapped), so it moves exactly as far
as the node did and Δ is kept.

Only Δ is order-free: it is an integer count. Start and every place are real-valued — measured
through trig from cartesian points — so they record the order the nodes moved in; a start that
does not look right is set again (a typed start or start ±), and Δ is untouched by it. Its
local vector is
`d = polar_to_cart((r+Δr)·ρ, (φ+Δφ)·2π/12s, (θ+Δθ)·2π/12s)` with its z
negated when φ = θ and both are past 6s: with the two angles equal, sin φ · sin θ is a square
and would otherwise never let d reach the negative side. What is
drawn from `C_a` is the link's global version, `C_b − C_a`, from the two recovered points.

Each angle's arc is the path its angle parts walk: from 0 to start, then each Δ tick one at a
time around the circle. Every point on it is drawn by d's own code (polar_to_cart, the z case); the φ arc at θ = 0,
the θ arc at the link's φ. So an arc ends where d points. A tick across the arc marks the end of
the start part and the end of the Δ part, each only when that part covers some distance.

## Messages

- **Put** (a → b, down a picked link): `C_a`. b recovers its own point `C_a + d` from it.
- **Moved** (b → each node it is linked with): b's new place. A target holding a link from b
  remeasures it (start = the measured direction − Δ), except every picked link during an
  all-step; b remeasures the links into itself. A source only records the place, to draw from.
- **Connect** (UI → b, carrying a's place): b measures a new link a→b and sends a **Linked** with
  its place, so a draws it and tells b when it moves. **Drop** (UI → b) sends a **Unlinked**.
- **Edits** (start, press, increment, locking — UI → b): b changes the angles it holds and places
  itself at `GridAt(C_a + d)`.
- **S** (UI → b): b's new `s` rescales start and Δ of every link into b by `new s / old s`, so each
  keeps its direction. Δ must land on the new `12·s` tick line, whole, which holds exactly for an
  `s` that is a multiple of `s / gcd(s, every Δ)`; an `s` that would put any link's Δ between
  ticks is replaced by the closest such multiple (a tie takes the larger), and the reason is shown
  on the line below the node. So Δ stays whole and steps stay exact in any order.
- **Pick** (UI → b, b → a): b grants a pick only if no other link into b is picked, and tells a
  with **PickSet**, so a knows to send b its point.
- **StepAll** (UI → every node, numbered): every node first steps Δ on the one picked link it
  holds, before any global point changes. Recovery comes last: a node with no picked link into it is rotated by nothing, so it
  stays placed and sends its point down each picked link out of it; a target sent its source's
  point sits at `C_a + d'` and sends its own point on. A node sends its point only once it is
  placed, so every point sent is the one it keeps. No picked link is remeasured and no global
  measurement goes back into a picked link.

## Time — machine ticks become one human tick

Nothing on this tab sleeps: not the node's goroutine, not its geometry goroutine (it writes the
block files once per human tick), and not the gesture actor while the tab is open (it reads
input every pass and refreshes the pills once per human tick); a node has no bead animation.

Each pass of the node's loop takes at most one message and counts one machine tick `k`. When `h = ⌊(now − epoch) / Δt⌋` changes, it applies, persists and publishes
what changed during that tick, then posts `(h, k)` to its timing slot. A StepAll carries the human
tick `H` it was sent on, and a node holds what the step changed until it is placed, it
has a **Moved** for that step from every node it is linked with (a node nothing rotates sends
its unchanged place), and `h ≥ H + 2`, so every node shows a step's links and points on the
same human tick — never new angles drawn from a point not yet recovered, and never an arrow
drawn to where a neighbour was before the step. So a node shows at most one
step per human tick whatever its `n` (machine ticks per human tick); `n` differs between nodes and
between ticks and changes nothing the node computes. The Frame pill shows Δt, `h`, the slowest
node's `n` and which node it is, and how many nodes are behind `h − 1`; Node places shows each node's `n`.

## Description

A node of the pair φ, θ 3 tab: holds its own place and the links into it, and recovers its own point from
its picked link in when that link's source sends its point.

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

None. A port is where a bead line attaches, and no bead travels a link: what crosses is a
message on the links mesh.

| Name | Direction | EdgeKind | Notes |
|------|-----------|----------|-------|

## Runtime status

- Loader-registered: yes
- TSX render: present
