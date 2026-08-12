"""The pair-node update drawings: one file per case.

Each panel is referenced from updates.html with <img>, which does NOT inherit the
site stylesheet, so every text style is baked in here. A standalone .svg also needs
xmlns or it renders as alt text.

GEOMETRY IS HALVED FROM THE FIRST CUT, TYPE IS NOT. Scaling the whole image down
would take the labels with it, which is the unreadability that was just fixed; so
the rings shrink and the 15px text stays 15px, and each canvas is sized to hold
whichever of the two is wider.
"""
import math, os

os.chdir(os.path.dirname(os.path.abspath(__file__)))

TOP, NORM, BOT, ARR, HOME, CAND = "#5fd68a", "#4ea1ff", "#9a9aa6", "#f9a825", "#e05fd8", "#4ea1ff"
MONO = 'font-family="ui-monospace,SFMono-Regular,Menlo,monospace" font-size="15"'
SANS = 'font-family="-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif" font-size="15"'
INK, DIM = 'fill="#e7e7ea"', 'fill="#9a9aa6"'

def P(cx, cy, R, idx):
    a = math.radians(15 * idx)
    return (round(cx + R * math.sin(a)), round(cy - R * math.cos(a)))

def arrow(cx, cy, R, idx, col, w=2.4, frac=0.84):
    x, y = P(cx, cy, R * frac, idx)
    tx, ty = P(cx, cy, R * frac + 9, idx)
    a = math.radians(15 * idx)
    px, py = math.cos(a), math.sin(a)
    b1 = (round(x + px * 4.5), round(y + py * 4.5))
    b2 = (round(x - px * 4.5), round(y - py * 4.5))
    return (f'<line x1="{cx}" y1="{cy}" x2="{x}" y2="{y}" stroke="{col}" stroke-width="{w}"/>'
            f'<polygon points="{tx},{ty} {b1[0]},{b1[1]} {b2[0]},{b2[1]}" fill="{col}"/>')

def ring(cx, cy, R, ticks=True):
    s = f'<circle cx="{cx}" cy="{cy}" r="{R}" fill="none" stroke="#3a3a44" stroke-width="1.4"/>'
    if ticks:
        maj, mnr = [], []
        for i in range(24):
            x1, y1 = P(cx, cy, R - 5, i)
            x2, y2 = P(cx, cy, R, i)
            (maj if i % 6 == 0 else mnr).append(f"M{x1} {y1}L{x2} {y2}")
        s += (f'<path d="{"".join(mnr)}" stroke="#3a3a44"/>'
              f'<path d="{"".join(maj)}" stroke="#9a9aa6" stroke-width="1.6"/>')
    return s

def arc(cx, cy, R, i0, i1, col, w=3):
    x0, y0 = P(cx, cy, R, i0)
    x1, y1 = P(cx, cy, R, i1)
    large = 1 if (i1 - i0) % 24 > 12 else 0
    return f'<path d="M{x0} {y0}A{R} {R} 0 {large} 1 {x1} {y1}" fill="none" stroke="{col}" stroke-width="{w}"/>'

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

W, H = 620, 176
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">']
xs = [70, 190, 310, 430, 550]
names = ["21", "22", "23", "0", "1"]
o.append(f'<line x1="370" y1="18" x2="370" y2="158" stroke="#3a3a44" stroke-dasharray="4 4"/>')
for x, name in zip(xs, names):
    o.append(f'<circle cx="{x}" cy="88" r="26" fill="#2b2b33" stroke="#3a3a44" stroke-width="1.4"/>')
    o.append(f'<text x="{x}" y="94" text-anchor="middle" {MONO} {INK}>{name}</text>')

for a, b in zip(xs, xs[1:]):
    o.append(f'<line x1="{a + 30}" y1="78" x2="{b - 36}" y2="78" stroke="{TOP}" stroke-width="2.2"/>')
    o.append(f'<polygon points="{b - 28},78 {b - 38},73 {b - 38},83" fill="{TOP}"/>')
    o.append(f'<line x1="{b - 30}" y1="100" x2="{a + 36}" y2="100" stroke="{HOME}" stroke-width="2.2"/>')
    o.append(f'<polygon points="{a + 28},100 {a + 38},95 {a + 38},105" fill="{HOME}"/>')
o.append(f'<text x="130" y="46" text-anchor="middle" {SANS} fill="{TOP}">next</text>')
o.append(f'<text x="490" y="140" text-anchor="middle" {SANS} fill="{HOME}">prev</text>')
o.append(f'<text x="370" y="170" text-anchor="middle" {SANS} {DIM}>the seam</text>')
o.append('</svg>')
out["panel-links"] = "\n".join(o)

W, H = 620, 250
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">']
boxes = [(24, 102, 140, 46, "setting", "#9a9aa6"),
         (400, 30, 190, 46, "perpendicular", HOME),
         (400, 174, 190, 46, "parallel", ARR)]
