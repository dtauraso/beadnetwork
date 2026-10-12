import math

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


def straight(x1, y1, x2, y2, col, dash=""):
    n = ((x2 - x1) ** 2 + (y2 - y1) ** 2) ** 0.5
    ux, uy = (x2 - x1) / n, (y2 - y1) / n
    bx, by = x2 - 11 * ux, y2 - 11 * uy
    return (f'<line x1="{x1}" y1="{y1}" x2="{round(bx)}" y2="{round(by)}" stroke="{col}" '
            f'stroke-width="2"{dash}/>'
            f'<polygon points="{x2},{y2} {round(bx - 5 * uy)},{round(by + 5 * ux)} '
            f'{round(bx + 5 * uy)},{round(by - 5 * ux)}" fill="{col}"/>')
