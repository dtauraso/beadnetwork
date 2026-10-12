const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — } a \textbf{ and } b \textbf{ have a channel each way}} & & \\[3pt]
a &=& \text{the centre} \\[3pt]
b &=& \text{the rotating node — it holds the rotation: start}_{b}\text{, end}_{b} \\[3pt]
L &:& a\text{'s } L_{b} = 1 \text{ and } b\text{'s } L_{a} = 1 \\[10pt]
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
\text{cur}_{b} &=& b\text{'s current angles in ticks, from start}_{b} \cdot s \text{ toward end}_{b} \\[3pt]
\rho &=& \text{the node radius} \\[3pt]
\varphi &=& \text{cur}_{b\,\varphi} \cdot 2\pi / 12s \\[3pt]
\theta &=& \text{cur}_{b\,\theta} \cdot 2\pi / 12s \\[3pt]
R &=& \text{start}_{b\,r} \cdot \rho \\[3pt]
d &=& \textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \varphi \\ \theta \end{bmatrix}\right) \\[10pt]
\rlap{\textbf{what the channels carry}} & & \\[3pt]
a \to b &:& C_{a} \text{ — } b \text{ needs it to derive } C_{b} = C_{a} + d \\[3pt]
b \to a &:& C_{b} \text{ — } a \text{ needs it to draw its vector from } C_{a} \text{ to } C_{b}
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
