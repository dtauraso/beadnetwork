const SPHERE_CARD_3_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_3_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
q_{j} &=& \text{quadrant } j \text{, clockwise, } j = 1 \ldots 4 \\[3pt]
q &=& \text{the quadrant number} \\[3pt]
a &=& \text{arrival 1's quadrant} \\[3pt]
b &=& \text{arrival 2's quadrant} \\[3pt]
12 &=& \text{1 full turn } \varphi \\[3pt]
12 &=& \text{1 full turn } \theta \\[3pt]
3 &=& \text{1 quarter turn } \varphi \\[3pt]
3 &=& \text{1 quarter turn } \theta \\[10pt]
\textbf{node 1} & & \\[3pt]
\begin{bmatrix} n_{1}q_{1}r_{1\,\varphi} & n_{1}q_{1}r_{2\,\varphi} \\ n_{1}q_{1}r_{1\,\theta} & n_{1}q_{1}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{1}q_{2}r_{1\,\varphi} & n_{1}q_{2}r_{2\,\varphi} \\ n_{1}q_{2}r_{1\,\theta} & n_{1}q_{2}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{1}q_{3}r_{1\,\varphi} & n_{1}q_{3}r_{2\,\varphi} \\ n_{1}q_{3}r_{1\,\theta} & n_{1}q_{3}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{1}q_{4}r_{1\,\varphi} & n_{1}q_{4}r_{2\,\varphi} \\ n_{1}q_{4}r_{1\,\theta} & n_{1}q_{4}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{1}\text{start}_{1}q_{1\,\varphi} \\ n_{1}\text{start}_{1}q_{1\,\theta} \\ n_{1}\text{start}_{1}q_{1\,q} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \\ 1 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{1}\text{start}_{2}q_{2\,\varphi} \\ n_{1}\text{start}_{2}q_{2\,\theta} \\ n_{1}\text{start}_{2}q_{2\,q} \end{bmatrix}
  &=& \begin{bmatrix} 3 \\ 3 \\ 2 \end{bmatrix} \\[10pt]
\textbf{node 2} & & \\[3pt]
\begin{bmatrix} n_{2}q_{1}r_{1\,\varphi} & n_{2}q_{1}r_{2\,\varphi} \\ n_{2}q_{1}r_{1\,\theta} & n_{2}q_{1}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{2}q_{2}r_{1\,\varphi} & n_{2}q_{2}r_{2\,\varphi} \\ n_{2}q_{2}r_{1\,\theta} & n_{2}q_{2}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{2}q_{3}r_{1\,\varphi} & n_{2}q_{3}r_{2\,\varphi} \\ n_{2}q_{3}r_{1\,\theta} & n_{2}q_{3}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{2}q_{4}r_{1\,\varphi} & n_{2}q_{4}r_{2\,\varphi} \\ n_{2}q_{4}r_{1\,\theta} & n_{2}q_{4}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{2}\text{start}_{1}q_{4\,\varphi} \\ n_{2}\text{start}_{1}q_{4\,\theta} \\ n_{2}\text{start}_{1}q_{4\,q} \end{bmatrix}
  &=& \begin{bmatrix} 3 \\ 3 \\ 4 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{2}\text{start}_{2}q_{1\,\varphi} \\ n_{2}\text{start}_{2}q_{1\,\theta} \\ n_{2}\text{start}_{2}q_{1\,q} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \\ 1 \end{bmatrix} \\[10pt]
\textbf{node 3} & & \\[3pt]
\begin{bmatrix} n_{3}q_{1}r_{1\,\varphi} & n_{3}q_{1}r_{2\,\varphi} \\ n_{3}q_{1}r_{1\,\theta} & n_{3}q_{1}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{3}q_{2}r_{1\,\varphi} & n_{3}q_{2}r_{2\,\varphi} \\ n_{3}q_{2}r_{1\,\theta} & n_{3}q_{2}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{3}q_{3}r_{1\,\varphi} & n_{3}q_{3}r_{2\,\varphi} \\ n_{3}q_{3}r_{1\,\theta} & n_{3}q_{3}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{3}q_{4}r_{1\,\varphi} & n_{3}q_{4}r_{2\,\varphi} \\ n_{3}q_{4}r_{1\,\theta} & n_{3}q_{4}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & 0 \\ 0 & 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{3}\text{start}_{1}q_{3\,\varphi} \\ n_{3}\text{start}_{1}q_{3\,\theta} \\ n_{3}\text{start}_{1}q_{3\,q} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \\ 3 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{3}\text{start}_{2}q_{3\,\varphi} \\ n_{3}\text{start}_{2}q_{3\,\theta} \\ n_{3}\text{start}_{2}q_{3\,q} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \\ 3 \end{bmatrix} \\[10pt]
\textbf{update test} & & \\[3pt]
\text{offset}_{0\,\varphi},\, \text{offset}_{0\,\theta} &\in& \{0, 1, 2, 3\} \\[6pt]
\begin{bmatrix} \text{pole}_{0\,\varphi} \\ \text{pole}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 + \text{offset}_{0\,\varphi} \\[6pt]
                      0 + \text{offset}_{0\,\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \Delta_{0\,\varphi} \\ \Delta_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\varphi} - \text{pole}_{0\,\varphi} \\[6pt]
                      \text{arrival}_{\theta} - \text{pole}_{0\,\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{direction}_{0\,\varphi} \\ \text{direction}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } \text{pole}_{0\,\varphi} < \Delta_{0\,\varphi} < 3 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } \text{pole}_{0\,\theta} < \Delta_{0\,\theta} < 3 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\textbf{one node} & & \\[3pt]
\begin{bmatrix} \text{arrival}_{1\,q_{a}\,\varphi} \\ \text{arrival}_{1\,q_{a}\,\theta} \end{bmatrix}
  &=& \text{from partner } 1 \\[6pt]
\begin{bmatrix} \text{arrival}_{2\,q_{b}\,\varphi} \\ \text{arrival}_{2\,q_{b}\,\theta} \end{bmatrix}
  &=& \text{from partner } 2 \\[6pt]
\begin{bmatrix} \text{arrival}_{\varphi} \\ \text{arrival}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{1\,q_{a}\,\varphi} \\[6pt]
                      \text{arrival}_{1\,q_{a}\,\theta} \end{bmatrix} \\[10pt]
\Delta_{0\,\varphi},\, \Delta_{0\,\theta} &\in& \{0, 1, 2, 3\} \\[6pt]
\begin{bmatrix} p_{0}r_{0\,\varphi} \\ p_{0}r_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\varphi} - (0 + \Delta_{0\,\varphi}) \\[6pt]
                      \text{arrival}_{\theta} - (0 + \Delta_{0\,\theta}) \end{bmatrix} \\[6pt]
