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
    # |t - a|, not max-min: a comparison is a subtraction, so ordering the operands first
    # subtracts twice to avoid a sign check the sign bit already answers.
    # No \bmod: both indices are already 0..23, so d never exceeds 23 and the code takes
    # no modulus. The min is the arc choice, which is a real choice between two different
    # numbers — unlike the case split it replaced, which only looked like wrap handling.
    "length": r"""
  d = \lvert t - a \rvert \\[6pt]
  L \;=\; \len(t,a) = \min(d,\; 24 - d)
""",
    # machineForGap: the third way R changes, and the only one that reads the arrival.
    # both machines as (states, inputs, transition function)
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
    # The arithmetic the machine COMES OUT OF, not arithmetic imitating the machine.
    #
    # These four say NOTHING about the machine. No tilt, no partner, no arrival, no mode,
    # no rest — x and y are two numbers on a ring of 24, and the claim is a fact about
    # them: shift one by 6, and the only gaps whose length lands on 0, 6 or 12 are the four
    # quarter gaps. That fact is true whether or not anything runs. Writing it in the
    # machine's own words (t, p, a, "the rests") made it read as the machine restated,
    # which is exactly what this card exists to deny — the machine is what you get when
    # you build something that lives on these numbers, not the other way round.
    #
    # Also: no Z_24, no congruence, no min. Each is a NAME for something the ring says out
    # loud, and naming it makes four lines of counting look like it needs the machinery.
    #
    # FOUR FILES, not one. The first cut was a single block that ran each claim and its
    # gloss across one long line: 399pt wide against 157-322pt for every other block on
    # the page, so it was the one block max-width:100% shrank and the only one that did
    # not render at the size it was set at. Setting the same content as one TALL block
    # instead just moved the problem — it overflowed pairmath's 7cm paper and dvisvgm
    # cropped an empty first page to 0pt. The page's own pattern is one card per block,
    # each its own file at a size you can read; these are the four steps of the argument.
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
    # WHICH MODE TO BE IN, written as the conditional it is: measure something, then let the
    # measurement decide. The earlier spelling set R_after with a cases brace, which reads as
    # a definition of R_after — a thing that IS one of two values — and hides that a node
    # computes the partner's tilt and compares it before it knows which. Same rule, written
    # as the decision it makes rather than the value it ends up with.
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
    # WHICH SLOT TO GO TO. This was two blocks — "the two questions" and "the turn" — and
    # between them they stated the same measurement twice: once as a pair of yes/no questions
    # and again as a cases brace over the answers. One conditional says it once. The measure
    # step is the whole of what the mode contributes, and it contributes it as R, a list of
    # numbers, which is why neither branch below mentions perpendicular or parallel.
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
    # ---- closed-*: the arithmetic that REPLACES the machine (arith.html) ----------------
    #
    # These are not the machine written out. There is no current state, no transition, no
    # rule applied over and over — you are given three numbers and you write down where the
    # node ends up and how many arrivals it takes. The machine's own vocabulary (states,
    # links, next/prev, settle) does not appear because none of it is needed to say this.
    #
    # Every claim here is swept against the running code by TestTheWalkIsClosedForm.
    "closed-numbers": r"""
  \quad t,\, a \;\in\; \{0,\, 1,\, \ldots,\, 23\} \\[9pt]
  \quad \text{\color{acc}this node} \quad
    \text{tilt} = t \qquad \text{normal} = t+6 \qquad \text{bottom} = t+12 \\[4pt]
  \quad {\color{dim}\text{only } t \text{ moves; the other two are read off it}} \\[9pt]
  \quad \text{\color{acc}what arrives} \quad a = p + 6
    \quad {\color{dim}\text{the partner's normal, not its tilt}} \\[9pt]
  \quad \text{\color{acc}how far apart} \\[4pt]
  \quad \ell = \bigl\lvert\, (t - a + 12 \bmod 24) - 12 \,\bigr\rvert
    \quad {\color{dim}0 \le \ell \le 12} \\[4pt]
  \quad {\color{dim}\text{slots from } t \text{ to } a,\text{ the shorter way round}}
""",
    "closed-far": r"""
  \quad \text{\color{acc}how far off the quarter} \\[4pt]
  \quad q = \lvert\, \ell - 6 \,\rvert
    \quad {\color{dim}\text{slots from } \ell \text{ to } 6} \\[9pt]
  \quad \text{\color{acc}how far off the nearest resting angle} \\[4pt]
  \quad {\color{par}\text{parallel} : \; f = q} \qquad
        {\color{perp}\text{perpendicular} : \; f = 6 - q}
""",
    # ONE ROUND, with no case in it. The magnitude story above (l, q, f) has thrown the sign
    # away, so it needs a comparison to get direction back. Keep the SIGNED gap instead and
    # both arrangements are one line, differing by a shift of 6 inside the modulus.
    "closed-round": r"""
  \quad \text{\color{acc}which way, and how far, from } a \text{ to } t \\[4pt]
  \quad d = t - a
    \quad {\color{dim}-23 \le d \le 23, \text{ negative when } a > t} \\[9pt]
  \quad \text{\color{acc}which way, and how far, to the nearest resting angle} \\[4pt]
  \quad {\color{par}\text{parallel} : \; e = (d \bmod 12) - 6} \\[5pt]
  \quad {\color{perp}\text{perpendicular} : \; e = (d + 6 \bmod 12) - 6} \\[4pt]
  \quad {\color{dim}d \bmod 12 \;\in\; \{0, \ldots, 11\}, \text{ so } -6 \le e \le 5} \\[9pt]
  \quad {\color{acc}t_{\mathrm{after}} = t - \operatorname{sign}(e)}
    \qquad {\color{acc}f = \lvert e \rvert} \\[11pt]
  \quad {\color{dim}\ell = \lvert (d + 12 \bmod 24) - 12 \rvert} \\[4pt]
  \quad {\color{dim}\text{writing } \lvert d \rvert \text{ for } d
    \text{ leaves } f \text{ equal and flips } \operatorname{sign}(e)}
""",
    "closed-run": r"""
  \quad {\color{acc}\text{arrivals}} \;=\; f \\[7pt]
  \quad {\color{acc}\text{lands at}} \;=\; t - f\operatorname{sign}(e) \\[7pt]
  \quad {\color{dim}\operatorname{sign}(e) \text{ from the first arrival, and never again}}
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

# px per pt for the displayed width, shared by every block so blocks of different shapes
# render at the same glyph size. Set against the DRAWINGS, not chosen in the abstract:
# panels.py labels its SVGs at 15px and those SVGs are shown at their authored width, so
# 15px is the page's at-a-glance size. At 1.5 the maths was the smaller of the two sizes on
# the page — uniform among itself, but under the thing it sits beside.
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
