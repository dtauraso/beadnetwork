# Summary ideas

What the current design does that is correct — not so much how it does it, but what it does.
Every property below can be seen by running the code.

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

---

## Vision

### Purpose

*Not answered yet:* what this is for beyond rotating links, and who it is for. This is the
first thing to write; the rest of the vision hangs from it.

### The bet

Motion is held as local, whole-number counts, and the global picture is only a view of
them. The usual approach keeps one global set of x, y, z floats and rotates it, so error
builds up and the order of the rotations leaks into the result. Here each link keeps its
own count of steps in φ, θ and r, and x, y, z are made from those counts once, at the end.

### What follows from the bet

- **Local data is the truth.** Each link's own start and Δ decide where it points; the
  global place is rebuilt from them, never the other way round for the control numbers.
- **Steps never drift.** A rotation is a count, so stepping back returns exactly to where
  it was, however many steps were made.
- **Order does not matter.** φ and θ are added on their own, so steps can be made in any
  order and land in the same place.
- **Each node moves on its own.** A node acts only on what its neighbours tell it; nothing
  coordinates the network from above.
- **The screen is the simulation.** The screen only draws what the simulation wrote, so
  there is no second copy that could disagree.

### What it rules out

- Accumulating rotations on top of each other in global floats, because that brings back
  drift and order dependence.
- Letting the global view become a second source of truth; it may be floats only where it
  is rebuilt and is not part of the control system.
- A central coordinator moving the nodes, because each node moving on its own is part of
  what is being shown.
- A screen that works out positions of its own.

### Where it is going

The roadmap makes every setting reachable and visible, then tests many concurrent
rotations and many nodes, which is where local counting and independent nodes are meant to
hold up. Past the roadmap, see [Future directions](#future-directions).

---

## Design principles

| Principle | Meaning |
|---|---|
| Local data as truth | Each link's own numbers are the source of truth. |
| Polar as the local coordinate system | Links are φ, θ, r — not x, y, z. |
| Integers as the control numbers | Changes in the local system are whole-number counts of steps. |
| Global as projection | The global picture is a view made from the local data. |
| Local → global adaptor | Turns local places and links into x, y, z. |
| Global reconstructions → local | The parts rebuilt from global places may be floats, and are not the main control system. |

The local-to-global adaptor was picked to get the local system visible on the screen, so
it can be tested.

---

## In one paragraph

There is a 3D simulation system. It uses local integer changes to angles to do 3D
rotations. It has a UI for setting angles and many other settings for 3D rotations.

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

It rotates links in 3D by steps. Each link has a change added to each of its two angles
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

A rotation is a number of steps in φ and a number of steps in θ. φ and θ are independent
angles: stepping one never changes the other. Where a link points is decided by those two
numbers alone. So steps in φ and θ can be made in any order, and the link ends up
pointing the same way.

- **To see it:** on one link step φ then θ, and read its start + Δ, the read-only row
  showing where the link points; step both back, then step θ then φ, and start + Δ is
  the same.
- **To read it:** `link.go`, `Link.Dir`: start + Δ, one sum per angle; `frame.go`,
  `Frame.Vec`: x, y, z made once, from those two sums.

### Moving nodes

A link turns about the node it starts from. Stepping a link moves the node at its end to
where that link now points. The other links touching that node are remeasured from where
their two nodes now are, so every link always describes where its two nodes really are.

- **To see it:** step a link, and its end node moves while the links touching it follow.
- **To read it:** `peer_moves.go`, `Peer.take`: the end node goes to the other node's
  place plus the link; `Peer.sat` and `Peer.moved`: the other links are remeasured
  (`link.go`, `Link.Remeasure`).

### The network

A link can be selected: marked to be turned together with the other selected links. One
step rotates every selected link at once. Each node moves on its own, once it has heard
where its neighbours are. It refuses changes that cannot be: a length below zero, or a
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
Repeating two different rotations with different centres makes the links swing like an
arm on its joints.

- **To see it:** keep repeating two rotations with different centres, and the links swing
  like an arm.
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

Each item, and the part of the design it serves.

| Item | Why |
|---|---|
| Make the UI easier to use for all the settings | Every property is checked by pressing buttons, so each setting must be reachable. |
| Test all the settings to make sure the screen shows the intent of the settings | The screen is how the local system is seen, so it must show what each setting means. |
| Test sequences of many concurrent rotation changes | Each node moves on its own, which is only shown when many move at once. |
| Add more nodes for testing scale | Each node has its own s, so scale is tested across many nodes, not one. |
| Build a node selection replacement | *Not answered yet.* |
| Build a camera replacement | *Not answered yet.* |
| Put the rings tab animation programming into the animation settings | So the ring tab animations are visible, easier to control, and set the same way as everything else. |

**Not answered yet:** why the node selection and the camera get replacements, and what
each replacement has to do.

---

## Future directions

Where the design goes after the roadmap, and why that follows from it.

What is already there that points ahead:

- the control numbers are whole-number counts held locally by each link,
- each node moves on its own from what its neighbours tell it,
- and the screen only draws what the simulation wrote.

**Not answered yet:**

- what this is for beyond rotating links,
- what it should be able to do that it cannot do now,
- and which of those is the one that matters most.
