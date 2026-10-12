const SPHERE_CARD_3_2_START = String.raw`\textbf{one node} & & \\[3pt]
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
                      \text{start}_{\theta} & \text{otherwise} \end{array}\right] \\[10pt]
\begin{bmatrix} \text{start_direction}_{0\,\varphi} \\ \text{start_direction}_{0\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 0,\; 3s\right) \\[6pt]
\begin{bmatrix} \text{start_direction}_{1\,\varphi} \\ \text{start_direction}_{1\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 3s,\; 6s\right) \\[6pt]
\begin{bmatrix} \text{start_direction}_{2\,\varphi} \\ \text{start_direction}_{2\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 6s,\; 9s\right) \\[6pt]
\begin{bmatrix} \text{start_direction}_{3\,\varphi} \\ \text{start_direction}_{3\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 9s,\; 12s\right) \\[10pt]
\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix}
      \begin{array}{@{}r@{\;}l@{}}
        & \text{start}_{\varphi} \\
        + & \text{start_direction}_{0\,\varphi} \\
        + & \text{start_direction}_{1\,\varphi} \\
        + & \text{start_direction}_{2\,\varphi} \\
        + & \text{start_direction}_{3\,\varphi}
      \end{array} \\[10pt]
      \begin{array}{@{}r@{\;}l@{}}
        & \text{start}_{\theta} \\
        + & \text{start_direction}_{0\,\theta} \\
        + & \text{start_direction}_{1\,\theta} \\
        + & \text{start_direction}_{2\,\theta} \\
        + & \text{start_direction}_{3\,\theta}
      \end{array}
      \end{bmatrix}`;
