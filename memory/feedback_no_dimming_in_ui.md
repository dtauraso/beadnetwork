---
name: feedback-no-dimming-in-ui
description: Never dim UI text/controls (opacity or a greyed colour) to de-emphasise them; distinguish by weight, size or position instead
metadata:
  type: feedback
---

Do not dim things in the editor UI. No `opacity` below 1 and no greyed-down colour
(`#555`, `#888`, `#ddd` against a light surface) applied to make something recede.

**Why:** stated directly by David — "don't dim things". A label that is on screen is on
screen to be read; dimming trades readability for emphasis that some other property already
carries. In the speed slider's tick row this meant five of six settings were harder to read
so the sixth could be marked — which `fontWeight: bold` was already doing on its own.

**How to apply:** mark state with WEIGHT, SIZE or POSITION, never with reduced contrast. If
something truly should not be read yet, do not render it. Related: check the surface a
widget actually sits on before picking a colour — the dark-overlay palette (`#ddd` text) on
the light `.toolbar` (`background: #fff`) rendered labels invisible, a mistake that reads
the same as deliberate dimming. See [[feedback_visuals_scrutiny]].
