const SPHERE_CARD_3_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_3_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
p_{1} &=& \text{the top pole} \\[3pt]
p_{2} &=& \text{the bottom pole} \\[3pt]
q &=& \text{the quadrant number} \\[3pt]
12 &=& \text{1 full turn } \varphi \\[3pt]
12 &=& \text{1 full turn } \theta \\[3pt]
3 &=& \text{1 quarter turn } \varphi \\[3pt]
3 &=& \text{1 quarter turn } \theta \\[6pt]
\begin{bmatrix} p_{1\,\varphi} \\ p_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} p_{2\,\varphi} \\ p_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} p_{3\,\varphi} \\ p_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} p_{4\,\varphi} \\ p_{4\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{arrival start}_{1\,1\,\varphi} \\ \text{arrival start}_{1\,1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{arrival start}_{1\,2\,\varphi} \\ \text{arrival start}_{1\,2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{arrival start}_{2\,4\,\varphi} \\ \text{arrival start}_{2\,4\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3 \\ 3 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{arrival start}_{3\,3\,\varphi} \\ \text{arrival start}_{3\,3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \end{bmatrix} \\[10pt]
\textbf{one node} & & \\[3pt]
\begin{bmatrix} \text{arrival}_{1\,q_{1}\,\varphi} \\ \text{arrival}_{1\,q_{1}\,\theta} \end{bmatrix}
  &=& \text{from partner } 1 \\[6pt]
\begin{bmatrix} \text{arrival}_{2\,q_{2}\,\varphi} \\ \text{arrival}_{2\,q_{2}\,\theta} \end{bmatrix}
  &=& \text{from partner } 2 \\[6pt]
\begin{bmatrix} \text{arrival}_{\varphi} \\ \text{arrival}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{1\,q_{1}\,\varphi} + \text{arrival}_{2\,q_{2}\,\varphi} \\[6pt]
                      \text{arrival}_{1\,q_{1}\,\theta} + \text{arrival}_{2\,q_{2}\,\theta} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{lean}_{\varphi} \\ \text{lean}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{1\,q_{1}\,\varphi} - \text{arrival}_{2\,q_{2}\,\varphi} \\[6pt]
                      \text{arrival}_{1\,q_{1}\,\theta} - \text{arrival}_{2\,q_{2}\,\theta} \end{bmatrix} \\[10pt]
\begin{bmatrix} \Delta_{p_1\,\varphi} \\ \Delta_{p_1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} |\, p_{1\,\varphi} - \text{arrival}_{\varphi} \,| \\[6pt]
                      |\, p_{1\,\theta} - \text{arrival}_{\theta} \,| \end{bmatrix} \\[6pt]
\begin{bmatrix} \Delta_{p_2\,\varphi} \\ \Delta_{p_2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} |\, p_{2\,\varphi} - \text{arrival}_{\varphi} \,| \\[6pt]
                      |\, p_{2\,\theta} - \text{arrival}_{\theta} \,| \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{zero}_{p_1\,\varphi} \\ \text{zero}_{p_1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & \text{if } \Delta_{p_1\,\varphi} = 0 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      0 & \text{if } \Delta_{p_1\,\theta} = 0 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{zero}_{p_2\,\varphi} \\ \text{zero}_{p_2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & \text{if } \Delta_{p_2\,\varphi} = 0 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      0 & \text{if } \Delta_{p_2\,\theta} = 0 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{acute}_{p_1\,\varphi} \\ \text{acute}_{p_1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } \Delta_{p_1\,\varphi} < 3 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } \Delta_{p_1\,\theta} < 3 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{acute}_{p_2\,\varphi} \\ \text{acute}_{p_2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } \Delta_{p_2\,\varphi} < 3 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } \Delta_{p_2\,\theta} < 3 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{offset}_{\varphi} \\ \text{offset}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \begin{array}{@{}l@{}} \text{zero}_{p_1\,\varphi} + \text{zero}_{p_2\,\varphi} \\ {} + \text{acute}_{p_1\,\varphi} + \text{acute}_{p_2\,\varphi} \end{array} \\[10pt]
                      \begin{array}{@{}l@{}} \text{zero}_{p_1\,\theta} + \text{zero}_{p_2\,\theta} \\ {} + \text{acute}_{p_1\,\theta} + \text{acute}_{p_2\,\theta} \end{array} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{arrival}_{\text{next}_{\varphi}} \\ \text{arrival}_{\text{next}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}_{\varphi} + \text{offset}_{\varphi} \\
                      \text{arrival}_{\theta} + \text{offset}_{\theta} \end{bmatrix} \\
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
