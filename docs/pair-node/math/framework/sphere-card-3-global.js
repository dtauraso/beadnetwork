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
\rho &=& \text{the node radius, a length} \\[3pt]
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
\rlap{\textbf{1. the card's start}_{j}\textbf{ as a local vector}} & & \\[3pt]
\begin{bmatrix} V_{j\,R} \\ V_{j\,\Phi} \\ V_{j\,\Theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{start}_{j\,r} \cdot \rho \\[3pt]
                      (\text{start}_{j\,\varphi} + t_{j\,\varphi}) \cdot (T_{\varphi} / 12s) \cdot c_{\varphi} \\[3pt]
                      (\text{start}_{j\,\theta} + t_{j\,\theta}) \cdot (T_{\theta} / 12s) \cdot c_{\theta} \end{bmatrix}
      \quad \text{in ticks: start}_{j} \text{ a multiple of } s,\; 0 \le t_{j} < s \\[10pt]
\rlap{\textbf{2. the node's centre}} & & \\[3pt]
C_{n} &=& \text{the exact tip that placed it} \quad \text{if a partner places } n \text{ (the placement card)} \\[3pt]
C_{n} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{n\,r}\, c_{r} \\ I_{n\,\varphi}\, c_{\varphi} \\ I_{n\,\theta}\, c_{\theta} \end{bmatrix}\right) \quad \text{otherwise} \\[10pt]
\rlap{\textbf{3. the tip of start}_{j}\textbf{ — the arrow runs from } C_{n} \textbf{ to it}} & & \\[3pt]
\text{tip}_{j} &=& C_{n} + \textbf{polar_to_cart}(V_{j}) \\[3pt]
C_{j} &=& \text{tip}_{j} \quad \text{if } k_{j} = 1 \\[10pt]
\rlap{\textbf{4. the scene index beside it — output only}} & & \\[3pt]
\begin{bmatrix} R_{j} \\ \Phi_{j} \\ \Theta_{j} \end{bmatrix}
  &=& \textbf{cart_to_polar}\left(\text{tip}_{j} - O\right) \\[6pt]
\begin{bmatrix} I_{j\,\varphi} \\ I_{j\,\theta} \\ I_{j\,r} \end{bmatrix}
  &=& \begin{bmatrix} \operatorname{round}(\Phi_{j} / c_{\varphi}) \bmod T_{\varphi} \\[3pt]
                      \operatorname{round}(\Theta_{j} / c_{\theta}) \bmod T_{\theta} \\[3pt]
                      \operatorname{round}(R_{j} / c_{r}) \end{bmatrix}
      \quad \text{if } k_{j} = 1 \text{; never read back into } C_{j}
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
