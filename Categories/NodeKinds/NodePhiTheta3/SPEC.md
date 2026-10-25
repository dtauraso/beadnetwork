# NodePhiTheta3

The pair φ, θ 3 tab's node, running the a→b builder (`docs/pair-node/phi-theta-3/builder/`) as a
network: each node holds its own float place and its own outgoing links, and a change to one node
reaches another only as a message between them. The builder is the source; this file and the Go
follow it. The node's logic is `Categories/Node/Links`'s `Peer`; this kind is its goroutine.

## The frame — shared by every node

| Name | Value |
|------|-------|
| O | the scene centre |
| T_φ, T_θ | `constants/max-index-phi.bin`, `max-index-theta.bin`: index steps a turn |
| c_r | `constants/constant-r.bin`: the length of one radial index step |
| ρ | `constants/rho.bin`: the length of one unit of a link's `r` (the node radius when absent) |
| s | `view/s.bin`: ticks a twelfth of a turn, for a node with no `s` of its own |

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

- **Put** (a → b): `C_a` and `d`. b sits at `gridAt(C_a + d)`.
- **Moved** (b → every node): b's new place. Each node remeasures its OWN links touching b
  (start = the measured direction − Δ), except the link that moved b, or every picked link during
  an all-step.
- **Pick** (UI → b, b → a): b grants a pick only if no other node has a picked link into b.
- **StepAll** (UI → every node, numbered): each node steps Δ on its picked links; a node with no
  picked link into it sends Put down its picked links and each target forwards after it sits; a
  node with one sends a **Trail** round the picked links, and the trail that comes back to it holds
  the whole ring, from which it places only itself (the builder's `moveAll` ring).

## Description

A node of the a→b builder: holds its own place and outgoing links, and places itself from a link's
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
