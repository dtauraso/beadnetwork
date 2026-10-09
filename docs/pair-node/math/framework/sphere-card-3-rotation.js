const SPHERE_CARD_3_ROTATION_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{shared — one card change, turned about the node the user selects}} & & \\[3pt]
O &=& \text{the scene centre} \\[3pt]
P &=& \text{the rotated node's centre} \\[3pt]
\Delta\varphi,\, \Delta\theta &=& \text{the local card's change for link } j \text{, in ticks: start}_{j} + t_{j} \text{ on start, the direction sum each round} \\[3pt]
a_{\varphi} &=& \Delta\varphi \cdot (T_{\varphi} / 12s) \cdot c_{\varphi} \\[3pt]
a_{\theta} &=& \Delta\theta \cdot (T_{\theta} / 12s) \cdot c_{\theta} \\[10pt]
\rlap{\textbf{1. centre}} & & \\[3pt]
C &=& \text{the centre of the selected node} \\[10pt]
\rlap{\textbf{2. local angles — two separate rotations}} & & \\[3pt]
\begin{bmatrix} r \\ \varphi \\ \theta \end{bmatrix} &=& \begin{bmatrix} |P - C| \\ \operatorname{acos}\!\left((P - C)_{y} / r\right) \\ \operatorname{atan2}\!\left((P - C)_{z},\; (P - C)_{x}\right) \end{bmatrix}
      \quad \text{measured once, when } C \text{ is set; kept from then on} \\[3pt]
\varphi &\leftarrow& \varphi + a_{\varphi} \quad \text{tips the vector toward or away from } {+y} \text{; } \theta \text{ and } r \text{ stay} \\[3pt]
\theta &\leftarrow& \theta + a_{\theta} \quad \text{swings the vector round the vertical through } C \text{; } \varphi \text{ and } r \text{ stay} \\[3pt]
& & \text{each changes only its own angle, so their order does not matter; a } 0 \text{ change moves nothing} \\[10pt]
\rlap{\textbf{3. local vector — starts at } C} & & \\[3pt]
d &=& \begin{bmatrix} r \sin\varphi \cos\theta \\ r \cos\varphi \\ r \sin\varphi \sin\theta \end{bmatrix} \\[10pt]
\rlap{\textbf{4. new place — only the end of the vector moves}} & & \\[3pt]
P' &=& C + d \quad \text{only this direction is converted, so no angle is wrapped and nothing jumps} \\[10pt]
\rlap{\textbf{5. read from } O \textbf{ — output only, never fed back into 2 to 4}} & & \\[3pt]
\begin{bmatrix} R' \\ \Phi' \\ \Theta' \end{bmatrix} &=& \begin{bmatrix} |P' - O| \\ \operatorname{acos}\!\left((P' - O)_{y} / R'\right) \\ \operatorname{atan2}\!\left((P' - O)_{z},\; (P' - O)_{x}\right) \end{bmatrix} \\[3pt]
& & \text{not } \varphi,\, \theta \text{: the node turned about } C \text{, so from } O \text{ it sweeps a different angle and } R' \text{ changes} \\[3pt]
& & \text{the scene index is these rounded (the global card, step 2)} \\[10pt]
\rlap{\textbf{example — } a_{\theta} = 90^{\circ} \textbf{,  } O = (0, 0, 0) \textbf{,  } C = (1, 0, 0) \textbf{,  } P = (2, 0, 0)} & & \\[3pt]
2. & & r = 1,\; \varphi = 90^{\circ},\; \theta = 0^{\circ} \;\to\; \theta = 90^{\circ} \\[3pt]
3. & & d = (1 \cdot \sin 90^{\circ} \cos 90^{\circ},\; 1 \cdot \cos 90^{\circ},\; 1 \cdot \sin 90^{\circ} \sin 90^{\circ}) = (0, 0, 1) \\[3pt]
4. & & P' = (1, 0, 0) + (0, 0, 1) = (1, 0, 1) \\[3pt]
5. & & R' = \sqrt{2},\; \Phi' = 90^{\circ},\; \Theta' = 45^{\circ} \quad \text{a local turn of } 90^{\circ} \text{ is } 45^{\circ} \text{ seen from } O \\[10pt]
\rlap{\textbf{everything not rotated}} & & \\[3pt]
& & \text{keeps its place relative to } C \text{ — it comes along as a unit} \\[3pt]
& & \text{chaining is these rotations one after another, with } C \text{ set again for each}
\end{array}
\]`;

function sphereCard3RotationMath(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3-rotation]')) {
  host.appendChild(sphereCard3RotationMath(SPHERE_CARD_3_ROTATION_FORMULAS));
}
