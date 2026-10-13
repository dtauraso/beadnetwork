const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — the edge } a \to b \textbf{, and one channel } b \to a} & & \\[10pt]
\rlap{\textbf{the link}} & & \\[3pt]
a \to b &:& \text{the edge in the topology — it makes the pair; wiring it makes } b \to a \\[3pt]
& & \text{carries } C_{a} \text{ to } b \\[3pt]
b \to a &:& b \text{'s channel back to a node with an edge to it} \\[3pt]
& & b \text{ has one } b \to x \text{ channel per node } x \text{ with an edge to it} \\[10pt]
\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{\textbf{the scene sphere — a position is an index}} & & \\[3pt]
O &=& \text{the scene centre, } [O_{x},\, O_{y},\, O_{z}] \\[3pt]
T_{\varphi} &=& \text{index steps in 1 full turn } \varphi \\[3pt]
T_{\theta} &=& \text{index steps in 1 full turn } \theta \\[3pt]
c_{\varphi} &=& 2\pi / T_{\varphi} \\[3pt]
c_{\theta} &=& 2\pi / T_{\theta} \\[3pt]
c_{r} &=& \text{the length of 1 radial index step} \\[3pt]
C_{a} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{a\,r}\, c_{r} \\ I_{a\,\varphi}\, c_{\varphi} \\ I_{a\,\theta}\, c_{\theta} \end{bmatrix}\right) \text{ — } a \text{'s own global, from } a \text{'s index} \\[10pt]
\rlap{\textbf{local, first — } b \textbf{'s table, one row per node with an edge into } b} & & \\[3pt]
& & \begin{array}{c|c|c} \text{from} & \text{start} & \text{end} \\ \hline a & \text{start} & \text{end} \\ c & \text{start} & \text{end} \end{array} \text{ — in ticks, held in } b \\[3pt]
& & a \text{ sends no angles — } b \text{ already has them} \\[3pt]
\rho &=& \text{the length of 1 radial tick} \\[3pt]
d_{ab} &=& \textbf{polar_to_cart}\left(\begin{bmatrix} \text{start}_{r}\, \rho \\ \text{start}_{\varphi} \cdot 2\pi / 12s \\ \text{start}_{\theta} \cdot 2\pi / 12s \end{bmatrix}\right) \text{ — row } a \text{'s start: the } a \to b \text{ vector, starting at } a \\[10pt]
\rlap{\textbf{global, last — } O \to a \;+\; a \to b} & & \\[3pt]
C_{b} &=& C_{a} + d_{ab} \text{ — in } b \text{, with } C_{a} \text{ from } a \to b \\[3pt]
& & \text{the pair } c \to b \text{ gives } C_{c} + d_{cb} \text{ the same way; the two are one point only when } d_{ca} + d_{ab} - d_{cb} = 0 \text{ (the } a \to b \text{ card)} \\[10pt]
\rlap{\textbf{convert — } C_{b} \textbf{ onto the scene sphere, in index steps}} & & \\[3pt]
I_{b\,r} &=& |C_{b} - O| \,/\, c_{r} \\[3pt]
I_{b\,\varphi} &=& \arccos\big((C_{b} - O)_{y} \,/\, |C_{b} - O|\big) \,/\, c_{\varphi} \\[3pt]
I_{b\,\theta} &=& \operatorname{atan2}\big((C_{b} - O)_{z},\, (C_{b} - O)_{x}\big) \,/\, c_{\theta} \\[3pt]
& & \text{kept exact — not rounded to whole index steps; a drag changes } I_{b} \text{ after this} \\[3pt]
& & C_{b} = O \text{: } |C_{b} - O| = 0 \text{, so } I_{b} = 0 \\[10pt]
\rlap{\textbf{place — } b \textbf{ on the scene sphere, from its index}} & & \\[3pt]
C_{b} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{b\,r}\, c_{r} \\ I_{b\,\varphi}\, c_{\varphi} \\ I_{b\,\theta}\, c_{\theta} \end{bmatrix}\right) \text{ — the same point, now held as } b \text{'s index} \\[10pt]
\rlap{\textbf{the triangle — } O,\; C_{a},\; C_{b}} & & \\[3pt]
O \to a &=& C_{a} - O \\[3pt]
a \to b &=& d_{ab} \text{ — row } a \text{'s start} \\[3pt]
O \to b &=& C_{b} - O \;=\; (O \to a) + (a \to b) \\[3pt]
|a \to b| &=& \text{start}_{r}\, \rho \text{ — need not be a whole number of index steps} \\[3pt]
b \to a &:& \text{carries } C_{b} \text{, so } a \text{ holds both ends} \\[3pt]
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
