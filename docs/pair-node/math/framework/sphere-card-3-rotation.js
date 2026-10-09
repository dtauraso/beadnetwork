const SPHERE_CARD_3_ROTATION_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{shared — one card change, turned about a centre the user picks}} & & \\[3pt]
O &=& \text{the scene centre} \\[3pt]
S &=& \text{the nodes the user selects} \\[3pt]
C &=& \text{the rotation centre: the average of the centres of the nodes in } S \text{, computed for each rotation} \\[3pt]
P &=& \text{the rotated node's centre} \\[3pt]
\Delta\varphi,\, \Delta\theta &=& \text{the local card's change for link } j \text{, in ticks: start}_{j} + t_{j} \text{ on start, the direction sum each round} \\[3pt]
a_{\varphi} &=& \Delta\varphi \cdot (T_{\varphi} / 12s) \cdot c_{\varphi} \\[3pt]
a_{\theta} &=& \Delta\theta \cdot (T_{\theta} / 12s) \cdot c_{\theta} \\[3pt]
\rlap{\textbf{the local vector — from } C \textbf{ to the node}} & & \\[3pt]
\begin{bmatrix} r \\ \varphi \\ \theta \end{bmatrix} &=& \text{its length and its two angles about } C \text{: } \varphi \text{ from } {+y} \text{, } \theta \text{ in the } x z \text{ plane from } {+x} \\[3pt]
& & \text{taken from } \textbf{cart_to_polar}(P - C) \text{ once, when } C \text{ is set; kept from then on, never measured again} \\[10pt]
\rlap{\textbf{two separate rotations — about } C} & & \\[3pt]
\varphi &\leftarrow& \varphi + a_{\varphi} \quad \text{the } \varphi \text{ rotation: it tips the vector toward or away from } {+y} \text{; } \theta \text{ and } r \text{ stay} \\[3pt]
\theta &\leftarrow& \theta + a_{\theta} \quad \text{the } \theta \text{ rotation: it swings the vector round the vertical through } C \text{; } \varphi \text{ and } r \text{ stay} \\[3pt]
& & \text{each changes only its own angle, so neither depends on the other and their order does not matter} \\[3pt]
& & \text{a } 0 \text{ change leaves its angle as it is, so it moves nothing} \\[10pt]
\rlap{\textbf{the node's new place}} & & \\[3pt]
P' &=& C + \textbf{polar_to_cart}\left(\begin{bmatrix} r \\ \varphi \\ \theta \end{bmatrix}\right) \quad \text{the vector's start stays at } C \text{; only its end moves} \\[3pt]
& & \text{only this direction is converted, so no angle is wrapped and nothing jumps} \\[10pt]
\rlap{\textbf{the node's global angles — read from } P' \textbf{, output only}} & & \\[3pt]
\begin{bmatrix} R' \\ \Phi' \\ \Theta' \end{bmatrix} &=& \textbf{cart_to_polar}(P' - O) \\[3pt]
& & \text{these are not } a_{\varphi},\, a_{\theta} \text{: the node turned about } C \text{, so seen from } O \text{ it sweeps a different angle,} \\[3pt]
& & \text{and its distance } R' \text{ from } O \text{ changes; they match only when } C = O \\[3pt]
& & \text{the scene index is these rounded (the global card, step 4); never read back into the rotation} \\[10pt]
\rlap{\textbf{example — } a_{\theta} = 90^{\circ} \textbf{,  } O = (0, 0, 0) \textbf{,  } C = (1, 0, 0) \textbf{,  } P = (2, 0, 0)} & & \\[3pt]
\begin{bmatrix} r \\ \varphi \\ \theta \end{bmatrix} &=& \textbf{cart_to_polar}(1, 0, 0) = \begin{bmatrix} 1 \\ 90^{\circ} \\ 0^{\circ} \end{bmatrix} \;\to\; \begin{bmatrix} 1 \\ 90^{\circ} \\ 90^{\circ} \end{bmatrix} \\[3pt]
P' &=& (1, 0, 0) + \textbf{polar_to_cart}(1, 90^{\circ}, 90^{\circ}) = (1, 0, 0) + (0, 0, 1) = (1, 0, 1) \\[3pt]
\Theta: 0^{\circ} \to 45^{\circ} &,& R: 2 \to \sqrt{2} \quad \text{a local turn of } 90^{\circ} \text{ is } 45^{\circ} \text{ seen from } O \\[10pt]
\rlap{\textbf{everything not rotated}} & & \\[3pt]
& & \text{keeps its place relative to } C \text{ — it comes along as a unit} \\[3pt]
& & \text{chaining is these rotations one after another, with } C \text{ computed again for each}
\end{array}
\]`;

function sphereCard3RotationMath(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3-rotation]')) {
  host.appendChild(sphereCard3RotationMath(SPHERE_CARD_3_ROTATION_FORMULAS));
}
