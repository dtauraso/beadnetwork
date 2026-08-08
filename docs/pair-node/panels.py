import math, os
os.chdir(os.path.dirname(os.path.abspath(__file__)))

TOP, NORM, BOT, ARR, HOME = "#5fd68a", "#4ea1ff", "#9a9aa6", "#f9a825", "#e05fd8"

# These panels are referenced with <img>, which does NOT inherit the page's CSS, so
# every text style is baked in here rather than left to .lbl / .lblm.
# NO fill here: several call sites add their own, and two fill attributes on one
# element is a duplicate-attribute parse error — the file silently stops being an
# image and <img> falls back to alt text.
MONO = 'font-family="ui-monospace,SFMono-Regular,Menlo,monospace" font-size="15"'
SANS = 'font-family="-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif" font-size="15"'
INK, DIM = 'fill="#e7e7ea"', 'fill="#9a9aa6"'


def P(cx, cy, R, idx):
    a = math.radians(15 * idx)
    return (round(cx + R * math.sin(a)), round(cy - R * math.cos(a)))


def arrow(cx, cy, R, idx, col, w=3, frac=0.84):
    x, y = P(cx, cy, R * frac, idx)
    tx, ty = P(cx, cy, R * frac + 13, idx)
    a = math.radians(15 * idx)
    px, py = math.cos(a), math.sin(a)
    b1 = (round(x + px * 6.5), round(y + py * 6.5))
    b2 = (round(x - px * 6.5), round(y - py * 6.5))
    return (f'<line x1="{cx}" y1="{cy}" x2="{x}" y2="{y}" stroke="{col}" stroke-width="{w}"/>'
            f'<polygon points="{tx},{ty} {b1[0]},{b1[1]} {b2[0]},{b2[1]}" fill="{col}"/>')


def ring(cx, cy, R, ticks=True):
    s = f'<circle cx="{cx}" cy="{cy}" r="{R}" fill="none" stroke="#3a3a44" stroke-width="1.6"/>'
    if ticks:
        maj, mnr = [], []
        for i in range(24):
            x1, y1 = P(cx, cy, R - 7, i)
            x2, y2 = P(cx, cy, R, i)
            (maj if i % 6 == 0 else mnr).append(f"M{x1} {y1}L{x2} {y2}")
        s += (f'<path d="{"".join(mnr)}" stroke="#3a3a44" stroke-width="1.2"/>'
              f'<path d="{"".join(maj)}" stroke="#9a9aa6" stroke-width="2"/>')
    return s


def arc(cx, cy, R, i0, i1, col, w=4):
    x0, y0 = P(cx, cy, R, i0)
    x1, y1 = P(cx, cy, R, i1)
    large = 1 if (i1 - i0) % 24 > 12 else 0
    return f'<path d="M{x0} {y0}A{R} {R} 0 {large} 1 {x1} {y1}" fill="none" stroke="{col}" stroke-width="{w}"/>'


out = {}

# ---- the triad, one panel per tilt
for t in (0, 6, 12, 17):
    W, H, cx, cy, R = 270, 322, 135, 145, 116
    o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
    o.append(arrow(cx, cy, R, t + 12, BOT, 2.6))
    o.append(arrow(cx, cy, R, t + 6, NORM, 3))
    o.append(arrow(cx, cy, R, t, TOP, 3.6))
    o.append(f'<text x="{cx}" y="296" text-anchor="middle" {MONO} {INK}>t = {t}</text>')
    o.append('</svg>')
    out[f"panel-triad-{t}"] = "\n".join(o)

# ---- the angle length, one panel per case
for sep, note in ((0, "perpendicular rests"), (6, "parallel rests"), (11, "steps")):
    W, H, cx, cy, R = 300, 356, 150, 150, 120
    o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
    if sep:
        o.append(arc(cx, cy, round(R * 0.5), 0, sep, HOME))
    o.append(arrow(cx, cy, R, 0, TOP, 5 if sep == 0 else 3.6))
    o.append(arrow(cx, cy, R, sep, ARR, 3, 0.84 if sep else 0.58))
    o.append(f'<text x="{cx}" y="304" text-anchor="middle" {MONO} fill="{HOME}">L = {sep}</text>')
    o.append(f'<text x="{cx}" y="330" text-anchor="middle" {SANS} {DIM}>{note}</text>')
    o.append('</svg>')
    out[f"panel-length-{sep}"] = "\n".join(o)

# ---- the run, one panel per step
frames = [(23, 10, None, None, ["n1 sends 5"]),
          (23, 10, 2, 5, ["n2 reads 5,", "one slot short"]),
          (23, 11, None, None, ["n2 turns to 11,", "sends 17"]),
          (23, 11, 1, 17, ["n1 reads 17: L 6,", "settles"])]
