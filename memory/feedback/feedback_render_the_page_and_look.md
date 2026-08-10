---
name: render the page and look at it
description: Self-derived. For any visual/layout change to the docs site, render it with headless Chrome and READ the screenshot. Reasoning from SVG coordinates and CSS got size, overflow and layout wrong three times in one session.
metadata:
  type: feedback
---

**Provenance:** self-derived on 2026-08-07, after David asked "look at the
rendered page?" — three rounds of size adjustments had all been wrong. See
[[feedback_debug_data_before_theory]], of which this is the visual case.

**The command** (the Chrome CLI, not the browser extension — no extension
needed):

```
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless=new --disable-gpu --hide-scrollbars --window-size=1400,2200 \
  --screenshot=shot.png "file:///abs/path/page.html"
```

Then Read the png. Render a second time at ~900px wide to check the responsive
layout, since the grid collapses at 720px.

**Why:** measuring the source cannot see what the page does.

- A bounds check on SVG `x`/`y` attributes says nothing about how wide the TEXT
  at that x renders. `<text x="486">6 — a quarter</text>` in a 520-wide viewBox
  passed the check and rendered clipped to "6 — a c".
- `width:100%` on an svg looks harmless and is not: a portrait canvas (520×606)
  stretched to an 1130px card renders 1340px tall — one circle filling a screen.
  Nothing in the source hints at this; the screenshot is obvious about it.
- Relative size is only judgeable rendered. "2× larger" against surrounding
  11px UI text was three separate wrong guesses from arithmetic, and one look.

**How to apply:** before reporting any layout/size/diagram change as done,
screenshot it and look. If the change was about SIZE, put the old and new
screenshots side by side rather than trusting the multiplier.
