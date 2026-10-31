# Open questions

Every question [summary_ideas.md](summary_ideas.md) leaves open, with the place it comes
from and an example of how an answer is expressed.

The examples are about other software on purpose (the database and the security system
from [summary_writing.md](summary_writing.md)), so they show the shape of an answer
without being mistaken for the answer. Write the real answer under **Answer**, then
replace the *Not answered yet* where the question comes from.

---

## 1. What is it for, beyond rotating links?

From: [Vision → Purpose](summary_ideas.md#purpose), [Future directions](summary_ideas.md#future-directions).

**Shape:** It lets *someone* do *something*, which *the usual way* cannot, because *reason*.
Say what it lets them do, not how it does it.

**Example:** "A migration tool lets a team change a live database's schema without taking
it offline, which hand-written migrations cannot, because they lock the tables they change."

**Answer:** It lets me make visual simulations by directing an AI, without reading the code
it writes. The usual way to trust code is to read it; I check by watching instead, so the
screen has to be the simulation itself, and every behaviour has to be testable by pressing
buttons and watching.

## 2. Who is it for?

From: [Vision → Purpose](summary_ideas.md#purpose).

**Shape:** A person or group named by what they are trying to do, not by job title alone.

**Example:** "Teams running a database that cannot go offline, who change its schema every
week."

**Answer:** A visual thinker for whom reading code and prose is the bottleneck.

## 3. Why does node selection get a replacement?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** The current one does *X*, which goes against *a part of the vision*; the
replacement serves that part. Name the part of the vision, not a dislike of the code.

**Example:** "The current secret store keeps keys in a config file, which breaks the rule
that a secret never touches disk; the replacement keeps them only in memory."

**Answer:** The current selection uses a raycast through global x, y, z, decided by one
central loop, so it skips the local polar system. The replacement selects with local polar
whole-number angles, measured by each node itself.

## 4. What does the node selection replacement have to do?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** Each thing it must do, written as a property someone can check by running it:
what to do, and what they then see.

**Example:** "Done when: a key is rotated while the service runs, and no file on disk ever
held it."

**Answer:**

```
           * A    sees M at ≈ 210°  → no
         ╱
  M ──────── * B  sees M at  180°   → selected
         ╲
           * C    sees M at ≈ 150°  → no
```

- The mouse is a node, M, with a link to each node near it.
- Each node measures the angle to M in its own local frame: φ and θ, in whole-number
  steps.
- **Selected:** the node that sees M at 180° ± k steps (a range, because an exact 180° is
  too hard to hit with a mouse). Every other node sees a different angle and does not
  respond.
- The other mouse states come from the gesture redesign
  ([question 9](#9-what-are-the-redesigned-gestures)).

## 5. Why does the camera get a replacement?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** Same as question 3: what the current one does, which part of the vision it goes
against, which part the replacement serves.

**Example:** "The current schema viewer redraws from its own cached copy, so it can show a
column that is already gone; the replacement reads the live schema only."

**Answer:** Same reason as the mouse. The current camera keeps decimal rotations stacked on
top of each other (`Categories/Scene/Camera/viewpoint.go`, `Rotate` and `Orbit`), outside
the local polar system. The replacement is a node, held in local polar whole-number steps
like everything else.

## 6. What does the camera replacement have to do?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** Same as question 4: checkable properties.

**Example:** "Done when: a column is dropped, and the viewer stops showing it without a
reload."

**Answer:**

```
          frame ┌─────────────────────┐
                │  ◯               ◯  │   ← each edge node sends
                │       *    *        │     its point + extra distance
                │  ◯       *       ◯  │
                └─────────────────────┘
   without the extra:  ◔  only a quarter of an edge node shows
```

- The camera is a node in the scene sphere, linked to every scene node (not the mouse
  node).
- Each node's angle from the camera is in whole-number φ, θ steps.
- The nodes at the edges of the range send their point plus an extra distance, so the
  whole node is inside the frame, not a quarter of it.
- The ranges, widened that way, become what three.js needs, made once from the totals
  (`Categories/Scene/Camera/SceneCamera.tsx`):

| Three.js needs | From the camera node |
|---|---|
| `position` | its link to the centre: φ, θ, r steps |
| `lookAt` | the middle of the edge nodes |
| `up` | its up direction: φ, θ steps |
| `fov` | the top-to-bottom range |
| `aspect` | the panel's width ÷ height |
| `near`, `far` | an r range, if wanted |

Notes:

- `fov` is the same angle above and below the middle, so the camera aims at the middle of
  the edge nodes, not at the centre node.
- `aspect` comes from the panel, so the camera fits whichever range is tighter, and the
  other follows from the panel with spare room.
- The extra distance is an angle that depends on how far away the node is:
  atan(node size ÷ distance), rounded up to whole steps so nothing is cut.

## 7. What should it be able to do that it cannot do now?

From: [Future directions](summary_ideas.md#future-directions).

**Shape:** It cannot yet *X*. It should, because *X* follows from *the bet* (or the
purpose). A capability, as an outcome, not a task list.

**Example:** "It cannot yet roll a schema change back. It should, because the purpose is
changing a live database safely, and a change that cannot be undone is not safe."

**Answer:** A "bird" flapping its wings. It may already be possible, but it has not been
tried, because the settings are scattered across pill dropdowns; it is the first thing to
make once the settings UI redo is done. It fits, because a bird is rotations combined and
repeated, which is what the local polar system already does (repeating two rotations with
different centres already swings links like an arm).

```
              F  front
              │
     L ────── C ────── R      C = centre, L/R = wings (sides),
              │                F/B = front and back
              B  back
```

- **Basic version, 5 nodes:** one centre node, two for front and back, two on the sides.
- **Flapping:** the side nodes rotate halfway around the centre and back.
- **Further:** the whole bird rotates while it flaps its wings.
- **Sequences:** the same centre can run two different sets of rotations, one after the
  other.

## 8. Which of those matters most?

From: [Future directions](summary_ideas.md#future-directions).

**Shape:** *X* matters most, because without it *the purpose* cannot be shown at all;
the others only make it better.

**Example:** "Rollback matters most: without it no team would trust the tool on a live
database, so speed and a nicer viewer would never be used."

**Answer:** Only one so far: the bird, so there is nothing to rank yet.

## 9. What are the redesigned gestures?

From: [Roadmap](summary_ideas.md#roadmap). The gestures may be redone from three:
pinch/unpinch, rotate, and pan.

**Shape:** For each gesture: what the hand does, and which whole-number Δ (φ, θ or r) on
which node it changes. A gesture that changes no Δ is outside the local system.

**Example:** "In a map app: pinch/unpinch changes the zoom level by whole levels; rotate
turns the map by whole 15° steps; pan moves the map by whole tiles."

**Answer:** The gestures will be redone from pinch/unpinch, rotate and pan, each one
changing whole-number steps in the local polar system. What each gesture changes, and
which current states get replaced, will be decided when it is built.
