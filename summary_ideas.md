# Summary ideas

What the current design does that is correct — not how it does it, but what it does.

- [Vision](#vision)
- [In one paragraph](#in-one-paragraph)
- [Design principles](#design-principles)
- [What it does](#what-it-does)
- [Roadmap](#roadmap)

Also:

- [summary_how_it_works.md](summary_how_it_works.md): the same properties, by mechanism,
  and the code to read for each.
- [summary_writing.md](summary_writing.md): how this summary is written, and how to express
  a vision.
- [summary_open_questions.md](summary_open_questions.md): every question still to be
  answered, with an example of how each answer is expressed.

---

## Vision

### Purpose

- **Who:** They are a visual thinker who codes by directing an AI.
- **What they found:** For what they wanted, basic visuals came quickly. They typed the
  change they wanted to see, the AI made it, and the code worked, though they had not read
  it and did not know the language.
- **The problem:** Beyond basic visuals, they were not satisfied by what the AI made with its
  defaults.
- **What they made:** They directed the AI away from those defaults to make their own 3D
  simulation.
- **What it is for:** It lets them make 3D simulations their way.

### The priority

The code the AI generated was a reading bottleneck, so they changed the priority: they check
what the AI builds by pressing buttons and watching the screen first, and read its code
second. So the screen draws only what the simulation holds.

---

## In one paragraph

There is a 3D simulation system. It does 3D rotation by adding and subtracting integers to
each link's spherical-coordinate angles, φ and θ, centred on the node the link starts from.
Any number of centres can rotate at the same time, each by its own rotation. Each link also
has a length, r, changed the same way, which sets how far the link reaches and never where
it points. It has a UI for setting angles, lengths and many other settings for 3D rotations.

### The tools

It is built only from tools in everyday use: spherical coordinates, whole-number counting,
addition, sin and cos and their inverses, subtracting one point from another, nodes
passing messages, and a clock.

---

## Design principles

An AI's defaults keep the truth in one global x, y, z frame. Their way keeps it in each
link's own integer counts of φ, θ and r, so a rotation is an exact count of steps that never
drifts. x, y, z are only worked out from those counts, to place and draw the nodes.

---

## What it does

### The parts

It is a network of nodes joined by links. Each link goes from one node to another and
holds a direction, φ and θ, and a length, r.

### Steps

A link is changed in steps: integers added to and subtracted from its φ, θ and r. Steps in
φ and θ rotate it; steps in r change its length. Stepping back the same number of times
returns a link to exactly where it was, however many steps are made.

### Two angles

A rotation is steps added to φ and steps added to θ, and stepping one never changes the
other. Because the steps are added, they can be made in any order and the link ends up
pointing the same way. Stepping r never changes where it points.

### Moving nodes

A link moves only the node at its other end. Rotating it swings that node about the node
the link starts from; stepping r moves that node nearer or farther along the link. Nodes
linked from that node are not carried along.

### The network

A rotation's centre is the start node of the link being rotated. Selecting links picks the
centres: select links that start from different nodes, and each of those nodes is a
centre. One step changes every selected link at once, each by its own step size, and every
node those links move moves in that one step. In one step, each node is moved by at most
one link, so a second selected link into the same node is refused, as is a length below
zero.

### Sequences

A scene holds several sequences, each a separate saved setup of nodes and links.

### The screen

What is on the screen is always what the simulation holds, and a scene reopens where it
was left.

---

## Roadmap

| | Item |
|---|---|
| **Now** | Make the UI easier to use for all the settings (in progress). |

After that, selection, the camera and the gestures are moved onto the same integer counts
([summary_open_questions.md](summary_open_questions.md)).
