const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — } a \textbf{ and } b \textbf{ have a channel each way}} & & \\[3pt]
a &=& \text{the centre} \\[3pt]
b &=& \text{the rotating node — it holds the rotation: start}_{b}\text{, end}_{b} \\[10pt]
\rlap{\textbf{two channels}} & & \\[3pt]
b \to a &:& b \text{'s channel to the node its position is derived from} \\[3pt]
a \to b &:& a \text{'s channel to the node its vector is drawn to} \\[10pt]
\rlap{\textbf{scene — one sphere every node sits on}} & & \\[3pt]
O &=& \text{the scene centre, } [O_{x},\, O_{y},\, O_{z}] \\[3pt]
T_{\varphi} &=& \text{index steps in 1 full turn } \varphi \\[3pt]
T_{\theta} &=& \text{index steps in 1 full turn } \theta \\[3pt]
s &:& 12s \text{ divides } T_{\varphi} \text{ and } T_{\theta} \text{, so a tick is a whole number of steps} \\[3pt]
c_{\varphi} &=& 2\pi / T_{\varphi} \\[3pt]
c_{\theta} &=& 2\pi / T_{\theta} \\[3pt]
c_{r} &=& \text{the length of 1 radial index step} \\[3pt]
\begin{bmatrix} I_{a\,\varphi} \\ I_{a\,\theta} \\ I_{a\,r} \end{bmatrix}
  &=& a\text{'s scene index (base + drag)} \\[3pt]
\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{a \textbf{'s centre — from its index}} & & \\[3pt]
C_{a} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{a\,r}\, c_{r} \\ I_{a\,\varphi}\, c_{\varphi} \\ I_{a\,\theta}\, c_{\theta} \end{bmatrix}\right) \\[10pt]
\rlap{b \textbf{'s offset — from its own rotation}} & & \\[3pt]
\rho &=& \text{the node radius} \\[3pt]
\varphi &=& \text{start}_{b\,\varphi} \cdot 2\pi / 12s \\[3pt]
\theta &=& \text{start}_{b\,\theta} \cdot 2\pi / 12s \\[3pt]
R &=& \text{start}_{b\,r} \cdot \rho \\[3pt]
d &=& \textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \varphi \\ \theta \end{bmatrix}\right) \\[10pt]
\rlap{\textbf{what the channels carry}} & & \\[3pt]
a \to b &:& C_{a} \text{ — } b \text{ needs it to derive } C_{b} = C_{a} + d \\[3pt]
b \to a &:& C_{b} \text{ — } a \text{ needs it to draw its vector from } C_{a} \text{ to } C_{b} \\[3pt]
& & b \text{ sends } C_{b} \text{ to } a \text{ only — each pair on its own, no chain} \\[10pt]
\rlap{\textbf{example — } s = 1 \textbf{,  } C_{a} = (1, 0, 0) \textbf{,  } \rho = 1} & & \\[3pt]
& & \text{start}_{b} = [3,\, 3,\, 1] \;\to\; \varphi = 90^{\circ},\; \theta = 90^{\circ},\; R = 1 \\[3pt]
& & d = (1 \cdot \sin 90^{\circ} \cos 90^{\circ},\; 1 \cdot \cos 90^{\circ},\; 1 \cdot \sin 90^{\circ} \sin 90^{\circ}) = (0, 0, 1) \\[3pt]
& & C_{b} = (1, 0, 0) + (0, 0, 1) = (1, 0, 1) \\[3pt]
& & a \text{ draws } (1, 0, 0) \to (1, 0, 1)
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
