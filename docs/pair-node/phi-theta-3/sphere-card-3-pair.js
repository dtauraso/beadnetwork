const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — the edge } a \to b \textbf{, and one channel } b \to a} & & \\[10pt]
\rlap{\textbf{the link}} & & \\[3pt]
a \to b &:& \text{the edge in the topology — it makes the pair; wiring it makes } b \to a \\[3pt]
b \to a &:& b \text{'s channel back to a node with an edge to it} \\[3pt]
& & b \text{ has one } b \to x \text{ channel per node } x \text{ with an edge to it} \\[10pt]
\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{\textbf{local, first — } b \textbf{ holds the angles of every edge into it}} & & \\[3pt]
\text{start}_{xb} &=& \begin{bmatrix} \text{start}_{xb\,\varphi} \\ \text{start}_{xb\,\theta} \end{bmatrix} \text{ — in ticks, held in } b \text{, one per edge } x \to b \\[3pt]
\varphi_{b} &=& \textstyle\sum_{x} \text{start}_{xb\,\varphi} \bmod 12s \\[3pt]
\theta_{b} &=& \textstyle\sum_{x} \text{start}_{xb\,\theta} \bmod 12s \\[3pt]
& & \text{when } b \text{ changes start}_{ab} \text{, it replaces only that term:} \\[3pt]
\varphi_{b} &\leftarrow& (\varphi_{b} - \text{old start}_{ab\,\varphi} + \text{new start}_{ab\,\varphi}) \bmod 12s \\[3pt]
\theta_{b} &\leftarrow& (\theta_{b} - \text{old start}_{ab\,\theta} + \text{new start}_{ab\,\theta}) \bmod 12s \\[3pt]
& & a \text{ sends no angles — } b \text{ already has them} \\[10pt]
\rlap{\textbf{global, last — from } O \textbf{ to each node, one per node}} & & \\[3pt]
O &=& \text{the scene centre, } [O_{x},\, O_{y},\, O_{z}] \\[3pt]
T_{\varphi} &=& \text{index steps in 1 full turn } \varphi \\[3pt]
T_{\theta} &=& \text{index steps in 1 full turn } \theta \\[3pt]
s &:& 12s \text{ divides } T_{\varphi} \text{ and } T_{\theta} \text{, so a tick is a whole number of steps} \\[3pt]
c_{\varphi} &=& 2\pi / T_{\varphi} \\[3pt]
c_{\theta} &=& 2\pi / T_{\theta} \\[3pt]
c_{r} &=& \text{the length of 1 radial index step} \\[3pt]
\begin{bmatrix} I_{b\,\varphi} \\ I_{b\,\theta} \\ I_{b\,r} \end{bmatrix}
  &=& b\text{'s scene index (base + drag)} \\[3pt]
C_{b} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{b\,r}\, c_{r} \\ I_{b\,\varphi}\, c_{\varphi} + \varphi_{b} \cdot 2\pi / 12s \\ I_{b\,\theta}\, c_{\theta} + \theta_{b} \cdot 2\pi / 12s \end{bmatrix}\right) \\[3pt]
& & \text{every node's global is built this way, from its own index and its own sums} \\[10pt]
\rlap{\textbf{the triangle — } O,\; C_{a},\; C_{b}} & & \\[3pt]
O \to a &:& C_{a} - O \text{ — } a \text{'s own global, built in } a \\[3pt]
O \to b &:& C_{b} - O \text{ — } b \text{'s own global, built in } b \\[3pt]
b \to a &:& \text{carries } C_{b} \text{, so } a \text{ holds both ends} \\[3pt]
a \to b &=& C_{b} - C_{a} \text{ — the third side, subtracted in } a \\[3pt]
\text{start}_{ab\,r} &=& |C_{b} - C_{a}| \,/\, \rho \text{ — derived, so it gets there; need not be whole} \\[3pt]
& & b \text{ re-sends } C_{b} \text{ on every } b \to x \text{ when } C_{b} \text{ changes} \\[10pt]
\rlap{\textbf{example — } s = 1 \textbf{,  } O = 0 \textbf{,  } c_{\varphi} = c_{\theta} = 30^{\circ} \textbf{,  } c_{r} = 1 \textbf{,  } C_{a} = (1, 0, 0)} & & \\[3pt]
& & I_{b} = [3,\, 0,\, 2] \;\to\; \varphi = 90^{\circ},\; \theta = 0^{\circ},\; R = 2 \\[3pt]
& & b \text{ holds start}_{ab} = [0,\, 2],\; \text{start}_{cb} = [0,\, 1] \;\to\; \varphi_{b} = 0,\; \theta_{b} = 3 \\[3pt]
& & C_{b} = \textbf{polar_to_cart}(2,\, 90^{\circ},\, 0^{\circ} + 90^{\circ}) = (0, 0, 2) \text{ — sent on } b \to a \\[3pt]
& & a \to b = (0, 0, 2) - (1, 0, 0) = (-1, 0, 2) \\[3pt]
& & b \text{ sets start}_{ab} = [0,\, 3] \;\to\; \varphi_{b} = 0 - 0 + 0 = 0,\; \theta_{b} = 3 - 2 + 3 = 4 \\[3pt]
& & C_{b} = \textbf{polar_to_cart}(2,\, 90^{\circ},\, 120^{\circ}) = (-1, 0, 1.73) \text{ — re-sent on } b \to a \\[3pt]
& & a \to b = (-1, 0, 1.73) - (1, 0, 0) = (-2, 0, 1.73)
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
