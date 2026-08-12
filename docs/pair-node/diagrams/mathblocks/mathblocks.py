"""Build one .tex + .svg per block of the update maths.

Rows are set in array{@{}l@{}} — a single LEFT-aligned column with no padding
either side — so every line starts at the same left edge. The obvious choice,
`aligned`, anchors each row at its & instead, which lines the rows up on their
equals signs and leaves ragged space in front of the shorter left-hand sides.
"""
import os, subprocess

os.chdir(os.path.dirname(os.path.abspath(__file__)))


OPEN = r"\[\begin{array}{@{}l@{}}"
CLOSE = r"\end{array}\]"

BODIES = {
    "read": r"""
  \mathrm{normal}(t) = t+6 \\[4pt]
  \mathrm{bottom}(t) = t+12 \\[4pt]
  \mathrm{sends} = \mathrm{normal}(t) \\[4pt]
  \mathrm{angle}(i) = \tfrac{\pi}{12}\, i
""",





    "length": r"""
  d = \lvert t - a \rvert \\[6pt]
  L \;=\; \len(t,a) = \min(d,\; 24 - d)
""",


    "machine-tilt": r"""
  S = \mathbb{Z}_{24} \qquad \Sigma = \{\,\mathrm{next},\; \mathrm{prev}\,\} \\[6pt]
  \delta(t,\, \mathrm{next}) = t + 1 \pmod{24} \\[4pt]
  \delta(t,\, \mathrm{prev}) = t - 1 \pmod{24}
""",
    "machine-mode": r"""
  S = \{\, {\color{dim}R_{\mathrm{setting}}},\; {\color{perp}R_{\perp}},\;
        {\color{par}R_{\parallel}} \,\}
  \qquad \Sigma = \{\, a,\; \mathrm{reset} \,\} \\[6pt]
  \delta(R_{\mathrm{setting}},\, a) = \begin{cases}
      {\color{perp}R_{\perp}}    & \len(t,\, a_{-6}) = 6 \\
      {\color{par}R_{\parallel}} & \text{otherwise}
    \end{cases} \\[6pt]
  \delta(R,\, a) = R \qquad R \neq R_{\mathrm{setting}} \\[4pt]
  \delta(R,\, \mathrm{reset}) = R_{\mathrm{setting}}
""",




















    "arith-groups": r"""
  \text{\color{acc}the ring} \quad x = 0,\,1,\,\ldots,\,23 \\[3pt]
  \quad {\color{dim}\text{counted round, so } 24 \text{ is } 0 \text{ again}} \\[8pt]
  \text{\color{acc}the half turn} \quad x \text{ and } x+12 \text{ name the same line} \\[3pt]
  \quad {\color{dim}\text{so a line is a number counted round twelve}}
""",
    "arith-distance": r"""
  \text{\color{acc}the length between two numbers} \\[4pt]
  \quad d = \lvert x - y \rvert \\[5pt]
  \quad L(x,y) = d \text{ or } 24 - d, \text{ whichever is shorter} \\[3pt]
  \quad {\color{dim}\text{never more than } 12}
""",
    "arith-gap": r"""
  \text{\color{acc}the four quarter gaps} \quad g = x - y \\[6pt]
  \quad {\color{par}g = 0 \text{ or } 12} \\[3pt]
  \qquad {\color{dim}x \text{ and } y \text{ name the same line}} \\[6pt]
  \quad {\color{perp}g = 6 \text{ or } 18} \\[3pt]
  \qquad {\color{dim}\text{a quarter of the ring between them}}
""",
    "arith-rests": r"""
  \text{\color{acc}shift one of the two by 6, and only those four land on }
    0,\,6,\,12 \\[6pt]
  \quad {\color{par}g = 0 \text{ or } 12 \;\Rightarrow\; L(x,\,y+6) = 6} \\[5pt]
  \quad {\color{perp}g = 6 \text{ or } 18 \;\Rightarrow\; L(x,\,y+6) = 0 \text{ or } 12}
    \\[5pt]
  \quad {\color{dim}\text{every other } g \;\Rightarrow\; L(x,\,y+6)
    \text{ is none of } 0,\,6,\,12}
""",





    "decide-mode": r"""
  \text{\color{acc}if } R \neq R_{\mathrm{setting}} :
    \quad {\color{dim}\text{nothing changes — the choice already stuck}} \\[9pt]
  \text{\color{acc}otherwise, measure} \\[5pt]
  \quad p = a_{-6}
    \quad {\color{dim}\text{the partner's tilt, backed out of the arrival}} \\[5pt]
  \quad G = \len(t,\, p)
    \quad {\color{dim}\text{the gap between the two tilts}} \\[9pt]
  \text{\color{acc}then decide} \\[5pt]
  \quad \text{if } G = 6 : \quad {\color{perp}R \leftarrow R_{\perp}} \\[5pt]
  \quad \text{else} : \qquad\;\; {\color{par}R \leftarrow R_{\parallel}}
""",
    "resting": r"""
  {\color{perp}R_{\perp} = \{0,\,12\}} \\[5pt]
  {\color{par}R_{\parallel} = \{6\}} \\[5pt]
  {\color{dim}R_{\mathrm{setting}} = \text{any}}
""",





    "decide-turn": r"""
  \text{\color{acc}measure}
    \quad {\color{dim}\text{where it is, and each way it could turn}} \\[5pt]
  \quad \ell = \len(t,\, a) \\[4pt]
  \quad \ell_{+} = \len(t_{+1},\, a) \\[4pt]
  \quad \ell_{-} = \len(t_{-1},\, a) \\[9pt]
  \text{\color{acc}how far off a rest}
    \quad {\color{dim}f_{+},\, f_{-} \text{ from } \ell_{+},\, \ell_{-} \text{ the same way}}
    \\[5pt]
  \quad {\color{par}\text{parallel} : \; f = \lvert \ell - 6 \rvert} \\[4pt]
  \quad {\color{perp}\text{perpendicular} : \; f = 6 - \lvert \ell - 6 \rvert} \\[4pt]
  \quad {\color{dim}\text{setting} : \; f = 0} \\[11pt]
  \text{\color{acc}then decide} \\[5pt]
  \quad f = 0 : \qquad\quad\; t_{\mathrm{after}} = t \\[5pt]
  \quad f_{+} \le f_{-} : \quad t_{\mathrm{after}} = t_{+1} \\[5pt]
  \quad \text{else} : \qquad\quad\;\, t_{\mathrm{after}} = t_{-1}
""",







    "other": r"""
  \text{every arrival} :\; r_{\mathrm{after}} = a \\[5pt]
  \text{panel } \blacktriangle\,\blacktriangledown :\;
    t_{\mathrm{after}} = t_{\pm 1}, \quad R_{\mathrm{after}} = R_{\mathrm{before}} \\[5pt]
  \text{reset} :\; t_{\mathrm{after}} = 0, \quad R_{\mathrm{after}} = R_{\mathrm{setting}}
""",
}

HEAD = ("%% update-%s.tex — one block of the update. Shared setup is pairmath.sty,\n"
        "%% which also carries the rebuild command.\n"
        "\\documentclass[12pt]{article}\n"
        "\\usepackage{pairmath}\n"
        "\\begin{document}\n")

















SCALE = 1.9

for name, body in BODIES.items():
    stem = f"update-{name}"
    open(stem + ".tex", "w").write((HEAD % name) + OPEN + body + CLOSE + "\n\\end{document}\n")
    subprocess.run(["latex", "-interaction=nonstopmode", stem + ".tex"], capture_output=True)
    r = subprocess.run(["dvisvgm", "--no-fonts", "--exact-bbox",
                        f"--output={stem}.svg", stem + ".dvi"], capture_output=True, text=True)
    out = r.stderr + r.stdout
    line = [l for l in out.splitlines() if "graphic size" in l]
    if not line:
        print(name, "FAILED")
        print(out[-600:])
        continue
    pt = float(line[0].split()[2].removesuffix("pt"))
    print(f'{stem}.svg  {pt:7.1f}pt  ->  style="width:{round(pt * SCALE)}px"')