\begin{bmatrix} x_{0\,\varphi} \\ x_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } 0 + \Delta_{0\,\varphi} < p_{0}r_{0\,\varphi} < 3 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } 0 + \Delta_{0\,\theta} < p_{0}r_{0\,\theta} < 3 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\Delta_{1\,\varphi},\, \Delta_{1\,\theta} &\in& \{0, 1, 2, 3\} \\[6pt]
\begin{bmatrix} p_{1}r_{1\,\varphi} \\ p_{1}r_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} (6 - \Delta_{1\,\varphi}) - \text{arrival}_{\varphi} \\[6pt]
                      (6 - \Delta_{1\,\theta}) - \text{arrival}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} x_{1\,\varphi} \\ x_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1 & \text{if } 6 - \Delta_{1\,\varphi} < p_{1}r_{1\,\varphi} < 3 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      1 & \text{if } 6 - \Delta_{1\,\theta} < p_{1}r_{1\,\theta} < 3 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\Delta_{2\,\varphi},\, \Delta_{2\,\theta} &\in& \{0, 1, 2, 3\} \\[6pt]
\begin{bmatrix} p_{2}r_{2\,\varphi} \\ p_{2}r_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\varphi} - (6 + \Delta_{2\,\varphi}) \\[6pt]
                      \text{arrival}_{\theta} - (6 + \Delta_{2\,\theta}) \end{bmatrix} \\[6pt]
\begin{bmatrix} x_{2\,\varphi} \\ x_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } 6 + \Delta_{2\,\varphi} < p_{2}r_{2\,\varphi} < 9 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } 6 + \Delta_{2\,\theta} < p_{2}r_{2\,\theta} < 9 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\Delta_{3\,\varphi},\, \Delta_{3\,\theta} &\in& \{0, 1, 2, 3\} \\[6pt]
\begin{bmatrix} p_{3}r_{3\,\varphi} \\ p_{3}r_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} (12 - \Delta_{3\,\varphi}) - \text{arrival}_{\varphi} \\[6pt]
                      (12 - \Delta_{3\,\theta}) - \text{arrival}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} x_{3\,\varphi} \\ x_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1 & \text{if } 12 - \Delta_{3\,\varphi} < p_{3}r_{3\,\varphi} < 9 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      1 & \text{if } 12 - \Delta_{3\,\theta} < p_{3}r_{3\,\theta} < 9 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{arrival}_{\text{next}_{\varphi}} \\ \text{arrival}_{\text{next}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\varphi} + \Delta_{0\,\varphi} \\
                      \text{arrival}_{\theta} + \Delta_{0\,\theta} \end{bmatrix} \\
\begin{bmatrix} \text{sent}^{1}_{\varphi} \\ \text{sent}^{1}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\text{next}_{\varphi}} \\
                      \text{arrival}_{\text{next}_{\theta}} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{sent}^{2}_{\varphi} \\ \text{sent}^{2}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\text{next}_{\varphi}} \\
                      \text{arrival}_{\text{next}_{\theta}} \end{bmatrix}
\end{array}
\]`;

function sphereCard3Figure(spec) {
  const fig = document.createElement('div');
  fig.className = 'keysphere';
  const svgEl = sphere(spec);
  fig.appendChild(svgEl);
  fig.appendChild(sphereCardToggles(svgEl));
  fig.appendChild(sphereEndControls(svgEl));
  return fig;
}

function sphereCard3Math(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3]')) {
  host.appendChild(sphereCard3Figure(SPHERE_CARD_3_SPEC));
  host.appendChild(sphereCard3Math(SPHERE_CARD_3_FORMULAS));
}
