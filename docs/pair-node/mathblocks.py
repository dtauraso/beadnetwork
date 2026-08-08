"""Build one .tex + .svg per block of the update maths.

Rows are set in array{@{}l@{}} — a single LEFT-aligned column with no padding
either side — so every line starts at the same left edge. The obvious choice,
`aligned`, anchors each row at its & instead, which lines the rows up on their
equals signs and leaves ragged space in front of the shorter left-hand sides.
"""
import os, subprocess

os.chdir(os.path.dirname(os.path.abspath(__file__)))

# one left-aligned column, no leading or trailing padding
OPEN = r"\[\begin{array}{@{}l@{}}"
CLOSE = r"\end{array}\]"

BODIES = {
    "read": r"""
  \mathrm{normal}(t) = t+6 \\[4pt]
  \mathrm{bottom}(t) = t+12 \\[4pt]
  \mathrm{sends} = \mathrm{normal}(t) \\[4pt]
  \mathrm{angle}(i) = \tfrac{\pi}{12}\, i
""",
    # max - min rather than |t - a|: the same number, but subtracted in the order that
    # cannot go negative, so it is a magnitude by construction instead of a signed
    # difference with its sign thrown away afterwards.
    # No \bmod either: both indices are already 0..23, so d never exceeds 23 and the code
    # takes no modulus. Writing one implies an operation that is not there.
    # min of the two arcs, not a case split. The cases read as wrap handling, which this
    # is not — the ring's 24 states already make wrapping impossible. What is being
    # chosen is WHICH OF THE TWO ARCS between the points is meant.
    "length": r"""
  d = \max(t,a) - \min(t,a) \\[6pt]
  L \;=\; \len(t,a) = \min(d,\; 24 - d)
""",
    # machineForGap: the third way R changes, and the only one that reads the arrival.
    "choose": r"""
  \text{\color{acc}when } R_{\mathrm{before}} = R_{\mathrm{setting}} : \\[7pt]
  \quad t_{\mathrm{partner}} = a_{-6} \\[5pt]
  \quad R_{\mathrm{after}} = \begin{cases}
      {\color{perp}R_{\perp}}    & \len(t_{\mathrm{before}},\, t_{\mathrm{partner}}) = 6 \\
      {\color{par}R_{\parallel}} & \text{otherwise}
    \end{cases}
""",
    "resting": r"""
  {\color{perp}R_{\perp} = \{0,\,12\}} \\[5pt]
  {\color{par}R_{\parallel} = \{6\}} \\[5pt]
  {\color{dim}R_{\mathrm{setting}} = \text{any}}
""",
    "questions": r"""
  \text{\color{acc}settle?} \quad L \in R \\[7pt]
  \text{\color{acc}up?} \quad \len(t_{+1},a) \text{ closer to } R \text{ than } \len(t_{-1},a)
""",
    "turn": r"""
  t_{\mathrm{after}} = \begin{cases}
      t_{\mathrm{before}} & \text{settle} \\
      t_{+1}              & \text{up} \\
      t_{-1}              & \text{otherwise}
    \end{cases}
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

SCALE = 1.5  # px per pt for the displayed width

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
