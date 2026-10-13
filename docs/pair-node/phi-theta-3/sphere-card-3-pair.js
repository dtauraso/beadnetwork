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
& & b \text{ re-sends } C_{b} \text{ on every } b \to x \text{ when } C_{b} \text{ changes} \\[10pt]
\rlap{\textbf{example — sun, earth, moon:  } s = 1 \textbf{,  } O = 0 \textbf{,  } \rho = 1 \textbf{,  the sun has no rows, so } C_{\text{sun}} = O} & & \\[3pt]
& & \text{edges: sun} \to \text{earth},\; \text{sun} \to \text{moon},\; \text{earth} \to \text{moon} \text{ — start} = [\varphi,\, \theta,\, r] \\[3pt]
& & \text{earth holds } \begin{array}{c|c} \text{from} & \text{start} \\ \hline \text{sun} & [3,\, 0,\, 10] \end{array} \;\to\; C_{\text{earth}} = (10, 0, 0) \\[3pt]
& & \text{moon holds } \begin{array}{c|c} \text{from} & \text{start} \\ \hline \text{sun} & [3,\, 0,\, 10] \\ \text{earth} & [3,\, 0,\, 1] \end{array} \;\to\; C_{\text{moon}} = (10, 0, 0) + (1, 0, 0) = (11, 0, 0) \\[3pt]
& & \text{read back: } r = 11,\; \varphi = 90^{\circ},\; \theta = 0^{\circ} \\[3pt]
& & \text{moon sets row earth's start} = [3,\, 6,\, 1] \;\to\; C_{\text{moon}} = (10, 0, 0) + (-1, 0, 0) = (9, 0, 0) \\[3pt]
& & \text{read back: } r = 9,\; \varphi = 90^{\circ},\; \theta = 0^{\circ} \text{ — the turn round the earth changed the radius} \\[3pt]
& & \text{earth} \to \text{moon} = (9, 0, 0) - (10, 0, 0) = (-1, 0, 0) \\[3pt]
& & \text{both sun rows set to } [3,\, 3,\, 10] \;\to\; C_{\text{earth}} = (0, 0, 10),\; C_{\text{moon}} = (0, 0, 10) + (-1, 0, 0) = (-1, 0, 10) \\[3pt]
& & \text{read back: } r = 10.05,\; \varphi = 90^{\circ},\; \theta = 95.7^{\circ} \\[3pt]
& & \text{earth} \to \text{moon} = (-1, 0, 10) - (0, 0, 10) = (-1, 0, 0) \text{ — the moon went round the sun with the earth}
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
