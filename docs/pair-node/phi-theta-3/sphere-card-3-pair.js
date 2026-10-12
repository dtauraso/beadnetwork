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
a &:& \text{needs where } b \text{ is, to draw its vector to } b \\[10pt]
\rlap{\textbf{one centre, many rotating nodes}} & & \\[3pt]
& & a \text{ has a channel each way with each } b_{i} \\[3pt]
C_{b_{i}} &=& C_{a} + d_{i} \quad \text{each } b_{i} \text{ holds its own rotation; they do not constrain each other} \\[10pt]
\rlap{\textbf{triangle — every node is a centre and a rotating node}} & & \\[3pt]
& & \text{a channel each way for each pair: 6 channels for 3 nodes} \\[3pt]
& & \text{each node derives its position from both centres; both must land on the same point} \\[3pt]
d_{12} + d_{23} + d_{31} &=& 0 \quad \text{the loop closes} \\[3pt]
& & \text{every start and end names the same rotation, so it stays closed} \\[3pt]
& & \text{and the triangle turns as one piece about } G = (C_{1} + C_{2} + C_{3}) / 3
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
