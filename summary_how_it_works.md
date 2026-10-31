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
