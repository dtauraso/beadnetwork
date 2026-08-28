const SPHERE_CARD_3_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_3_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
\begin{bmatrix} \text{seed}_{\varphi} \\ \text{seed}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0,\, \tau_{\varphi} - 1 \\ 0,\, \tau_{\theta} - 1 \end{bmatrix} \\
& & \quad\text{one seed per angle} \\[6pt]
\begin{bmatrix} A_{c_{\varphi}} \\ A_{c_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \text{seed}_{\varphi} \\ \text{seed}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} B_{c_{\varphi}} \\ B_{c_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} (\text{seed}_{\varphi} + \tau_{\varphi}/3) \bmod \tau_{\varphi} \\[3pt]
                      (\text{seed}_{\theta} + \tau_{\theta}/3) \bmod \tau_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} C_{c_{\varphi}} \\ C_{c_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} (\text{seed}_{\varphi} + 2\tau_{\varphi}/3) \bmod \tau_{\varphi} \\[3pt]
                      (\text{seed}_{\theta} + 2\tau_{\theta}/3) \bmod \tau_{\theta} \end{bmatrix} \\
& & \quad\text{a third of a turn apart, one side per node} \\[10pt]
\textbf{one node} & & \\[3pt]
\text{center} &=& (c_{\varphi},\, c_{\theta},\, c_{r}) \\[3pt]
\tau_{\varphi} &=& \text{the whole turn on } \varphi \\[3pt]
\tau_{\theta} &=& \text{the whole turn on } \theta \\[3pt]
\begin{bmatrix} \text{top}_{\varphi} \\ \text{top}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[3pt]
\begin{bmatrix} \text{bottom}_{\varphi} \\ \text{bottom}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} (\text{top}_{\varphi} + \tau_{\varphi}/2) \bmod \tau_{\varphi} \\
                      (\text{top}_{\theta} + \tau_{\theta}/2) \bmod \tau_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{arrival}^{1}_{\varphi} \\ \text{arrival}^{1}_{\theta} \end{bmatrix},
\begin{bmatrix} \text{arrival}^{2}_{\varphi} \\ \text{arrival}^{2}_{\theta} \end{bmatrix}
  &=& \text{one from each partner, per round} \\[6pt]
\begin{bmatrix} \text{distance}^{k}_{\text{top}_{\varphi}} \\ \text{distance}^{k}_{\text{top}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} |\, \text{top}_{\varphi} - \text{arrival}^{k}_{\varphi} \,| \bmod \tau_{\varphi}/4 \\[6pt]
                      |\, \text{top}_{\theta} - \text{arrival}^{k}_{\theta} \,| \bmod \tau_{\theta}/4 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{distance}^{k}_{\text{bottom}_{\varphi}} \\ \text{distance}^{k}_{\text{bottom}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} |\, \text{bottom}_{\varphi} - \text{arrival}^{k}_{\varphi} \,| \bmod \tau_{\varphi}/4 \\[6pt]
                      |\, \text{bottom}_{\theta} - \text{arrival}^{k}_{\theta} \,| \bmod \tau_{\theta}/4 \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{offset}_{\varphi} \\ \text{offset}_{\theta} \end{bmatrix}
  &=& \text{the rule for TWO arrivals, per angle} \\
& & \quad\text{which hemisphere each arrival is in} \\
& & \quad\text{which of the two distances is longer} \\
& & \quad\text{and what the pair of them does together} \\[6pt]
\begin{bmatrix} \text{center}_{\text{next}_{\varphi}} \\ \text{center}_{\text{next}_{\theta}} \\ \text{center}_{\text{next}_{r}} \end{bmatrix}
  &=& \begin{bmatrix} c_{\varphi} + \text{offset}_{\varphi} \\
                      c_{\theta} + \text{offset}_{\theta} \\
                      c_{r} \end{bmatrix} \\
\begin{bmatrix} \text{sent}_{\varphi} \\ \text{sent}_{\theta} \\ \text{sent}_{r} \end{bmatrix}
  &=& \begin{bmatrix} \text{center}_{\text{next}_{\varphi}} \\
                      \text{center}_{\text{next}_{\theta}} \\
                      \text{center}_{\text{next}_{r}} \end{bmatrix} \\
& & \quad\text{the same value to both partners}
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
