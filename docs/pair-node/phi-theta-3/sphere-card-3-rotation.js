const SPHERE_CARD_3_ROTATION_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{shared — each pair on its own: } b \textbf{ is where } a \to b \textbf{ says}} & & \\[3pt]
O &=& \text{the scene centre} \\[3pt]
a,\, b &=& \text{two nodes, with } a\text{'s } k_{b} = 1 \\[3pt]
C_{a} &=& a\text{'s centre} \\[3pt]
\rho &=& \text{the node radius} \\[10pt]
\rlap{\textbf{1. the } a \to b \textbf{ vector, in ticks}} & & \\[3pt]
\begin{bmatrix} \varphi_{\text{ticks}} \\ \theta_{\text{ticks}} \end{bmatrix} &=& \text{start}_{b} + t_{b} \quad \text{before the first round and after a reset} \\[3pt]
&=& k_{b} \cdot \text{local_arrival}_{b} \quad \text{after a round, unless that is } 0 \text{ — a } 0 \text{ is no change, so the last non-zero one stays} \\[3pt]
r &=& \text{start}_{b\,r} \quad \text{no round changes it} \\[10pt]
\rlap{\textbf{2. as an angle and a length}} & & \\[3pt]
\varphi &=& \varphi_{\text{ticks}} \cdot (T_{\varphi} / 12s) \cdot c_{\varphi} \\[3pt]
\theta &=& \theta_{\text{ticks}} \cdot (T_{\theta} / 12s) \cdot c_{\theta} \\[3pt]
R &=& r \cdot \rho \\[10pt]
\rlap{\textbf{3. the vector — starts at } C_{a}} & & \\[3pt]
d &=& \begin{bmatrix} R \sin\varphi \cos\theta \\ R \cos\varphi \\ R \sin\varphi \sin\theta \end{bmatrix} \\[10pt]
\rlap{\textbf{4. } b \textbf{ goes to its tip}} & & \\[3pt]
C_{b} &=& C_{a} + d \\[3pt]
& & a \text{ sends it at load, on start, when } a\text{'s start, } t\text{, } k \text{ or } s \text{ changes, and after a round that changed the vector} \\[3pt]
& & b \text{ passes nothing on — each pair on its own, no chain, no loop; reset moves nothing} \\[10pt]
\rlap{\textbf{5. read from } O \textbf{ — output only}} & & \\[3pt]
\begin{bmatrix} R_{b} \\ \Phi_{b} \\ \Theta_{b} \end{bmatrix} &=& \begin{bmatrix} |C_{b} - O| \\ \operatorname{acos}\!\left((C_{b} - O)_{y} / R_{b}\right) \\ \operatorname{atan2}\!\left((C_{b} - O)_{z},\; (C_{b} - O)_{x}\right) \end{bmatrix} \\[3pt]
& & \text{not } \varphi,\, \theta \text{: those are about } C_{a} \text{; the scene index is these rounded (the global card, step 2), saved like a drag} \\[10pt]
\rlap{\textbf{example — } s = 1 \textbf{,  } O = (0, 0, 0) \textbf{,  } C_{a} = (1, 0, 0) \textbf{,  } \rho = 1} & & \\[3pt]
1.\;2. & & \varphi_{\text{ticks}} = 3,\; \theta_{\text{ticks}} = 3,\; r = 1 \;\to\; \varphi = 90^{\circ},\; \theta = 90^{\circ},\; R = 1 \\[3pt]
3. & & d = (1 \cdot \sin 90^{\circ} \cos 90^{\circ},\; 1 \cdot \cos 90^{\circ},\; 1 \cdot \sin 90^{\circ} \sin 90^{\circ}) = (0, 0, 1) \\[3pt]
4. & & C_{b} = (1, 0, 0) + (0, 0, 1) = (1, 0, 1) \\[3pt]
5. & & R_{b} = \sqrt{2},\; \Phi_{b} = 90^{\circ},\; \Theta_{b} = 45^{\circ}
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
