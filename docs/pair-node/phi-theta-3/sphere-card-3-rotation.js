const SPHERE_CARD_3_ROTATION_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{shared — each pair on its own: } b \textbf{ is where } a \to b \textbf{ says}} & & \\[3pt]
a,\, b &=& \text{two nodes, with } a\text{'s } L_{b} = 1 \\[3pt]
C_{a} &=& a\text{'s centre} \\[3pt]
\rho &=& \text{the node radius} \\[10pt]
\rlap{\textbf{1. the } a \to b \textbf{ vector, in ticks}} & & \\[3pt]
\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \\ \text{start}_{r} \end{bmatrix} &=& a\text{'s start} \\[10pt]
\rlap{\textbf{2. as an angle and a length}} & & \\[3pt]
\varphi &=& \text{start}_{\varphi} \cdot 2\pi / 12s \\[3pt]
\theta &=& \text{start}_{\theta} \cdot 2\pi / 12s \\[3pt]
R &=& \text{start}_{r} \cdot \rho \\[10pt]
\rlap{\textbf{3. the vector — starts at } C_{a}} & & \\[3pt]
d &=& \begin{bmatrix} R \sin\varphi \cos\theta \\ R \cos\varphi \\ R \sin\varphi \sin\theta \end{bmatrix} \\[10pt]
\rlap{\textbf{4. } b \textbf{ goes to its head}} & & \\[3pt]
C_{b} &=& C_{a} + d \\[3pt]
& & b \text{ passes nothing on — each pair on its own, no chain, no loop} \\[10pt]
\rlap{\textbf{example — } s = 1 \textbf{,  } C_{a} = (1, 0, 0) \textbf{,  } \rho = 1} & & \\[3pt]
1.\;2. & & \text{start} = [3,\, 3,\, 1] \;\to\; \varphi = 90^{\circ},\; \theta = 90^{\circ},\; R = 1 \\[3pt]
3. & & d = (1 \cdot \sin 90^{\circ} \cos 90^{\circ},\; 1 \cdot \cos 90^{\circ},\; 1 \cdot \sin 90^{\circ} \sin 90^{\circ}) = (0, 0, 1) \\[3pt]
4. & & C_{b} = (1, 0, 0) + (0, 0, 1) = (1, 0, 1)
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
