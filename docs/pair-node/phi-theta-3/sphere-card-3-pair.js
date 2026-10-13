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
v &=& \textstyle\sum_{\text{rows}} d_{x} \text{ — the recovered vector, in cartesian} \\[3pt]
& & \text{the rows' turns change where and how long } v \text{ is; this is the only way to know them} \\[3pt]
\rlap{\textbf{convert — } v \textbf{ onto the scene sphere, in index steps}} & & \\[3pt]
T_{\varphi} &=& \text{index steps in 1 full turn } \varphi \\[3pt]
T_{\theta} &=& \text{index steps in 1 full turn } \theta \\[3pt]
c_{\varphi} &=& 2\pi / T_{\varphi} \\[3pt]
c_{\theta} &=& 2\pi / T_{\theta} \\[3pt]
c_{r} &=& \text{the length of 1 radial index step} \\[3pt]
I_{b\,r} &=& |v| \,/\, c_{r} \\[3pt]
I_{b\,\varphi} &=& \arccos(v_{y} / |v|) \,/\, c_{\varphi} \\[3pt]
I_{b\,\theta} &=& \operatorname{atan2}(v_{z},\, v_{x}) \,/\, c_{\theta} \\[3pt]
& & \text{kept exact — not rounded to whole index steps; a drag changes } I_{b} \text{ after this} \\[3pt]
& & \text{no rows: } v = 0 \text{, so } I_{b} = 0 \\[3pt]
\rlap{\textbf{place — } b \textbf{ on the scene sphere, from its index}} & & \\[3pt]
C_{b} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{b\,r}\, c_{r} \\ I_{b\,\varphi}\, c_{\varphi} \\ I_{b\,\theta}\, c_{\theta} \end{bmatrix}\right) \\[3pt]
& & \text{every node's global is built this way, from its own rows} \\[10pt]
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
