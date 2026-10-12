import os, subprocess

os.chdir(os.path.dirname(os.path.abspath(__file__)))

LEFT_ALIGNED_NO_PADDING_ARRAY_OPEN = r"\[\begin{array}{@{}l@{}}"
LEFT_ALIGNED_NO_PADDING_ARRAY_CLOSE = r"\end{array}\]"

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
  S = \{\, {\color{dim}R_{\mathrm{setting}}},\; {\color{par}R_{\parallel}} \,\}
  \qquad \Sigma = \{\, a,\; \mathrm{reset} \,\} \\[6pt]
  \delta(R_{\mathrm{setting}},\, a) = {\color{par}R_{\parallel}} \\[6pt]
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
  \quad {\color{dim}g = 6 \text{ or } 18} \\[3pt]
  \qquad {\color{dim}\text{a quarter of the ring between them}}
""",
    "arith-rests": r"""
  \text{\color{acc}shift one of the two by 6, and only those four land on }
    0,\,6,\,12 \\[6pt]
  \quad {\color{par}g = 0 \text{ or } 12 \;\Rightarrow\; L(x,\,y+6) = 6} \\[5pt]
  \quad {\color{dim}g = 6 \text{ or } 18 \;\Rightarrow\; L(x,\,y+6) = 0 \text{ or } 12}
    \\[5pt]
  \quad {\color{dim}\text{every other } g \;\Rightarrow\; L(x,\,y+6)
    \text{ is none of } 0,\,6,\,12}
""",

    "resting": r"""
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

FAMILY_ENV = dict(os.environ, TEXINPUTS=".." + os.pathsep + os.environ.get("TEXINPUTS", ""))


def family_and_stem(name):
    family, _, rest = name.partition("-")
    return (family, rest) if rest else (name, name)


for name, body in BODIES.items():
    family, stem = family_and_stem(name)
    os.makedirs(family, exist_ok=True)
    tex_name = stem + ".tex"
    content = (HEAD % name) + LEFT_ALIGNED_NO_PADDING_ARRAY_OPEN + body + \
        LEFT_ALIGNED_NO_PADDING_ARRAY_CLOSE + "\n\\end{document}\n"
    open(os.path.join(family, tex_name), "w").write(content)
    subprocess.run(["latex", "-interaction=nonstopmode", tex_name], cwd=family,
                    env=FAMILY_ENV, capture_output=True)
    r = subprocess.run(["dvisvgm", "--no-fonts", "--exact-bbox",
                        f"--output={stem}.svg", stem + ".dvi"], cwd=family,
                        capture_output=True, text=True)
    out = r.stderr + r.stdout
    line = [l for l in out.splitlines() if "graphic size" in l]
    if not line:
        print(name, "FAILED")
        print(out[-600:])
        continue
    pt = float(line[0].split()[2].removesuffix("pt"))
    print(f'{family}/{stem}.svg  {pt:7.1f}pt  ->  style="width:{round(pt * SCALE)}px"')
