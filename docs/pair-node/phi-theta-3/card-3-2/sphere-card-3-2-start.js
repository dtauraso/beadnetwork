const SPHERE_CARD_3_2_START = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{one node} & & \\[3pt]
\rlap{\textbf{move start up to } m s \textbf{ in direction cw or ccw, stopping at end}} & & \\[3pt]
\begin{bmatrix} \text{gap_cw}_{\varphi} \\ \text{gap_cw}_{\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} (\text{end}_{\varphi} + 12s - \text{start}_{\varphi}) \bmod 12s \\ (\text{end}_{\theta} + 12s - \text{start}_{\theta}) \bmod 12s \end{array}\right] \\[6pt]
\begin{bmatrix} \text{gap_ccw}_{\varphi} \\ \text{gap_ccw}_{\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} (\text{start}_{\varphi} + 12s - \text{end}_{\varphi}) \bmod 12s \\ (\text{start}_{\theta} + 12s - \text{end}_{\theta}) \bmod 12s \end{array}\right] \\[6pt]
\begin{bmatrix} \text{gap}_{\varphi} \\ \text{gap}_{\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} \text{cw}\, \text{gap_cw}_{\varphi} + \text{ccw}\, \text{gap_ccw}_{\varphi} \\ \text{cw}\, \text{gap_cw}_{\theta} + \text{ccw}\, \text{gap_ccw}_{\theta} \end{array}\right] \\[6pt]
\begin{bmatrix} \text{step}_{\varphi} \\ \text{step}_{\theta} \end{bmatrix}
  &=& \left[\begin{array}{l} \min(m s,\, \text{gap}_{\varphi}) \\ \min(m s,\, \text{gap}_{\theta}) \end{array}\right] \\[6pt]
\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix}
  &=& \left[\begin{array}{ll} (\text{start}_{\varphi} + \text{cw}\, \text{step}_{\varphi} + \text{ccw}\, (12s - \text{step}_{\varphi})) \bmod 12s & \text{if } \text{cw} \oplus \text{ccw} \\[3pt]
                      \text{start}_{\varphi} & \text{otherwise} \\[6pt]
                      (\text{start}_{\theta} + \text{cw}\, \text{step}_{\theta} + \text{ccw}\, (12s - \text{step}_{\theta})) \bmod 12s & \text{if } \text{cw} \oplus \text{ccw} \\[3pt]
                      \text{start}_{\theta} & \text{otherwise} \end{array}\right]
\end{array}
\]`;
