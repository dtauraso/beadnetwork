const SPHERE_CARD_3_2_FORMULAS =String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
s &=& \text{the tick mark scale} \\[3pt]
s &=& 1 \\[3pt]
12s &=& \text{1 full turn } \varphi \\[3pt]
12s &=& \text{1 full turn } \theta \\[3pt]
m &=& \text{the jump constant} \\[3pt]
m &\in& \{1,\, 2,\, \ldots,\, 12\} \\[3pt]
m &=& 1 \text{ by default} \\[3pt]
L_{j} &=& 1 \text{ if a link reaches node } j \text{, } 0 \text{ if not} \\[3pt]
\text{pole_offset}_{j\,\varphi},\, \text{pole_offset}_{j\,\theta} &\in& \{0, 1s, 2s, 3s\} \\[3pt]
\text{start}_{\varphi},\, \text{start}_{\theta} &\in& \{0,\, 1s,\, \ldots,\, 11s\} \\[3pt]
\text{end}_{\varphi},\, \text{end}_{\theta} &\in& \{0,\, 1s,\, \ldots,\, 11s\} \\[3pt]
k_{j} &\in& \{0, 1\} \\[3pt]
\text{cw} &=& \text{clockwise} \\[3pt]
\text{ccw} &=& \text{counterclockwise} \\[10pt]
\textbf{node 1} & & \\[3pt]
\begin{bmatrix} \text{start}_{2\,\varphi} \\ \text{start}_{2\,\theta} \\ \text{start}_{2\,r} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{start}_{3\,\varphi} \\ \text{start}_{3\,\theta} \\ \text{start}_{3\,r} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{2\,\varphi} \\ \text{pole_offset}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2s \\ 2s \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{3\,\varphi} \\ \text{pole_offset}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3s \\ 3s \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{2} \\ k_{3} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 1 \\ 1 \end{bmatrix} \\[10pt]
\textbf{node 2} & & \\[3pt]
\begin{bmatrix} \text{start}_{1\,\varphi} \\ \text{start}_{1\,\theta} \\ \text{start}_{1\,r} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{start}_{3\,\varphi} \\ \text{start}_{3\,\theta} \\ \text{start}_{3\,r} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{1\,\varphi} \\ \text{pole_offset}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3s \\ 3s \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{3\,\varphi} \\ \text{pole_offset}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1s \\ 1s \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{1} \\ k_{3} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 0 \\ 1 \end{bmatrix} \\[10pt]
\textbf{node 3} & & \\[3pt]
\begin{bmatrix} \text{start}_{1\,\varphi} \\ \text{start}_{1\,\theta} \\ \text{start}_{1\,r} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{start}_{2\,\varphi} \\ \text{start}_{2\,\theta} \\ \text{start}_{2\,r} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{1\,\varphi} \\ \text{pole_offset}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2s \\ 2s \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{2\,\varphi} \\ \text{pole_offset}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1s \\ 1s \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{1} \\ k_{2} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \\ 0 \end{bmatrix} \\[10pt]
\text{cw} &\in& \{0, 1\} \\[6pt]
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
