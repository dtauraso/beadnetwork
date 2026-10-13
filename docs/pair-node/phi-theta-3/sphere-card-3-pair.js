const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — } a \textbf{ and } b \textbf{ have a channel each way}} & & \\[10pt]
\rlap{\textbf{two channels}} & & \\[3pt]
b \to a &:& b \text{'s channel back to a node that sends to it} \\[3pt]
a \to b &:& a \text{'s channel to the node its vector is drawn to} \\[3pt]
& & b \text{ has one in-channel per node that sends to it} \\[10pt]
\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{\textbf{local, first — the pair's angles, from } a \to b \textbf{'s start}} & & \\[3pt]
\text{from}_{a} &=& \begin{bmatrix} \text{start}_{ab\,\varphi} \\ \text{start}_{ab\,\theta} \end{bmatrix} \text{ — in ticks} \\[10pt]
\rlap{b \textbf{ adds the angles from every node that sends to it}} & & \\[3pt]
\varphi_{b} &=& \textstyle\sum_{x} \text{from}_{x\,\varphi} \bmod 12s \\[3pt]
\theta_{b} &=& \textstyle\sum_{x} \text{from}_{x\,\theta} \bmod 12s \\[3pt]
& & b \text{ keeps each sender's angles; new angles from } a \text{ replace only } a\text{'s:} \\[3pt]
\varphi_{b} &\leftarrow& (\varphi_{b} - \text{old from}_{a\,\varphi} + \text{new from}_{a\,\varphi}) \bmod 12s \text{, the same for } \theta_{b} \\[10pt]
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
\rlap{\textbf{what the channels carry}} & & \\[3pt]
a \to b &:& \text{from}_{a} \text{ — } b \text{ adds it into } \varphi_{b},\, \theta_{b} \\[3pt]
b \to a &:& C_{b} \text{ — } a \text{ draws its vector from } C_{a} \text{ to } C_{b} \\[10pt]
\rlap{\textbf{example — } s = 1 \textbf{,  } O = 0 \textbf{,  } c_{\varphi} = c_{\theta} = 30^{\circ} \textbf{,  } c_{r} = 1 \textbf{,  } C_{a} = (1, 0, 0)} & & \\[3pt]
& & I_{b} = [3,\, 0,\, 2] \;\to\; \varphi = 90^{\circ},\; \theta = 0^{\circ},\; R = 2 \\[3pt]
& & \text{from}_{a} = [0,\, 2],\; \text{from}_{c} = [0,\, 1] \;\to\; \varphi_{b} = 0,\; \theta_{b} = 3 \\[3pt]
& & C_{b} = \textbf{polar_to_cart}(2,\, 90^{\circ},\, 0^{\circ} + 90^{\circ}) = (0, 0, 2) \\[3pt]
& & a \text{ draws } (1, 0, 0) \to (0, 0, 2) \\[3pt]
& & a \text{ sends } [0,\, 3] \;\to\; \theta_{b} = 3 - 2 + 3 = 4 \\[3pt]
& & C_{b} = \textbf{polar_to_cart}(2,\, 90^{\circ},\, 120^{\circ}) = (-1, 0, 1.73) \\[3pt]
& & a \text{ draws } (1, 0, 0) \to (-1, 0, 1.73)
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
