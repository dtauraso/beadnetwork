# Summary ideas

What the current design does that is correct — not so much how it does it, but what it does.
Every property below can be seen by running the editor.

- [Vision](#vision)
- [Design principles](#design-principles)
- [In one paragraph](#in-one-paragraph)
- [What it does](#what-it-does)
- [Roadmap](#roadmap)
- [Future directions](#future-directions)

Also:

- [summary_how_it_works.md](summary_how_it_works.md): the same properties, by mechanism.
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
second.

### What follows from the priority

- **The screen draws what the simulation wrote.** Where that is an angle (a node's pole, the
  camera), the screen works it out as x, y, z.

### Where it is going

Many links will change their angles at once, enough to make a bird that flaps its wings and
changes the way it faces while it flaps, checked by watching it. The way there is moving selection, the camera and the gestures
into the local polar system ([Design principles](#design-principles)); the steps are in the
[Roadmap](#roadmap), and [Future directions](#future-directions) says why they follow.

---

## Design principles

An AI's defaults keep the truth in one global x, y, z frame. Their way, in the
[Purpose](#purpose), keeps it in each link's local, whole-number Δ, because each link moves
in its own polar frame; so where a link points is an exact count of steps that never drifts.
Everything else is worked out from those counts through x, y, z, as floats: a node is
placed at its source node's x, y, z plus its link's vector, converted back into φ, θ and r, and
the start of every other link is measured from where its two nodes now are. Only Δ stays
exact. From that:

- **A link is a vector.** It has a direction, φ and θ, and a length, r, each a start plus a
  whole-number Δ.
- **Stepping a link moves the node at its end.** The links that move with it keep their
  vectors; every other link touching a moved node has its start recalculated from where its
  two nodes now are, and its Δ stays.
- **Each node moves on its own.** A node places itself from what its neighbours tell it, or
  from a place set on it directly in the UI.

It rules out:

- **Selecting nodes with one central raycast through global x, y, z.** It starts and ends
  in x, y, z and is never worked out from any node's own steps; each node measures the mouse
  in its own local steps instead
  ([question 3](summary_open_questions.md#3-why-does-node-selection-get-a-replacement)).
- **A camera made of decimal angle changes stacked on each other.** Its numbers are not
  whole-number steps and are not worked out from any; the camera becomes a node held in
  whole-number steps like everything else
  ([question 5](summary_open_questions.md#5-why-does-the-camera-get-a-replacement)).

| Principle | Meaning |
|---|---|
| Local data as truth | Each link's own Δ is the source of truth. |
| Polar as the local coordinate system | Links are φ, θ, r — not x, y, z. |
| Integers as the control numbers | Changes in the local system are whole-number counts of steps. |
| Global as the working step | Node places and link starts are worked out in x, y, z from the local data. |
| Local → global adaptor | Converts local places and links into x, y, z. |
| Global reconstructions → local | Starts and node places are rebuilt from global places, as floats; Δ is not. |

The local-to-global adaptor serves [the priority](#the-priority): it places the nodes in
space, so the local system can be seen on the screen, where it is checked.

---

## In one paragraph

There is a 3D simulation system. It moves links in 3D by adding local integer changes to
their angles. It has a UI for setting those angles and many other settings for how they change.

### The tools

It is built only from tools in everyday use: polar coordinates, whole-number counting,
addition, sin and cos and their inverses, subtracting one point from another, nodes
passing messages, and a clock.

---

## What it does

Each property says what it does, how **to see it** in the editor, and which code **to read**.

### The parts

It is a network of nodes joined by links. Each link goes from one node to another and
holds a direction, φ and θ, and a length, r. Each node holds the links that come into it.

- **To see it:** open the Links pill, the table of every link; each row is one link a→b
  with its φ, θ and r.
- **To read it:** `Categories/Node/Links/link.go`, type `Link`: `From`, `To`,
  `Start` (φ, θ, r), `Delta`, `Increment`.

### Steps

It changes where links point in 3D by steps. Each link has a change added to each of its two angles
and its length: Δφ is the change added to φ, Δθ the change added to θ, and Δr the change
added to r. Each one is a whole number of steps.

Stepping one back the same number of times returns it to exactly where it was. However
many steps are made, forward and back, it never drifts away from the count of steps that
were made.

- **To see it:** press a link's step up button (▲) and then its step down button (▼) the
  same number of times, and its Δ is back where it was.
- **To read it:** `link.go`, `Link.Step`: a step adds sign × increment to one whole-number
  Δ, and nothing else.

### Two angles

A change in where a link points is a number of steps in φ and a number of steps in θ. φ and θ are independent
angles: stepping one never changes the other. Where a link points is decided by those two
numbers alone. So steps in φ and θ can be made in any order, and the link ends up
pointing the same way.

- **To see it:** on one link step φ then θ, and read its start + Δ, the read-only row
  showing where the link points; step both back, then step θ then φ, and start + Δ is
  the same.
- **To read it:** `link.go`, `Link.Dir`: start + Δ, one sum per angle; `frame.go`,
  `Frame.Vec`: x, y, z made once, from those two sums.

### Moving nodes

A link stays attached to the node it starts from; only its angles and length change.
Stepping a link moves the node at its end to where that link now points. The other links touching that node are rebuilt from the two
nodes' new positions, so the picture stays in step with the links.

- **To see it:** step a link, and its end node moves while the links touching it follow.
- **To read it:** `peer_moves.go`, `Peer.take`: the end node goes to the other node's
  place plus the link; `Peer.sat` and `Peer.moved`: the other links are remeasured
  (`link.go`, `Link.Remeasure`).

### The network

A link can be selected: marked to be stepped together with the other selected links. One
step changes the angles of every selected link at once. Each node moves on its own, once it has heard
where its neighbours are. A node refuses changes that cannot be: a length below zero, or a
second selected link into the same node.

- **To see it:** in sequence 1 press the both ▲, the button that steps every selected
  link, and nodes 2 and 3 both move in one step.
- **To read it:**
  - `peer_moves.go`, `Peer.stepAll`, `Peer.tell` and `Peer.recover`: step, send the
    place, place from it.
  - `peer.go`, `Peer.Ready`: shown once every neighbour is heard.
  - `Peer.stepR` and `peer_edits.go`, `Peer.press`: a length below zero is refused.
  - `peer_edits.go`, `Peer.pickReq`: a second selected link is refused.

### Time

Changes play out at a speed a person can watch, and that speed is set from the UI.
Repeatedly stepping two links that start from different nodes makes the links move like
an arm bending at its joints.

- **To see it:** keep stepping two links that start from different nodes, and the links
  move like an arm bending.
- **To read it:** `Categories/Clock/human_tick.go`, `HumanTick`: every node counts ticks
  from one shared start.

### Sequences

A scene holds several sequences. Each sequence is a separate saved setup of nodes and
links.

- **To see it:** switch sequences, and each shows its own nodes and links.
- **To read it:** `Categories/Scene/Scenes/sequences.go`, `SequenceDir`: one folder per
  sequence.

### The screen

The simulation writes out every place and link, and the screen only draws them, so what
is on the screen is the simulation itself, not a second copy of it that could disagree.
Node places and links are saved as they change, so a scene reopens where it was left.

- **To see it:** close and reopen the scene, and it is where it was left.
- **To read it:** `Categories/Node/nodefile_place.go`, `WritePlace`, and
  `link_files.go`, `WriteLinks`: one value per file, written by the node that owns it.

---

## Roadmap

The local polar system (φ, θ, r in whole-number steps) is how the system runs, so the
existing parts (selection, camera, gestures) are changed to fit it.

The roadmap is at least a few months of work; it shows the direction, not a schedule. Each item,
and the part of the design it serves.

| | Item | Why |
|---|---|---|
| **Now** | Make the UI easier to use for all the settings (in progress) | Every property is checked by pressing buttons, so each setting must be reachable. In the 3 phi theta tab the settings are scattered across pill dropdowns, so the settings UI is getting a full redo before the settings can be tested. |
| **Next** | Make the "bird" animation | It should be easier once the settings UI is done; it is where the design is going ([question 7](summary_open_questions.md#7-what-should-it-be-able-to-do-that-it-cannot-do-now)). |
| **Next** | Test all the settings to make sure the screen shows the intent of the settings | The screen is how the local system is seen, so it must show what each setting means. |
| **Later** | Test sequences of many concurrent angle changes | Each node moves on its own, which is only shown when many move at once. |
| **Later** | Add more nodes for testing scale | Each node has its own s, so scale is tested across many nodes, not one. |
| **Later** | Build a node selection replacement | The current one is a central raycast through global x, y, z; the replacement has each node measure the mouse's angle in its own local polar steps ([questions 3 and 4](summary_open_questions.md#3-why-does-node-selection-get-a-replacement)). |
| **Later** | Build a camera replacement | The current one stacks decimal angle changes; the replacement is a node in the scene sphere whose angle ranges, in local polar steps, set the frame ([questions 5 and 6](summary_open_questions.md#5-why-does-the-camera-get-a-replacement)). |
| **Later** | Put the rings tab animation programming into the animation settings (after the camera) | So the ring tab animations are visible, easier to control, and set the same way as everything else. |
| **Later** | Redesign the gestures as pinch/unpinch, two-finger angle change and pan | So each gesture changes whole-number steps in the local polar system ([question 9](summary_open_questions.md#9-what-are-the-redesigned-gestures)). |

---

## Future directions

Where the design goes after the roadmap, and why that follows from it.

What is already there that points ahead:

- the control numbers are whole-number counts held locally by each link,
- each node moves on its own from what its neighbours tell it,
- and the screen only draws what the simulation wrote.

Answered:

- what this is for: making 3D simulations their way
  ([question 1](summary_open_questions.md#1-what-is-it-for-beyond-rotating-links)),
- what it should do next: a flapping "bird", once the settings UI is done
  ([question 7](summary_open_questions.md#7-what-should-it-be-able-to-do-that-it-cannot-do-now)),
- which matters most: only the bird so far
  ([question 8](summary_open_questions.md#8-which-of-those-matters-most)).
