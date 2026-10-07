const SPHERE_CARD_3_GLOBAL_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{scene} & & \textbf{— one sphere every node sits on} \\[3pt]
O &=& \text{the scene centre, } [O_{x},\, O_{y},\, O_{z}] \\[3pt]
T_{\text{saved}} &=& \text{the turn the saved positions are written in} \\[3pt]
T_{\varphi} &=& \operatorname{lcm}(T_{\text{saved}\,\varphi},\, 12s) \text{ index steps in 1 full turn } \varphi \\[3pt]
T_{\theta} &=& \operatorname{lcm}(T_{\text{saved}\,\theta},\, 12s) \text{ index steps in 1 full turn } \theta \\[3pt]
\text{a saved position} &\times& T / T_{\text{saved}} \text{ at load} \\[3pt]
c_{\varphi} &=& 2\pi / T_{\varphi} \\[3pt]
c_{\theta} &=& 2\pi / T_{\theta} \\[3pt]
c_{r} &=& \text{the length of 1 radial index step} \\[3pt]
\rho &=& \text{the node radius} \,/\, c_{r} \\[3pt]
\varphi &:& \text{measured from the } +y \text{ axis} \\[3pt]
\theta &:& \text{measured in the } x z \text{ plane from the } +x \text{ axis} \\[10pt]
\textbf{node } n & & \\[3pt]
\begin{bmatrix} I_{n\,\varphi} \\ I_{n\,\theta} \\ I_{n\,r} \end{bmatrix}
  &=& \text{node } n\text{'s scene index (base + drag)} \\[10pt]
\rlap{\textbf{polar_to_cart}\left(\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}\right)} & & \\[3pt]
\begin{bmatrix} x \\ y \\ z \end{bmatrix}
  &=& \begin{bmatrix} R \sin\Phi \cos\Theta \\ R \cos\Phi \\ R \sin\Phi \sin\Theta \end{bmatrix} \\[10pt]
\rlap{\textbf{cart_to_polar}\left(\begin{bmatrix} x \\ y \\ z \end{bmatrix}\right)} & & \\[3pt]
\begin{bmatrix} R \\ \Phi \\ \Theta \end{bmatrix}
  &=& \begin{bmatrix} \sqrt{x^{2} + y^{2} + z^{2}} \\ \operatorname{atan2}\!\left(\sqrt{x^{2} + z^{2}},\; y\right) \\ \operatorname{atan2}(z,\; x) \end{bmatrix} \\[10pt]
\textbf{1. card ticks to an index offset} & & \\[3pt]
\begin{bmatrix} o_{j\,\varphi} \\ o_{j\,\theta} \\ o_{j\,r} \end{bmatrix}
  &=& \begin{bmatrix} \text{start}_{j\,\varphi} \cdot (T_{\varphi} / 12s) \\[3pt]
                      \text{start}_{j\,\theta} \cdot (T_{\theta} / 12s) \\[3pt]
                      \operatorname{round}\!\left(\text{start}_{j\,r} \cdot \rho\right) \end{bmatrix} \\[10pt]
\textbf{2. the node's centre} & & \\[3pt]
C_{n} &=& O + \textbf{polar_to_cart}\left(\begin{bmatrix} I_{n\,r}\, c_{r} \\ I_{n\,\varphi}\, c_{\varphi} \\ I_{n\,\theta}\, c_{\theta} \end{bmatrix}\right) \\[10pt]
\textbf{3. the tip of start}_{j} & & \\[3pt]
\text{tip}_{j} &=& C_{n} + \textbf{polar_to_cart}\left(\begin{bmatrix} o_{j\,r}\, c_{r} \\ o_{j\,\varphi}\, c_{\varphi} \\ o_{j\,\theta}\, c_{\theta} \end{bmatrix}\right) \\[10pt]
\textbf{4. the tip measured from the scene centre} & & \\[3pt]
\begin{bmatrix} R_{j} \\ \Phi_{j} \\ \Theta_{j} \end{bmatrix}
  &=& \textbf{cart_to_polar}\left(\text{tip}_{j} - O\right) \\[10pt]
\textbf{5. back to a scene index} & & \\[3pt]
\begin{bmatrix} \text{tip}_{j\,\varphi} \\ \text{tip}_{j\,\theta} \\ \text{tip}_{j\,r} \end{bmatrix}
  &=& \begin{bmatrix} \operatorname{round}(\Phi_{j} / c_{\varphi}) \bmod T_{\varphi} \\[3pt]
                      \operatorname{round}(\Theta_{j} / c_{\theta}) \bmod T_{\theta} \\[3pt]
                      \operatorname{round}(R_{j} / c_{r}) \end{bmatrix} \\[10pt]
\textbf{6. what uses it} & & \\[3pt]
\text{the start}_{j} \text{ arrow} &:& \text{drawn from } C_{n} \text{ to } O + \textbf{polar_to_cart}\left(\begin{bmatrix} \text{tip}_{j\,r}\, c_{r} \\ \text{tip}_{j\,\varphi}\, c_{\varphi} \\ \text{tip}_{j\,\theta}\, c_{\theta} \end{bmatrix}\right) \\[6pt]
\begin{bmatrix} I_{j\,\varphi} \\ I_{j\,\theta} \\ I_{j\,r} \end{bmatrix}
  &=& \begin{bmatrix} \text{tip}_{j\,\varphi} \\ \text{tip}_{j\,\theta} \\ \text{tip}_{j\,r} \end{bmatrix}
      \quad \text{if } k_{j} = 1
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
