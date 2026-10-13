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
\rlap{\textbf{many starts and ends, grouped by } k} & & \\[3pt]
k &=& xb \text{ — the id of one edge } x \to b \text{ into this node } b \\[3pt]
K_{b} &=& \text{the ids of every edge into } b \text{, one } k \text{ per edge} \\[3pt]
& & \text{start}_{k},\, \text{end}_{k} \text{ share the id } k \text{; start}_{k} \text{ is the pair card's start}_{xb} \\[6pt]
\text{start}_{k\,\varphi},\, \text{start}_{k\,\theta} &\in& \{0,\, 1s,\, \ldots,\, 11s\} \\[3pt]
\text{start}_{k\,r} &\in& \mathbb{Z}_{\ge 2} \\[3pt]
\text{start}_{k} &=& \begin{bmatrix} \text{start}_{k\,\varphi} \\ \text{start}_{k\,\theta} \\ \text{start}_{k\,r} \end{bmatrix} \\[6pt]
\text{end}_{k\,\varphi},\, \text{end}_{k\,\theta} &\in& \{0,\, 1s,\, \ldots,\, 11s\} \\[3pt]
\text{end}_{k} &=& \begin{bmatrix} \text{end}_{k\,\varphi} \\ \text{end}_{k\,\theta} \end{bmatrix} \\[6pt]
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