for x, y, w, h, name, col in boxes:
    o.append(f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="9" fill="#2b2b33" stroke="{col}"/>')
    o.append(f'<text x="{x + w // 2}" y="{y + 29}" text-anchor="middle" {MONO} fill="{col}">{name}</text>')

def straight(x1, y1, x2, y2, col, dash=""):
    n = ((x2 - x1) ** 2 + (y2 - y1) ** 2) ** 0.5
    ux, uy = (x2 - x1) / n, (y2 - y1) / n
    bx, by = x2 - 11 * ux, y2 - 11 * uy
    return (f'<line x1="{x1}" y1="{y1}" x2="{round(bx)}" y2="{round(by)}" stroke="{col}" '
            f'stroke-width="2"{dash}/>'
            f'<polygon points="{x2},{y2} {round(bx - 5 * uy)},{round(by + 5 * ux)} '
            f'{round(bx + 5 * uy)},{round(by - 5 * ux)}" fill="{col}"/>')

o.append(straight(168, 118, 396, 56, TOP))
o.append(straight(168, 132, 396, 194, TOP))
o.append(f'<text x="286" y="118" text-anchor="middle" {SANS} fill="{TOP}">the gap, at the</text>')
o.append(f'<text x="286" y="136" text-anchor="middle" {SANS} fill="{TOP}">first arrival</text>')

o.append(straight(400, 40, 172, 106, "#ff6b6b", ' stroke-dasharray="5 4"'))
o.append(straight(400, 210, 172, 144, "#ff6b6b", ' stroke-dasharray="5 4"'))
o.append(f'<text x="286" y="60" text-anchor="middle" {SANS} fill="#ff6b6b">RESET</text>')
o.append(f'<text x="286" y="204" text-anchor="middle" {SANS} fill="#ff6b6b">RESET</text>')
o.append(f'<text x="310" y="240" text-anchor="middle" {SANS} {DIM}>no edge between the two chosen modes — a choice sticks</text>')
o.append('</svg>')
out["panel-modes"] = "\n".join(o)

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

W, H, cx, cy, R = 436, 280, 150, 140, 106
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
o.append(arc(cx, cy, R + 12, 3, 4, HOME, 4))
wx, wy = P(cx, cy, R + 12, 3.5)
o.append(f'<path d="M{wx + 6} {wy - 2} L280 66" stroke="{HOME}" stroke-width="1"/>')
o.append(f'<text x="288" y="70" {MONO} fill="{HOME}" text-anchor="start">one step = π/12</text>')
for i, lab in [(0, "0"), (6, "6"), (12, "12"), (18, "18")]:
    x, y = P(cx, cy, R + 18, i)
    anc = "middle" if i % 12 == 0 else ("start" if i == 6 else "end")
    o.append(f'<text x="{x}" y="{y + 6}" text-anchor="{anc}" {SANS} {DIM}>{lab}</text>')
for i in range(24):
    x, y = P(cx, cy, R, i)
    o.append(f'<circle cx="{x}" cy="{y}" r="4" fill="{"#5fd68a" if i % 6 == 0 else "#4a4a56"}"/>')
o.append(arc(cx, cy, R + 7, 23, 0, "#4ea1ff", 3))
o.append('</svg>')
out["panel-state"] = "\n".join(o)

W, H, cx, cy, R = 300, 300, 150, 140, 108
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
for h in (6, 18):
    x, y = P(cx, cy, R, h)
    o.append(f'<circle cx="{x}" cy="{y}" r="7" fill="{HOME}"/>')
    o.append(f'<text x="{x}" y="{y + 24}" text-anchor="middle" {SANS} fill="{HOME}">rests</text>')
for c in (9, 11):
    x, y = P(cx, cy, R, c)
    o.append(f'<circle cx="{x}" cy="{y}" r="6" fill="#2f2f37" stroke="{CAND}" stroke-width="2.4"/>')
    lx, ly = P(cx, cy, R + 16, c)
    o.append(f'<text x="{lx}" y="{ly + 16}" text-anchor="middle" {SANS} fill="{CAND}">{c}</text>')
o.append(arrow(cx, cy, R, 0, ARR, 2.6))
o.append(f'<text x="{cx}" y="22" text-anchor="middle" {SANS} fill="{ARR}">a = 0</text>')
o.append(arrow(cx, cy, R, 10, TOP, 3))
o.append(f'<text x="{cx - 16}" y="{cy + 66}" text-anchor="end" {SANS} fill="{TOP}">t = 10</text>')
o.append('</svg>')
out["panel-step"] = "\n".join(o)

W, H, cx, cy, R = 260, 250, 130, 125, 104
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
o.append(arrow(cx, cy, R, 17, ARR, 2.4))
o.append(arrow(cx, cy, R, 15, BOT, 2.2))
o.append(arrow(cx, cy, R, 9, NORM, 2.6))
o.append(arrow(cx, cy, R, 3, TOP, 3))
o.append('</svg>')
out["panel-frame"] = "\n".join(o)

for k, v in out.items():
    v = v.replace('<svg ', '<svg xmlns="http://www.w3.org/2000/svg" ', 1)
    open(k + ".svg", "w").write(v)
print(len(out), "panels written")
