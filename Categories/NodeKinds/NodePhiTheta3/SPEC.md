# NodePhiTheta3

The pair φ, θ 3 tab's node: each node holds its own float place and its own outgoing links, and a
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
| its links | `links/<me>To<j>/`, replaced as a whole by `drag/links/` once one is edited |

## A link a→b — owned by a

`start (φ, θ, r)`, `Δ (φ, θ)`, `increment (φ, θ)`, `lock (φ, θ) ∈ {−1, 0, 1}`, `locking`, `picked`.
Angles are in ticks of a's `s`. Its direction is `start + Δ`, wrapped to a turn, and its vector is
`d = polar_to_cart(r·ρ, (φ+Δφ)·2π/12s, (θ+Δθ)·2π/12s)`, drawn as an arrow from `C_a`.

## Messages

- **Put** (a → b): `C_a` and `d`. b sits at `GridAt(C_a + d)`.
- **Moved** (b → each node with a link into b): b's new place. Each remeasures its link to b
  (start = the measured direction − Δ), except the link that moved b, or every picked link during
  an all-step; b remeasures its own links itself. A node with no link into b is not told.
- **Connect** (UI → a, carrying b's place): a measures a new link a→b and sends b **Linked**, so b
  adds a to the nodes it tells when it moves, and answers with a Moved. **Drop** sends b
  **Unlinked**.
- **Pick** (UI → b, b → a): b grants a pick only if no other node has a picked link into b.
- **StepAll** (UI → every node, numbered): every node steps Δ on its picked links first — the
  locals change before any global point does. Then each node's point is recovered from its one
  picked link in (a node has at most one): the source sends `p_a` and its new `d'` down each picked
  link, and b moves halfway to `p_a + d'` and sends its own point on, until it moves less than
  1e-9. A node with no picked link into it is rotated by nothing, so it stays and sends its point as
  final; a target told a final point sits exactly at `p_a + d'` and sends its own as final. Where
  every node has a picked link in, every point is recovered at once and the centre holds. No picked
  link is remeasured, and no node looks for a loop.

## Time — machine ticks become one human tick

Nothing on this tab sleeps: not the node's goroutine, not its geometry goroutine (it writes the
block files once per human tick), and not the gesture actor while the tab is open (it reads
input every pass and refreshes the pills once per human tick); a node has no bead animation.

Each pass of the node's loop takes at most one message and counts one machine tick `k`. When `h = ⌊(now − epoch) / Δt⌋` changes, it applies, persists and publishes
what changed during that tick, then posts `(h, k)` to its timing slot. A StepAll carries the human
tick `H` it was sent on, and a node holds what the step changed until its point is final and
`h ≥ H + 2`, so every node shows a step's links and points on the same human tick — never new
angles drawn from a point not yet recovered. So a node shows at most one
step per human tick whatever its `n` (machine ticks per human tick); `n` differs between nodes and
between ticks and changes nothing the node computes. The Frame pill shows Δt, `h`, the slowest
node's `n` and which node it is, and how many nodes are behind `h − 1`; Node places shows each node's `n`.

## Description

A node of the pair φ, θ 3 tab: holds its own place and outgoing links, and places itself from a link's
vector when the link's source sends it.

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
