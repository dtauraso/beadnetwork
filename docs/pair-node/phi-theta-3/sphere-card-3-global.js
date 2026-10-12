const SPHERE_CARD_3_GLOBAL_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{scene — one sphere every node sits on}} & & \\[3pt]
O &=& \text{the scene centre, } [O_{x},\, O_{y},\, O_{z}] \\[3pt]
T_{\varphi} &=& \text{index steps in 1 full turn } \varphi \\[3pt]
T_{\theta} &=& \text{index steps in 1 full turn } \theta \\[3pt]
s &:& 12s \text{ divides } T_{\varphi} \text{ and } T_{\theta} \text{, so a tick is a whole number of steps} \\[3pt]
c_{\varphi} &=& 2\pi / T_{\varphi} \\[3pt]
c_{\theta} &=& 2\pi / T_{\theta} \\[3pt]
c_{r} &=& \text{the length of 1 radial index step} \\[3pt]
\varphi &:& \text{measured from the } +y \text{ axis} \\[3pt]
\theta &:& \text{measured in the } x z \text{ plane from the } +x \text{ axis} \\[10pt]
\rlap{\textbf{node } n} & & \\[3pt]
\begin{bmatrix} I_{n\,\varphi} \\ I_{n\,\theta} \\ I_{n\,r} \end{bmatrix}
  &=& \text{node } n\text{'s scene index (base + drag)} \\[10pt]
\rlap{\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)} & & \\[3pt]
\begin{bmatrix} x \\ y \\ z \end{bmatrix}
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{\textbf{cart_to_polar}\left(\begin{bmatrix} x \\ y \\ z \end{bmatrix}\right)} & & \\[3pt]
\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}
  &=& \begin{bmatrix} \sqrt{x^{2} + y^{2} + z^{2}} \\ \operatorname{atan2}\!\left(\sqrt{x^{2} + z^{2}},\; y\right) \\ \operatorname{atan2}(z,\; x) \end{bmatrix} \\[10pt]
\rlap{\textbf{1. the node's centre — from its index}} & & \\[3pt]
C_{n} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{n\,r}\, c_{r} \\ I_{n\,\varphi}\, c_{\varphi} \\ I_{n\,\theta}\, c_{\theta} \end{bmatrix}\right) \\[10pt]
\rlap{\textbf{2. the scene index after a rotation — output only}} & & \\[3pt]
\begin{bmatrix} R_{n} \\ \Phi_{n} \\ \Theta_{n} \end{bmatrix}
  &=& \textbf{cart_to_polar}\left(P'_{n} - O\right) \quad P'_{n} \text{ from the rotation card, step 4} \\[6pt]
\begin{bmatrix} I_{n\,\varphi} \\ I_{n\,\theta} \\ I_{n\,r} \end{bmatrix}
  &=& \begin{bmatrix} \operatorname{round}(\Phi_{n} / c_{\varphi}) \bmod T_{\varphi} \\[3pt]
                      \operatorname{round}(\Theta_{n} / c_{\theta}) \bmod T_{\theta} \\[3pt]
                      \operatorname{round}(R_{n} / c_{r}) \end{bmatrix}
      \quad \text{never read back into the rotation}
\end{array}
\]`;

function sphereCard3GlobalMath(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3-global]')) {
  host.appendChild(sphereCard3GlobalMath(SPHERE_CARD_3_GLOBAL_FORMULAS));
}