for n, (t1, t2, who, ai, cap) in enumerate(frames):
    # H leaves room BELOW the second ring (which ends at 286 + 82 = 368) for the
    # caption — at 400 the first caption line sat on top of the ring.
    W, H, R = 250, 432, 82
    o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">']
    for row, (t, node) in enumerate([(t1, 1), (t2, 2)]):
        cx, cy = 140, 100 + row * 186
        # ticks ON: the whole point of a frame is WHICH slot the arrow sits in
        o.append(ring(cx, cy, R) + arrow(cx, cy, R, t, TOP, 3))
        if who == node:
            o.append(arrow(cx, cy, R, ai, ARR, 3))
        o.append(f'<text x="{cx - R - 12}" y="{cy + 5}" text-anchor="end" {MONO} {INK}>n{node}</text>')
    for i, line in enumerate(cap):
        o.append(f'<text x="125" y="{394 + i * 20}" text-anchor="middle" {SANS} {DIM}>{line}</text>')
    o.append('</svg>')
    out[f"panel-run-{n + 1}"] = "\n".join(o)

# ---- reset, one panel per side
for name, t, extra in (("before", 17, True), ("after", 0, False)):
    W, H, cx, cy, R = 280, 336, 140, 150, 118
    o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
    o.append(arrow(cx, cy, R, (t + 12) % 24, BOT, 2.6))
    o.append(arrow(cx, cy, R, (t + 6) % 24, NORM, 3))
    o.append(arrow(cx, cy, R, t, TOP, 3.6))
    if extra:
        o.append(arrow(cx, cy, R, 8, ARR, 3))
    label = "t = 17, r set" if extra else "t = 0, r unset"
    o.append(f'<text x="{cx}" y="300" text-anchor="middle" {MONO} {INK}>{label}</text>')
    o.append(f'<text x="{cx}" y="324" text-anchor="middle" {SANS} {DIM}>{name}</text>')
    o.append('</svg>')
    out[f"panel-reset-{name}"] = "\n".join(o)

# ---- the state: the ring itself, its axis marks and the one-step wedge
W, H, cx, cy, R = 520, 500, 260, 250, 196
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
o.append(arc(cx, cy, R + 22, 3, 4, HOME, 5))
o.append('<path d="M368 150 L408 112" stroke="#e05fd8" stroke-width="1.2"/>')
o.append(f'<text x="300" y="158" {MONO} fill="{HOME}" text-anchor="middle">one step = π/12</text>')
for i, lab in [(0, "0"), (6, "6"), (12, "12"), (18, "18")]:
    x, y = P(cx, cy, R + 30, i)
    anc = "middle" if i % 12 == 0 else ("start" if i == 6 else "end")
    o.append(f'<text x="{x}" y="{y + 6}" text-anchor="{anc}" {SANS} {DIM}>{lab}</text>')
for i in range(24):
    x, y = P(cx, cy, R, i)
    o.append(f'<circle cx="{x}" cy="{y}" r="6" fill="{"#5fd68a" if i % 6 == 0 else "#4a4a56"}"/>')
o.append(arc(cx, cy, R + 12, 23, 0, "#4ea1ff", 4))
o.append('</svg>')
out["panel-state"] = "\n".join(o)

# ---- one step, worked: H = {6}, arrival at 0, tilt at 10
W, H, cx, cy, R = 520, 500, 260, 250, 198
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
for h in (6, 18):
    x, y = P(cx, cy, R, h)
    o.append(f'<circle cx="{x}" cy="{y}" r="10" fill="{HOME}"/>')
for h, anc in ((6, "middle"), (18, "middle")):
    x, y = P(cx, cy, R, h)
    o.append(f'<text x="{x}" y="{y + 34}" text-anchor="{anc}" {SANS} fill="{HOME}">rests</text>')
for c in (9, 11):
    x, y = P(cx, cy, R, c)
    o.append(f'<circle cx="{x}" cy="{y}" r="9" fill="#2f2f37" stroke="#4ea1ff" stroke-width="3"/>')
    lx, ly = P(cx, cy, R + 30, c)
    o.append(f'<text x="{lx}" y="{ly + 14}" text-anchor="middle" {SANS} fill="#4ea1ff">{c}</text>')
o.append(arrow(cx, cy, R, 0, ARR, 3.4))
o.append(f'<text x="{cx}" y="30" text-anchor="middle" {SANS} fill="{ARR}">a = 0</text>')
o.append(arrow(cx, cy, R, 10, TOP, 4.2))
o.append(f'<text x="238" y="352" text-anchor="end" {SANS} fill="{TOP}">t = 10</text>')
o.append('</svg>')
out["panel-step"] = "\n".join(o)

# ---- what a report puts on the frame: the four angles it writes
W, H, cx, cy, R = 520, 460, 260, 230, 190
o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
o.append(arrow(cx, cy, R, 17, ARR, 3.4))
o.append(arrow(cx, cy, R, 15, BOT, 3))
o.append(arrow(cx, cy, R, 9, NORM, 3.6))
o.append(arrow(cx, cy, R, 3, TOP, 4.2))
o.append('</svg>')
out["panel-frame"] = "\n".join(o)

for k, v in out.items():
    # xmlns is REQUIRED for a standalone .svg loaded through <img>. Inline SVG in an
    # HTML document does not need it, which is why leaving it out looked fine until
    # these became separate files — then every one of them rendered as alt text.
    v = v.replace('<svg ', '<svg xmlns="http://www.w3.org/2000/svg" ', 1)
    open(k + ".svg", "w").write(v)
print(len(out), "panels:", " ".join(sorted(out)))
