const SPHERE_CARD_3_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_3_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
a &=& \text{the step scalar, the same on both axes} \\[3pt]
a &=& 2 \\[3pt]
\tau_{\varphi} &=& \text{the angle index on } \varphi \\[3pt]
\tau_{\theta} &=& \text{the angle index on } \theta \\[6pt]
\begin{bmatrix} \text{seed}^{A}_{\varphi} \\ \text{seed}^{A}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2\tau_{\varphi},\, 8\tau_{\varphi} \\ 2\tau_{\theta},\, 8\tau_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{seed}^{B}_{\varphi} \\ \text{seed}^{B}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 24\tau_{\varphi},\, 30\tau_{\varphi} \\ 24\tau_{\theta},\, 30\tau_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{seed}^{C}_{\varphi} \\ \text{seed}^{C}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 13\tau_{\varphi},\, 19\tau_{\varphi} \\ 13\tau_{\theta},\, 19\tau_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} A_{c_{\varphi}} \\ A_{c_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \text{seed}^{A}_{\varphi} \\ \text{seed}^{A}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} B_{c_{\varphi}} \\ B_{c_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} (\text{seed}^{B}_{\varphi} + \tau_{\varphi}/3) \bmod \tau_{\varphi} \\[3pt]
                      (\text{seed}^{B}_{\theta} + \tau_{\theta}/3) \bmod \tau_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} C_{c_{\varphi}} \\ C_{c_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} (\text{seed}^{C}_{\varphi} + 2\tau_{\varphi}/3) \bmod \tau_{\varphi} \\[3pt]
                      (\text{seed}^{C}_{\theta} + 2\tau_{\theta}/3) \bmod \tau_{\theta} \end{bmatrix} \\[10pt]
\textbf{one node} & & \\[3pt]
\text{center} &=& (c_{\varphi},\, c_{\theta},\, c_{r}) \\[3pt]
\begin{bmatrix} \text{top}_{\varphi} \\ \text{top}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[3pt]
\begin{bmatrix} \text{bottom}_{\varphi} \\ \text{bottom}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} (\text{top}_{\varphi} + \tau_{\varphi}/2) \bmod \tau_{\varphi} \\
                      (\text{top}_{\theta} + \tau_{\theta}/2) \bmod \tau_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{arrival}^{1}_{\varphi} \\ \text{arrival}^{1}_{\theta} \end{bmatrix}
  &=& \text{from partner } 1 \\[6pt]
\begin{bmatrix} \text{arrival}^{2}_{\varphi} \\ \text{arrival}^{2}_{\theta} \end{bmatrix}
  &=& \text{from partner } 2 \\[6pt]
\begin{bmatrix} \text{distance}^{k}_{\text{top}_{\varphi}} \\ \text{distance}^{k}_{\text{top}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} |\, \text{top}_{\varphi} - \text{arrival}^{k}_{\varphi} \,| \bmod \tau_{\varphi}/4 \\[6pt]
                      |\, \text{top}_{\theta} - \text{arrival}^{k}_{\theta} \,| \bmod \tau_{\theta}/4 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{distance}^{k}_{\text{bottom}_{\varphi}} \\ \text{distance}^{k}_{\text{bottom}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} |\, \text{bottom}_{\varphi} - \text{arrival}^{k}_{\varphi} \,| \bmod \tau_{\varphi}/4 \\[6pt]
                      |\, \text{bottom}_{\theta} - \text{arrival}^{k}_{\theta} \,| \bmod \tau_{\theta}/4 \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{offset}_{\varphi} \\ \text{offset}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix}
      0 & \begin{array}{@{}l@{}} \text{if } \text{distance}^{k}_{\text{top}_{\varphi}} = 0 \\ \text{and } \text{distance}^{k}_{\text{bottom}_{\varphi}} = 0 \end{array} \\[8pt]
      0 &
    \end{bmatrix} \ (1_{\varphi}) \\
& & \begin{bmatrix}
      -1 & \text{if } \text{distance}^{k}_{\text{top}_{\varphi}} < \tau_{\varphi}/4 \\
      0 &
    \end{bmatrix} \ (2_{\varphi}) \\
& & \begin{bmatrix}
      -1 & \text{if } \text{distance}^{k}_{\text{bottom}_{\varphi}} < \tau_{\varphi}/4 \\
      0 &
    \end{bmatrix} \ (3_{\varphi}) \\
& & \begin{bmatrix}
      0 & \text{otherwise} \\
      0 &
    \end{bmatrix} \ (4_{\varphi}) \\
& & \begin{bmatrix}
      0 & \\[8pt]
      0 & \begin{array}{@{}l@{}} \text{if } \text{distance}^{k}_{\text{top}_{\theta}} = 0 \\ \text{and } \text{distance}^{k}_{\text{bottom}_{\theta}} = 0 \end{array}
    \end{bmatrix} \ (1_{\theta}) \\
& & \begin{bmatrix}
      0 & \\
      -1 & \text{if } \text{distance}^{k}_{\text{top}_{\theta}} < \tau_{\theta}/4
    \end{bmatrix} \ (2_{\theta}) \\
& & \begin{bmatrix}
      0 & \\
      -1 & \text{if } \text{distance}^{k}_{\text{bottom}_{\theta}} < \tau_{\theta}/4
    \end{bmatrix} \ (3_{\theta}) \\
& & \begin{bmatrix}
      0 & \\
      0 & \text{otherwise}
    \end{bmatrix} \ (4_{\theta}) \\[6pt]
\begin{bmatrix} \text{center}_{\text{next}_{\varphi}} \\ \text{center}_{\text{next}_{\theta}} \\ \text{center}_{\text{next}_{r}} \end{bmatrix}
  &=& \begin{bmatrix} c_{\varphi} + \text{offset}_{\varphi} \\
                      c_{\theta} + \text{offset}_{\theta} \\
                      c_{r} \end{bmatrix} \\
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
