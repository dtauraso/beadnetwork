const SPHERE_CARD_3_2_START = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{each } k \in K_{b} \textbf{, on its own} & & \\[3pt]
\rlap{\textbf{move start}_{k} \textbf{ up to } m s \textbf{ in direction cw or ccw, stopping at end}_{k}} & & \\[3pt]
\begin{bmatrix} \text{gap_cw}_{k\,\varphi} \\ \text{gap_cw}_{k\,\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} (\text{end}_{k\,\varphi} + 12s - \text{start}_{k\,\varphi}) \bmod 12s \\ (\text{end}_{k\,\theta} + 12s - \text{start}_{k\,\theta}) \bmod 12s \end{array}\right] \\[6pt]
\begin{bmatrix} \text{gap_ccw}_{k\,\varphi} \\ \text{gap_ccw}_{k\,\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} (\text{start}_{k\,\varphi} + 12s - \text{end}_{k\,\varphi}) \bmod 12s \\ (\text{start}_{k\,\theta} + 12s - \text{end}_{k\,\theta}) \bmod 12s \end{array}\right] \\[6pt]
\begin{bmatrix} \text{gap}_{k\,\varphi} \\ \text{gap}_{k\,\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} \text{cw}\, \text{gap_cw}_{k\,\varphi} + \text{ccw}\, \text{gap_ccw}_{k\,\varphi} \\ \text{cw}\, \text{gap_cw}_{k\,\theta} + \text{ccw}\, \text{gap_ccw}_{k\,\theta} \end{array}\right] \\[6pt]
\begin{bmatrix} \text{step}_{k\,\varphi} \\ \text{step}_{k\,\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} \min(m s,\, \text{gap}_{k\,\varphi}) \\ \min(m s,\, \text{gap}_{k\,\theta}) \end{array}\right] \\[6pt]
\begin{bmatrix} \text{start}_{k\,\varphi} \\ \text{start}_{k\,\theta} \end{bmatrix}
  &=& \left[\begin{array}{ll} (\text{start}_{k\,\varphi} + \text{cw}\, \text{step}_{k\,\varphi} + \text{ccw}\, (12s - \text{step}_{k\,\varphi})) \bmod 12s & \text{if } \text{cw} \oplus \text{ccw} \\[3pt]
                      \text{start}_{k\,\varphi} & \text{otherwise} \\[6pt]
                      (\text{start}_{k\,\theta} + \text{cw}\, \text{step}_{k\,\theta} + \text{ccw}\, (12s - \text{step}_{k\,\theta})) \bmod 12s & \text{if } \text{cw} \oplus \text{ccw} \\[3pt]
                      \text{start}_{k\,\theta} & \text{otherwise} \end{array}\right] \\[6pt]
& & \text{a move of start}_{k} \text{ is the pair card's "b changes start}_{a \to b}\text{" with } k = a \to b
\end{array}
\]`;
