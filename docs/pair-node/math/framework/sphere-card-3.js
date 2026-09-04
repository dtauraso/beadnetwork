const SPHERE_CARD_3_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_3_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
\text{center} &=& (\text{center}_{\varphi},\, \text{center}_{\theta},\, \text{center}_{r}) \\[6pt]
\begin{bmatrix} \text{center}^{1}_{\varphi} \\ \text{center}^{1}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{center}^{2}_{\varphi} \\ \text{center}^{2}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 10 \\ 10 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{center}^{3}_{\varphi} \\ \text{center}^{3}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 6 \\ 6 \end{bmatrix} \\[10pt]
p_{0} &=& \text{the top pole} \\[3pt]
p_{1} &=& \text{the bottom pole} \\[3pt]
12 &=& \text{1 full turn } \varphi \\[3pt]
12 &=& \text{1 full turn } \theta \\[3pt]
4 &=& \text{1 quarter turn } \varphi \\[3pt]
4 &=& \text{1 quarter turn } \theta \\[10pt]
\textbf{1} & & \\[3pt]
\begin{bmatrix} p^{1}_{0\,\varphi} \\ p^{1}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \end{bmatrix} \\[6pt]
\begin{bmatrix} p^{1}_{1\,\varphi} \\ p^{1}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3 \\ 3 \end{bmatrix} \\[10pt]
\textbf{2} & & \\[3pt]
\begin{bmatrix} p^{2}_{0\,\varphi} \\ p^{2}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 9 \\ 9 \end{bmatrix} \\[6pt]
\begin{bmatrix} p^{2}_{1\,\varphi} \\ p^{2}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 11 \\ 11 \end{bmatrix} \\[10pt]
\textbf{3} & & \\[3pt]
\begin{bmatrix} p^{3}_{0\,\varphi} \\ p^{3}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 5 \\ 5 \end{bmatrix} \\[6pt]
\begin{bmatrix} p^{3}_{1\,\varphi} \\ p^{3}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 7 \\ 7 \end{bmatrix} \\[10pt]
\textbf{one node} & & \\[3pt]
\begin{bmatrix} \text{arrival}^{1}_{\varphi} \\ \text{arrival}^{1}_{\theta} \end{bmatrix}
  &=& \text{from partner } 1 \\[6pt]
\begin{bmatrix} \text{arrival}^{2}_{\varphi} \\ \text{arrival}^{2}_{\theta} \end{bmatrix}
  &=& \text{from partner } 2 \\[6pt]
\begin{bmatrix} \text{arrival}_{\varphi} \\ \text{arrival}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{arrival}^{1}_{\varphi} + \text{arrival}^{2}_{\varphi} \\[6pt]
                      \text{arrival}^{1}_{\theta} + \text{arrival}^{2}_{\theta} \end{bmatrix} \\[10pt]
\begin{bmatrix} \Delta_{p_0\,\varphi} \\ \Delta_{p_0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} |\, p_{0\,\varphi} - \text{arrival}_{\varphi} \,| \\[6pt]
                      |\, p_{0\,\theta} - \text{arrival}_{\theta} \,| \end{bmatrix} \\[6pt]
\begin{bmatrix} \Delta_{p_1\,\varphi} \\ \Delta_{p_1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} |\, p_{1\,\varphi} - \text{arrival}_{\varphi} \,| \\[6pt]
                      |\, p_{1\,\theta} - \text{arrival}_{\theta} \,| \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{zero}_{p_0\,\varphi} \\ \text{zero}_{p_0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & \text{if } \Delta_{p_0\,\varphi} = 0 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      0 & \text{if } \Delta_{p_0\,\theta} = 0 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{zero}_{p_1\,\varphi} \\ \text{zero}_{p_1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 & \text{if } \Delta_{p_1\,\varphi} = 0 \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      0 & \text{if } \Delta_{p_1\,\theta} = 0 \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{acute}_{p_0\,\varphi} \\ \text{acute}_{p_0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } \Delta_{p_0\,\varphi} < \tau_{\varphi} \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } \Delta_{p_0\,\theta} < \tau_{\theta} \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{acute}_{p_1\,\varphi} \\ \text{acute}_{p_1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } \Delta_{p_1\,\varphi} < \tau_{\varphi} \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } \Delta_{p_1\,\theta} < \tau_{\theta} \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{offset}_{\varphi} \\ \text{offset}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \begin{array}{@{}l@{}} \text{zero}_{p_0\,\varphi} + \text{zero}_{p_1\,\varphi} \\ {} + \text{acute}_{p_0\,\varphi} + \text{acute}_{p_1\,\varphi} \end{array} \\[10pt]
                      \begin{array}{@{}l@{}} \text{zero}_{p_0\,\theta} + \text{zero}_{p_1\,\theta} \\ {} + \text{acute}_{p_0\,\theta} + \text{acute}_{p_1\,\theta} \end{array} \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{center}_{\text{next}_{\varphi}} \\ \text{center}_{\text{next}_{\theta}} \\ \text{center}_{\text{next}_{r}} \end{bmatrix}
  &=& \begin{bmatrix} \text{center}_{\varphi} + \text{offset}_{\varphi} \\
                      \text{center}_{\theta} + \text{offset}_{\theta} \\
                      \text{center}_{r} \end{bmatrix} \\
\begin{bmatrix} \text{sent}_{\varphi} \\ \text{sent}_{\theta} \\ \text{sent}_{r} \end{bmatrix}
  &=& \begin{bmatrix} \text{center}_{\text{next}_{\varphi}} \\
                      \text{center}_{\text{next}_{\theta}} \\
                      \text{center}_{\text{next}_{r}} \end{bmatrix}
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
