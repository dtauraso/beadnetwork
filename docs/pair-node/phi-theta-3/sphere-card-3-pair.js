const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — } a \textbf{ and } b \textbf{ have a channel each way}} & & \\[3pt]
a &=& \text{the centre} \\[3pt]
b &=& \text{the rotating node — it holds the rotation: start}_{b}\text{, end}_{b} \\[10pt]
\rlap{\textbf{two channels}} & & \\[3pt]
b \to a &:& b \text{'s channel to the node its position is derived from} \\[3pt]
a \to b &:& a \text{'s channel to the node its vector is drawn to} \\[3pt]
& & \text{the vector is on the } a \to b \text{ channel} \\[10pt]
\rlap{\textbf{what the channels carry}} & & \\[3pt]
b &:& \text{needs } C_{a} \text{ to derive } C_{b} = C_{a} + d \\[3pt]
a &:& \text{needs where } b \text{ is, to draw its vector to } b
\end{array}
\]`;

function sphereCard3PairMath(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3-pair]')) {
  host.appendChild(sphereCard3PairMath(SPHERE_CARD_3_PAIR_FORMULAS));
}
