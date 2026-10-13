const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — the edge } a \to b \textbf{, and one channel } b \to a} & & \\[10pt]
\rlap{\textbf{the link}} & & \\[3pt]
a \to b &:& \text{the edge in the topology — it makes the pair; wiring it makes } b \to a \\[3pt]
b \to a &:& b \text{'s channel back to a node with an edge to it} \\[3pt]
& & b \text{ has one } b \to x \text{ channel per node } x \text{ with an edge to it} \\[10pt]
\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{\textbf{local, first — } b \textbf{'s table, one row per node with an edge into } b} & & \\[3pt]
& & \begin{array}{c|c|c} \text{from} & \text{start} & \text{end} \\ \hline a & \text{start} & \text{end} \\ c & \text{start} & \text{end} \end{array} \text{ — in ticks, held in } b \\[3pt]
& & a \text{ sends no angles — } b \text{ already has them} \\[10pt]
\rlap{\textbf{global, last — from } O \textbf{ to each node, one per node}} & & \\[3pt]
O &=& \text{the scene centre, } [O_{x},\, O_{y},\, O_{z}] \\[3pt]
\rho &=& \text{the length of 1 radial tick} \\[3pt]
d_{x} &=& \textbf{polar_to_cart}\left(\begin{bmatrix} \text{start}_{r}\, \rho \\ \text{start}_{\varphi} \cdot 2\pi / 12s \\ \text{start}_{\theta} \cdot 2\pi / 12s \end{bmatrix}\right) \text{ — row } x \text{'s vector, from row } x \text{'s start} \\[3pt]
C_{b} &=& O + \textstyle\sum_{\text{rows}} d_{x} \\[3pt]
& & \text{every node's global is built this way, from its own rows} \\[3pt]
\rlap{\textbf{read back — } b \textbf{'s radius and angles, from the trig}} & & \\[3pt]
v &=& C_{b} - O \\[3pt]
r_{b} &=& |v| \\[3pt]
\varphi_{b} &=& \arccos(v_{y} / r_{b}) \\[3pt]
\theta_{b} &=& \operatorname{atan2}(v_{z},\, v_{x}) \\[3pt]
& & \text{the rows' turns change where and how long } v \text{ is; this is the only way to know them} \\[10pt]
\rlap{\textbf{the triangle — } O,\; C_{a},\; C_{b}} & & \\[3pt]
O \to a &:& C_{a} - O \text{ — } a \text{'s own global, built in } a \\[3pt]
O \to b &:& C_{b} - O \text{ — } b \text{'s own global, built in } b \\[3pt]
b \to a &:& \text{carries } C_{b} \text{, so } a \text{ holds both ends} \\[3pt]
a \to b &=& C_{b} - C_{a} \text{ — the third side, subtracted in } a \\[3pt]
|a \to b| &=& |C_{b} - C_{a}| \text{ — derived; need not be whole} \\[3pt]
& & b \text{ re-sends } C_{b} \text{ on every } b \to x \text{ when } C_{b} \text{ changes}
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
