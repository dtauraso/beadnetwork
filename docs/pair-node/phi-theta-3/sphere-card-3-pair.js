const SPHERE_CARD_3_PAIR_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{the pair — } a \textbf{ and } b \textbf{ have a channel each way}} & & \\[10pt]
\rlap{\textbf{two channels}} & & \\[3pt]
b \to a &:& b \text{'s channel to the node its position is derived from} \\[3pt]
a \to b &:& a \text{'s channel to the node its vector is drawn to} \\[10pt]
\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{\textbf{local, first — the pair's offset, from } a \to b \textbf{'s start}} & & \\[3pt]
\rho &=& \text{the node radius} \\[3pt]
\varphi &=& \text{start}_{ab\,\varphi} \cdot 2\pi / 12s \\[3pt]
\theta &=& \text{start}_{ab\,\theta} \cdot 2\pi / 12s \\[3pt]
R &=& \text{start}_{ab\,r} \cdot \rho \\[3pt]
d &=& \textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \varphi \\ \theta \end{bmatrix}\right) \\[10pt]
\rlap{\textbf{global, last — derived for the view}} & & \\[3pt]
O &=& \text{the scene centre, } [O_{x},\, O_{y},\, O_{z}] \\[3pt]
T_{\varphi} &=& \text{index steps in 1 full turn } \varphi \\[3pt]
T_{\theta} &=& \text{index steps in 1 full turn } \theta \\[3pt]
s &:& 12s \text{ divides } T_{\varphi} \text{ and } T_{\theta} \text{, so a tick is a whole number of steps} \\[3pt]
c_{\varphi} &=& 2\pi / T_{\varphi} \\[3pt]
c_{\theta} &=& 2\pi / T_{\theta} \\[3pt]
c_{r} &=& \text{the length of 1 radial index step} \\[3pt]
\begin{bmatrix} I_{a\,\varphi} \\ I_{a\,\theta} \\ I_{a\,r} \end{bmatrix}
  &=& a\text{'s scene index (base + drag)} \\[3pt]
C_{a} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{a\,r}\, c_{r} \\ I_{a\,\varphi}\, c_{\varphi} \\ I_{a\,\theta}\, c_{\theta} \end{bmatrix}\right) \\[3pt]
C_{b} &=& C_{a} + d \\[10pt]
\rlap{\textbf{deriving — along a link adds its } d \textbf{, against it subtracts}} & & \\[3pt]
\text{along } a \to b &:& C_{b} = C_{a} + d_{ab} \\[3pt]
\text{against } c \to b &:& C_{c} = C_{b} - d_{cb} \\[10pt]
\rlap{\textbf{a loop — } c \to a,\; a \to b,\; c \to b \textbf{ all set}} & & \\[3pt]
g &=& d_{ca} + d_{ab} - d_{cb} \\[3pt]
& & \text{the derived places agree} \iff g = 0 \\[10pt]
\rlap{\textbf{what the channels carry}} & & \\[3pt]
a \to b &:& C_{a} \text{ — } b \text{ needs it to derive } C_{b} = C_{a} + d \\[3pt]
b \to a &:& C_{b} \text{ — } a \text{ needs it to draw its vector from } C_{a} \text{ to } C_{b} \\[3pt]
& & b \text{ sends } C_{b} \text{ to } a \text{ only — each pair on its own, no chain} \\[10pt]
\rlap{\textbf{example — } s = 1 \textbf{,  } C_{a} = (1, 0, 0) \textbf{,  } \rho = 1} & & \\[3pt]
& & \text{start}_{ab} = [3,\, 3,\, 2] \;\to\; \varphi = 90^{\circ},\; \theta = 90^{\circ},\; R = 2 \\[3pt]
& & d = (2 \cdot \sin 90^{\circ} \cos 90^{\circ},\; 2 \cdot \cos 90^{\circ},\; 2 \cdot \sin 90^{\circ} \sin 90^{\circ}) = (0, 0, 2) \\[3pt]
& & C_{b} = (1, 0, 0) + (0, 0, 2) = (1, 0, 2) \\[3pt]
& & a \text{ draws } (1, 0, 0) \to (1, 0, 2)
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
