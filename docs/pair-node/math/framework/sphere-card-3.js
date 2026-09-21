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
\text{offset}_{\varphi},\, \text{offset}_{\theta} &\in& \{0, 1, 2, 3\} \\[6pt]
\rlap{\textbf{dir_down}(p_{n}, \text{offset}, qt_{n})} & & \\[3pt]
\begin{bmatrix} \text{pole}_{\varphi} \\ \text{pole}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} p_{n} + \text{offset}_{\varphi} \\[6pt]
                      p_{n} + \text{offset}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \Delta_{\varphi} \\ \Delta_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\varphi} - \text{pole}_{\varphi} \\[6pt]
                      \text{arrival}_{\theta} - \text{pole}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{direction}_{\varphi} \\ \text{direction}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } \text{pole}_{\varphi} < \Delta_{\varphi} < qt_{n} \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } \text{pole}_{\theta} < \Delta_{\theta} < qt_{n} \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\rlap{\textbf{dir_up}(p_{n}, \text{offset}, qt_{n})} & & \\[3pt]
\begin{bmatrix} \text{pole}_{\varphi} \\ \text{pole}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} p_{n} - \text{offset}_{\varphi} \\[6pt]
                      p_{n} - \text{offset}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \Delta_{\varphi} \\ \Delta_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{pole}_{\varphi} - \text{arrival}_{\varphi} \\[6pt]
                      \text{pole}_{\theta} - \text{arrival}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{direction}_{\varphi} \\ \text{direction}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1 & \text{if } \text{pole}_{\varphi} < \Delta_{\varphi} < qt_{n} \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      1 & \text{if } \text{pole}_{\theta} < \Delta_{\theta} < qt_{n} \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\textbf{one node} & & \\[3pt]
\begin{bmatrix} \text{arrival}_{1\,q_{a}\,\varphi} \\ \text{arrival}_{1\,q_{a}\,\theta} \end{bmatrix}
  &=& \text{from partner } 1 \\[6pt]
\begin{bmatrix} \text{arrival}_{2\,q_{b}\,\varphi} \\ \text{arrival}_{2\,q_{b}\,\theta} \end{bmatrix}
  &=& \text{from partner } 2 \\[6pt]
\begin{bmatrix} \text{arrival}_{\varphi} \\ \text{arrival}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{1\,q_{a}\,\varphi} \\[6pt]
                      \text{arrival}_{1\,q_{a}\,\theta} \end{bmatrix} \\[10pt]
\text{offset}_{k\,\varphi},\, \text{offset}_{k\,\theta} &\in& \{0, 1, 2, 3\}, \; k = 0 \ldots 3 \\[6pt]
\begin{bmatrix} \text{direction}_{0\,\varphi} \\ \text{direction}_{0\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}(0, \text{offset}_{0}, 3) \\[6pt]
\begin{bmatrix} \text{direction}_{1\,\varphi} \\ \text{direction}_{1\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}(6, \text{offset}_{1}, 3) \\[6pt]
\begin{bmatrix} \text{direction}_{2\,\varphi} \\ \text{direction}_{2\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}(6, \text{offset}_{2}, 9) \\[6pt]
\begin{bmatrix} \text{direction}_{3\,\varphi} \\ \text{direction}_{3\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}(12, \text{offset}_{3}, 9) \\[10pt]
\begin{bmatrix} \text{arrival}_{\text{next}_{\varphi}} \\ \text{arrival}_{\text{next}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\varphi} + \text{offset}_{0\,\varphi} \\
                      \text{arrival}_{\theta} + \text{offset}_{0\,\theta} \end{bmatrix} \\
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
