# Summary ideas

What the current design does that is correct — not how it does it, but what it does.
Every property below can be seen by running the editor.

- [Vision](#vision)
- [Design principles](#design-principles)
- [In one paragraph](#in-one-paragraph)
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
second.

### What follows from the priority

- **The screen draws what the simulation wrote.**

### Where it is going

Many nodes will rotate at once, enough to make a bird that flaps its wings and turns while it
flaps, checked by watching it ([Roadmap](#roadmap)).

---

## Design principles

An AI's defaults keep the truth in one global x, y, z frame. Their way keeps it in each
link's own integer angle values: 3D rotation is done by adding and subtracting them, so a
rotation is an exact count of steps that never drifts. x, y, z are only worked out from
those values, to place and draw the nodes.

It rules out:

- **Selecting nodes with one central raycast through global x, y, z**
  ([question 3](summary_open_questions.md#3-why-does-node-selection-get-a-replacement)).
- **A camera made of decimal rotations stacked on each other**
  ([question 5](summary_open_questions.md#5-why-does-the-camera-get-a-replacement)).

---

## In one paragraph

There is a 3D simulation system. It does 3D rotation by adding and subtracting integer
angle values, held locally by each link. It has a UI for setting angles and many other
settings for 3D rotations.

### The tools

It is built only from tools in everyday use: polar coordinates, whole-number counting,
addition, sin and cos and their inverses, subtracting one point from another, nodes
passing messages, and a clock.

---

## What it does

Each property says what it does, and how **to see it** in the editor.

### The parts

It is a network of nodes joined by links. Each link goes from one node to another and
holds a direction, φ and θ, and a length, r.

- **To see it:** open the Links pill, the table of every link; each row is one link a→b
  with its φ, θ and r.

### Steps

It does 3D rotation by adding and subtracting integer angle values. Stepping back the same
number of times returns a link to exactly where it was, however many steps are made.

- **To see it:** press a link's step up button (▲) and then its step down button (▼) the
  same number of times, and its Δ is back where it was.

### Two angles

A rotation is steps in φ and steps in θ, and stepping one never changes the other, so they
can be made in any order and the link ends up pointing the same way.

- **To see it:** on one link step φ then θ, and read its start + Δ; step both back, then
  step θ then φ, and start + Δ is the same.

### Moving nodes

A link rotates about the node it starts from, and the node at its end moves with it.

- **To see it:** step a link, and its end node moves while the links touching it follow.

### The network

Selected links rotate together in one step. Each node moves on its own, and refuses
changes that cannot be: a length below zero, or a second selected link into the same node.

- **To see it:** in sequence 1 press the both ▲, the button that steps every selected
  link, and nodes 2 and 3 both move in one step.

### Time

Changes play out at a speed a person can watch, set from the UI. Repeating two rotations
with different centres makes the links swing like an arm on its joints.

- **To see it:** keep repeating two rotations with different centres, and the links swing
  like an arm.

### Sequences

A scene holds several sequences, each a separate saved setup of nodes and links.

- **To see it:** switch sequences, and each shows its own nodes and links.

### The screen

What is on the screen is the simulation itself, not a second copy that could disagree, and
a scene reopens where it was left.

- **To see it:** close and reopen the scene, and it is where it was left.

---

## Roadmap

| | Item |
|---|---|
| **Now** | Make the UI easier to use for all the settings (in progress). |
| **Next** | Make the "bird" animation ([question 7](summary_open_questions.md#7-what-should-it-be-able-to-do-that-it-cannot-do-now)). |

After that, selection, the camera and the gestures are moved onto integer angle values
too ([summary_open_questions.md](summary_open_questions.md)).
