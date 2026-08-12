from panel_primitives import TOP, NORM, BOT, ARR, HOME, MONO, SANS, INK, DIM, P, arrow, ring, arc
from panel_defs_extra import (
    build_panel_links,
    build_panel_modes,
    build_panel_state,
    build_panel_step,
    build_panel_frame,
)


def build_panels():
    out = {}

    for t in (0, 6, 12, 17):
        W, H, cx, cy, R = 152, 194, 76, 80, 64
        o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
        o.append(arrow(cx, cy, R, t + 12, BOT, 2))
        o.append(arrow(cx, cy, R, t + 6, NORM, 2.4))
        o.append(arrow(cx, cy, R, t, TOP, 2.8))
        o.append(f'<text x="{cx}" y="176" text-anchor="middle" {MONO} {INK}>t = {t}</text>')
        o.append('</svg>')
        out[f"panel-triad-{t}"] = "\n".join(o)

    for sep, note in ((0, "perpendicular rests"), (3, "acute — steps"), (6, "parallel rests"),
                      (11, "obtuse — steps"), (12, "perpendicular rests")):
        W, H, cx, cy, R = 186, 218, 93, 82, 66
        o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
        if sep:
            o.append(arc(cx, cy, round(R * 0.5), 0, sep, HOME))
        o.append(arrow(cx, cy, R, 0, TOP, 4 if sep == 0 else 2.8))
        o.append(arrow(cx, cy, R, sep, ARR, 2.4, 0.84 if sep else 0.55))
        o.append(f'<text x="{cx}" y="180" text-anchor="middle" {MONO} fill="{HOME}">L = {sep}</text>')
        o.append(f'<text x="{cx}" y="202" text-anchor="middle" {SANS} {DIM}>{note}</text>')
        o.append('</svg>')
        out[f"panel-length-{sep}"] = "\n".join(o)

    frames = [(23, 10, None, None, ["n1 sends 5"]),
              (23, 10, 2, 5, ["n2 reads 5,", "one slot short"]),
              (23, 11, None, None, ["n2 turns to 11,", "sends 17"]),
              (23, 11, 1, 17, ["n1 reads 17: L 6,", "settles"])]
    for n, (t1, t2, who, ai, cap) in enumerate(frames):
        W, H, R = 152, 258, 45
        o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">']
        for row, (t, node) in enumerate([(t1, 1), (t2, 2)]):
            cx, cy = 92, 56 + row * 104
            o.append(ring(cx, cy, R) + arrow(cx, cy, R, t, TOP, 2.4))

            if who == node:
                o.append(arrow(cx, cy, R, ai, ARR, 2.4))
            o.append(f'<text x="{cx - R - 8}" y="{cy + 5}" text-anchor="end" {MONO} {INK}>n{node}</text>')
        for i, line in enumerate(cap):
            o.append(f'<text x="76" y="{224 + i * 19}" text-anchor="middle" {SANS} {DIM}>{line}</text>')
        o.append('</svg>')
        out[f"panel-run-{n + 1}"] = "\n".join(o)

    for name, t, extra in (("before", 17, True), ("after", 0, False)):
        W, H, cx, cy, R = 164, 204, 82, 82, 66
        o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
        o.append(arrow(cx, cy, R, (t + 12) % 24, BOT, 2))
        o.append(arrow(cx, cy, R, (t + 6) % 24, NORM, 2.4))
        o.append(arrow(cx, cy, R, t, TOP, 2.8))
        if extra:
            o.append(arrow(cx, cy, R, 8, ARR, 2.2))
        label = "t = 17, r set" if extra else "t = 0, r unset"
        o.append(f'<text x="{cx}" y="180" text-anchor="middle" {MONO} {INK}>{label}</text>')
        o.append(f'<text x="{cx}" y="200" text-anchor="middle" {SANS} {DIM}>{name}</text>')
        o.append('</svg>')
        out[f"panel-reset-{name}"] = "\n".join(o)

    out["panel-links"] = build_panel_links()
    out["panel-modes"] = build_panel_modes()

    for name, a, note in (("perp", 12, "gap 6 → perpendicular"), ("par", 9, "gap 3 → parallel")):
        W, H, cx, cy, R = 200, 232, 100, 84, 68
        partner = (a - 6) % 24
        o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
        o.append(arc(cx, cy, round(R * 0.46), 0, partner, HOME))
        o.append(arrow(cx, cy, R, a, ARR, 2.2))

        px, py = P(cx, cy, R * 0.84, partner)
        o.append(f'<line x1="{cx}" y1="{cy}" x2="{px}" y2="{py}" stroke="#9a9aa6" stroke-width="2" '
                 f'stroke-dasharray="5 4"/>')
        o.append(arrow(cx, cy, R, 0, TOP, 2.8))
        o.append(f'<text x="{cx}" y="196" text-anchor="middle" {SANS} {DIM}>{note}</text>')
        o.append(f'<text x="{cx}" y="218" text-anchor="middle" {SANS} {DIM}>a = {a}, partner = {partner}</text>')
        o.append('</svg>')
        out[f"panel-gap-{name}"] = "\n".join(o)

    out["panel-state"] = build_panel_state()
    out["panel-step"] = build_panel_step()
    out["panel-frame"] = build_panel_frame()

    return out
