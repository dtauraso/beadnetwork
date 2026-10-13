const SPHERE_CARD_3_SET_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{setting } a \to b} & & \\[3pt]
a &\in& \{1,\, 2,\, 3\} \\[10pt]
b &\in& \{1,\, 2,\, 3\},\; b \neq a \\[3pt]
\text{start}_{ab} &=& \begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \\ \text{start}_{r} \end{bmatrix} \\[3pt]
\text{end}_{ab} &=& \begin{bmatrix} \text{end}_{\varphi} \\ \text{end}_{\theta} \end{bmatrix}
\end{array}
\]`;

function sphereCard3SetMath(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3-set]')) {
  host.appendChild(sphereCard3SetMath(SPHERE_CARD_3_SET_FORMULAS));
}
