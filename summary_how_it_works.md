# How it works

The properties in [summary_ideas.md](summary_ideas.md), described by mechanism.

## The parts

The software is a 3D simulation system. It is a network of nodes joined by links. Each
link is held by the node it goes into, and holds its own polar start (φ, θ, r) and an
integer Δ on top of it, one for each of φ, θ and r.

## Steps

Angles are counted in ticks: a whole turn is 12 × s ticks, where s is the node's own
scale. Integers are the main control numbers, so a rotation is a count of steps rather
than an arbitrary decimal value. Each press moves Δ by the link's whole-number increment.

## Two angles

A link's direction is worked out from its start plus its Δφ and Δθ, each added on its
own, and turned into x, y, z once, from the totals. No step is applied on top of another
step, so the order the steps were made in never reaches the result.

## Moving nodes

Each node keeps a global place, which is polar too (r, φ, θ around the scene's centre),
held as floats and saved to a file. A local-to-global adaptor turns places and links into
x, y, z. A stepped link places its node: the node goes to the other node's place plus the
link. The adaptor was picked so the local system can be seen, and tested, on screen.

Going the other way, global places are turned back into local links:

- When a node moves, its links that did not move it are remeasured from the two nodes'
  places, and its neighbours remeasure the links from it into them.
- Only a link's start is rebuilt, as floats, because the start is not part of the control
  system; its Δ stays a whole number of steps.
- While the selected links are stepped together, they are kept, not remeasured.
- When one link is edited on its own, that link places its node, and every other link at
  the node is remeasured, selected or not.

## The network

Each node runs on its own goroutine, with its own inbox.

1. A step of all the selected links starts when the UI sends it to every node, numbered
   as a round.
2. A node whose selected link was stepped waits for the place of the node that link
   starts from, then places itself from that place plus the link.
3. Each node sends its new place to its neighbours, and shows the step only once it has
   heard that round from all of them.

Each node checks its own links' changes, and refuses a length below zero or a second
selected link into itself.

## Time

Every node counts ticks from one shared clock, each with its own copy. The tick length is
set from the UI, and a step is shown a couple of ticks after it is sent.

## Sequences

Each sequence is its own folder of nodes and links, one value per file.

## The screen

The simulation writes each place, each link and everything drawn to files, one writer per
file. The screen reads those files and draws them; it works out no positions of its own.
A UI sits on top of this for setting the angles and the many other settings that control
the 3D rotations.

## The rules, to check without the screen

These rules are enough to recompute every position from the saved values, with no code
from this repo. `go run ./scripts/readscene -json <scene dir>` prints the saved values:
the frame (ρ, constant-r, max-index-φ, max-index-θ), each node's s and place, and each
link's start, Δ, increment and whether it is selected. `go run ./scripts/readscene -steps
<scene dir>` presses each of the eight step buttons once on every sequence, in memory, with
the same link code the editor runs, and prints each before and after in that form, so a
step can be checked without the editor. Nothing on disk changes.

For a link, s is the s of the node it goes to, and one turn is T = 12·s ticks.

1. **Where a link points.** φ = start φ + Δφ and θ = start θ + Δθ, each wrapped into
   0 ≤ angle < T; r = start r + Δr.
2. **A step.** It adds the increment, plus or minus, to one Δ (φ, θ or r), or with the
   "both" button to all three, and changes nothing else in that link. An angle's Δ is wrapped into 0 ≤ Δ < T; Δr is not wrapped. A step that would
   take r below 0 is refused.
3. **A link's vector.** With a(t) = t · 2π / T:
   x = r·ρ·sin a(φ)·cos a(θ), y = r·ρ·cos a(φ), z = r·ρ·sin a(φ)·sin a(θ).
   One case differs: when φ is past half a turn (φ > T/2) and φ = θ, z is negated. There
   both angles count as negative, and without the flip sin φ · sin θ would be a square
   that never lets the link reach the negative-z side.
4. **A node's position.** The same formula with r · constant-r for the length,
   φ · 2π / max-index-φ and θ · 2π / max-index-θ for the angles, and no z flip.
5. **Every link closes.** For every link a→b: position(a) + vector(a→b) = position(b), to
   floating-point rounding.
6. **What a step of the selected links changes.** Each Δ of a stepped link that changes
   changes by exactly its increment (wrapped, as in rule 2) and the link's start stays the same. No other link's Δ changes. A link that was
   not stepped but touches a node that moved gets a new start, the one that makes rule 5
   hold again for it.

Rules 1–5 are checked on one state; rule 6 is checked between a before and an after.
