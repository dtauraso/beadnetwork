---
name: feedback_one_value_two_places
description: Stretch, jitter, wobble and lag are usually one value living in two places; look for the second copy before adjusting either
metadata:
  type: feedback
---

Six visual defects in one session — panels stretching, the scene shrinking and expanding per
drag step, a scale wobbling by 9.87%, labels trailing their nodes, labels appearing beside
their nodes, panels drifting down the screen — were all the same fault: **one value held in
two places, with a delay or a disagreement between them.** They presented as six unrelated
symptoms and were fixed by six edits of the same shape.

**Why:** a rendering pipeline cannot express "these two numbers disagree" as nothing. It has
to resolve it, and the only ways it can are as a scale, an offset, or a lag. So a stretch is
not evidence of a sizing bug and a jitter is not evidence of a timing bug — both are evidence
that some quantity has two sources. The instances:

- fov derived by Go from the viewport height and streamed back → one frame drawn with the
  previous height's fov, every drag step
- the fov read from the DOM box while the aspect came from r3f's measurement → a fraction of
  a pixel apart, enough to wander
- the fov moved onto `useThree().size` → React state lands a frame *behind* the aspect r3f
  writes imperatively, so it got ten times worse
- the panel bitmap allocated from the view height → panels redrawn onto a different surface
  at every pane height
- label positions routed through `requestAnimationFrame` + `setState` → a fixed frame of lag,
  5.36px at the median while moving, 0 at rest
- the label row list and text held in `useState` → a copy of what the buffer already says

**How to apply:** when something stretches, jitters, wobbles or lags, ask "what quantity does
this depend on, and how many places is it computed or stored?" before adjusting anything.
Three specific traps, each of which cost real time here:

- **Picking a better source is not the fix, and can invert the sign.** Chasing which size to
  read was tried three times before the answer turned out to be that no size should be read
  to compute a scale. Transmit the quantity a side OWNS, then combine it locally with what
  the other side owns — never transmit the combination.
- **A quantity that is exactly right whenever its inputs are consistent is not a wrong value
  with a wrong source.** It is two values that should have been one. `ppwu` read exactly
  1.3626 on every settled frame while varying elsewhere; that pattern is the signature.
- **"It changes rarely, so it's worth a render/copy" is the same fault with an excuse.**
  Rarely stale is still stale. See [[feedback_reflect_dont_create_store]].

Measure before theorising ([[feedback_debug_data_before_theory]]): each of these was
identified from a logged number, and two confident fixes made from reasoning alone were
wrong — one of them made the defect ten times worse. The useful log holds BOTH values on one
line, so the disagreement is visible rather than inferred; a log of one side measuring itself
stays self-consistent under the very fault being hunted.

Related: [[feedback_make_bug_class_unrepresentable]] — the durable repair is a formulation
with no second source to disagree with, not a check that the two agree.
