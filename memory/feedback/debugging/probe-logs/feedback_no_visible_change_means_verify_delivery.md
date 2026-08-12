---
name: no-visible-change-means-verify-delivery
description: A repeated "I see no change" is not a signal to adjust harder — it is a signal to verify the change reaches the screen, then ask what IS on screen. Two rounds of pixel-guessing preceded the one question that settled it.
type: feedback
---

**Rule:** when the user reports no visible change twice, stop adjusting. Verify the new code is what is running, then ask what is actually on screen. Adjusting harder assumes delivery, which is the thing in doubt.

**Where this was paid for (2026-08-05, speed-slider tick labels):** four attempts, only the last aimed at the real defect.

- *"They're barely visible"* — the labels were `#ddd` at 0.5 opacity, the DARK overlay-panel palette, but the slider portals into `.toolbar`, which is `background: #fff`. Near-white on white. **Check which surface a widget sits on before picking a colour; this webview's two palettes are not interchangeable.**
- *"A few px different, not completely different glyphs"* — the fix rebuilt each fraction from full-size digits around a slash, changing how big everything looked. The ask was a gap, and only a gap.
- *"The ticks are the same"*, twice — both times the deployed code WAS current (bundle fresh, cache-busted on mtime). The changes were simply too small to see. The useful move was not another adjustment but asking what was on screen, which established in one question that the new code was live and the problem was sizing.

**The real defect, visible only in a screenshot:** ¼/½/¾ render as precomposed stacked glyphs whose internal spacing is not adjustable. The fraction has to be COMPOSED to make the gap settable — but at `FRAC_EM`, the size the glyph draws its own digits at, so nothing changes size and only the gap moves.

**Related trap — a field added to preserve a behaviour can hide the evidence a change happened at all.** (2026-08-06, pair-one-implementation) Restoring a sign inversion as a `Sign` field made end 2's coplanar normal point opposite end 1's, so the editor looked completely unchanged. A screenshot showed node 1's normal pointing right and node 2's left. The field was out; both ends now derive the same directions.

See [[render-the-page-and-look]], [[runtime-breadcrumbs-beat-static-analysis]].
