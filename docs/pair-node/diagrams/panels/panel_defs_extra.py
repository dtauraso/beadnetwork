from panel_primitives import TOP, NORM, BOT, ARR, HOME, CAND, MONO, SANS, INK, DIM, P, arrow, ring, arc, straight


def build_panel_links():
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
    return "\n".join(o)


def build_panel_modes():
    W, H = 620, 250
    o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">']
    boxes = [(24, 102, 140, 46, "setting", "#9a9aa6"),
             (400, 30, 190, 46, "perpendicular", HOME),
             (400, 174, 190, 46, "parallel", ARR)]
    for x, y, w, h, name, col in boxes:
        o.append(f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="9" fill="#2b2b33" stroke="{col}"/>')
        o.append(f'<text x="{x + w // 2}" y="{y + 29}" text-anchor="middle" {MONO} fill="{col}">{name}</text>')

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
    return "\n".join(o)


def build_panel_state():
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
    return "\n".join(o)


def build_panel_step():
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
    return "\n".join(o)


def build_panel_frame():
    W, H, cx, cy, R = 260, 250, 130, 125, 104
    o = [f'<svg width="{W}" height="{H}" viewBox="0 0 {W} {H}">', ring(cx, cy, R)]
    o.append(arrow(cx, cy, R, 17, ARR, 2.4))
    o.append(arrow(cx, cy, R, 15, BOT, 2.2))
    o.append(arrow(cx, cy, R, 9, NORM, 2.6))
    o.append(arrow(cx, cy, R, 3, TOP, 3))
    o.append('</svg>')
    return "\n".join(o)
