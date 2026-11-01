# Summary ideas

What the current design does that is correct — not how it does it, but what it does.

- [Vision](#vision)
- [In short](#in-short)
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

## In short

There is a 3D simulation system made of nodes joined by links, each link going from one
node to another. It does 3D rotation by adding and subtracting integers to each link's
spherical-coordinate angles, φ and θ, centred on the node the link starts from. φ is how
far the link tips down from straight up; θ is how far it turns around the up direction.
What rotates is each link's direction, about its own centre.

Take one link. The node it starts from is its centre, and it rotates about that node. Its
angles are counts: a full turn is split into a whole number of equal steps, and φ and θ each
say how many of those steps. How many steps make a turn is set by s: a larger s splits the
turn into more, smaller steps, and the counts stay whole. Every link's angles are
measured against the same up and around directions; only the centre is the link's own. Rotating it adds to or subtracts
from those counts. Rotating it needs only its own counts and the place of the node it
starts from; no chain of other links is followed to reach it. It moves only the node at its
other end; nodes linked from that node stay where they are, unless their link is selected
too, and the links to them change to reach them. It also has a length, r, changed the same
way, which sets how far it reaches and never where it points.

Every link is that one link, at the same time. Each holds its own centre and its own
counts, and no other link's rotation is ever applied to it: when a node it joins is moved,
it is only remeasured to reach the nodes where they now are, and its counts stay as they
were. So any number of nodes are centres
at once, each link turning by its own amount, and none of them has to be pictured together:
one link is the whole of it.

It has a UI for setting angles, lengths and many other settings for 3D rotations.

### The tools

It is built only from tools in everyday use: spherical coordinates, whole-number counting,
addition, sin and cos and their inverses, subtracting one point from another, and nodes
passing messages.

---

## Design principles

The truth is the set of numbers everything else is worked out from. An AI's defaults keep
it in one global x, y, z frame, in decimals, and rotates by applying one rotation on top of
another, so each rotation rounds a little and the rounding builds up. Their way keeps it in
each link's own integer counts of φ, θ and r, so a rotation is an exact count of steps, the
counts never drift, and no rotation is ever applied on top of another. Each node's place is
worked out from the place of the node its link starts from plus that link's counts. Places
are decimals, so rounding enters only in measuring places, never in combining rotations, and
a step and its reverse land on exactly the same numbers.

---

## What it does

### The parts

It is a network of nodes joined by links. Each link goes from one node to another and
holds a direction, φ and θ, and a length, r, all measured from the node it starts from,
against the same up and around directions as every other link.

Each link's counts are counted from where it started. The start is set when the link is
made, and set again when the link is remeasured; a step never changes it, and the counts
stay whole numbers.

### Steps

A link is changed in steps: integers added to and subtracted from its φ, θ and r. Each
link has its own step size, a whole number, so one step adds or subtracts that many. Steps
in φ and θ rotate it; steps in r change its length. Stepping back the same number of times
returns a link to exactly where it was, however many steps are made.

Each node has its own s, a whole number, for the links into it: a full turn is 12 × s steps.
A larger s splits the turn into more steps, each a smaller part of the turn. Changing s
rescales the links' counts, and s is only set to a number that keeps them whole.

### Two angles

A rotation is steps added to φ and steps added to θ, and stepping one never changes the
other. Because the steps are added, they can be made in any order and the link ends up
pointing the same way. Stepping r never changes where it points.

A link pointing straight up or straight down lies on the line θ turns around, so stepping θ
leaves it where it is; stepping φ tips it off that line.

### Moving nodes

A link moves only the node at its other end. Rotating it swings that node about the node
the link starts from; stepping r moves that node nearer or farther along the link.

Nodes linked from that node are not carried along: they stay where they are, and each link
from the moved node to one of them changes to reach where that node still is. The one
exception is a link that is selected too: its end node is placed from the moved node's new
place, in the same step.

### The network

A rotation's centre is the start node of the link being rotated. Selecting links picks the
centres: each selected link's start node is a centre, and links selected from the same
node share it. One step changes every selected link at once, each by its own step size, and
every node those links move moves in that one step. Each node moves itself: its start node
sends it its place, and it places itself from that place plus its link. In one step, each
node is moved by at most one link, so a second selected link into the same node is refused,
as is a length below zero.

### Sequences

Motion over time is the series of angle changes made step after step: which links are
selected, so which nodes are centres, and how far each link turns. No link holds a motion;
a motion is only that series of steps.

A scene is what the editor opens and saves. It holds several sequences, each a separate
setup of nodes and links, kept on its own.

### The screen

What is on the screen is always what the simulation holds, and a scene reopens where it
was left.

---

## Roadmap

| | Item |
|---|---|
| **Now** | Make the UI easier to use for all the settings (in progress). |

After that, selection, the camera and the gestures, which still work in x, y, z, are moved
onto the same integer counts
([summary_open_questions.md](summary_open_questions.md)).
