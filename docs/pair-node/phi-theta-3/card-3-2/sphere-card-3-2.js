const SPHERE_CARD_3_2_FORMULAS =String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
s &=& \text{the tick mark scale} \\[3pt]
s &=& 1 \\[3pt]
12s &=& \text{1 full turn } \varphi \\[3pt]
12s &=& \text{1 full turn } \theta \\[3pt]
m &=& \text{the jump constant} \\[3pt]
m &\in& \{1,\, 2,\, \ldots,\, 12\} \\[3pt]
m &=& 1 \text{ by default} \\[6pt]
\text{start}_{\varphi},\, \text{start}_{\theta} &\in& \{0,\, 1s,\, \ldots,\, 11s\} \\[3pt]
\text{start} &=& \begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \\ \text{start}_{r} \end{bmatrix} \\[6pt]
\text{end}_{\varphi},\, \text{end}_{\theta} &\in& \{0,\, 1s,\, \ldots,\, 11s\} \\[3pt]
\text{end} &=& \begin{bmatrix} \text{end}_{\varphi} \\ \text{end}_{\theta} \end{bmatrix} \\[6pt]
\text{cw} &=& \text{clockwise} \\[3pt]
\text{ccw} &=& \text{counterclockwise} \\[3pt]
\text{cw} &\in& \{0, 1\} \\[3pt]
\text{ccw} &\in& \{0, 1\}
\end{array}
\]
${SPHERE_CARD_3_2_START}`;

function sphereCard32Math(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3-2]')) {
  host.appendChild(sphereCard32Math(SPHERE_CARD_3_2_FORMULAS));
}
