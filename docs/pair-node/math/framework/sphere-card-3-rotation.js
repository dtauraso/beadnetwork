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
\text{Rot} &=& \text{Rot}_{\theta} \cdot \text{Rot}_{\varphi} \quad \text{two independent turns, each about its own fixed axis; the identity when both are } 0 \text{, so a } 0 \text{ change moves nothing} \\[10pt]
\rlap{\textbf{Rot}_{\varphi}\textbf{ — a turn by } a_{\varphi} \textbf{ about the fixed } z \textbf{ axis, tipping } {+y} \textbf{ toward } {+x}} & & \\[3pt]
\text{Rot}_{\varphi} &=& \begin{bmatrix} \cos a_{\varphi} & \sin a_{\varphi} & 0 \\ -\sin a_{\varphi} & \cos a_{\varphi} & 0 \\ 0 & 0 & 1 \end{bmatrix}
      \quad z \text{ stays; } x, y \text{ swing on a circle; the length stays} \\[10pt]
\rlap{\textbf{Rot}_{\theta}\textbf{ — a turn by } a_{\theta} \textbf{ about the fixed vertical } y \textbf{ axis}} & & \\[3pt]
\text{Rot}_{\theta} &=& \begin{bmatrix} \cos a_{\theta} & 0 & -\sin a_{\theta} \\ 0 & 1 & 0 \\ \sin a_{\theta} & 0 & \cos a_{\theta} \end{bmatrix}
      \quad y \text{ stays; } x, z \text{ swing on a circle; the length stays} \\[3pt]
& & \text{turns about different axes do not commute, so when a round changes both, } \varphi \text{ turns first, then } \theta \\[10pt]
\rlap{\textbf{rotate the local vector — about } C} & & \\[3pt]
1. \;\; d &=& P - C \quad \text{the local vector: it starts at } C \text{ and ends at the node} \\[3pt]
2. \;\; d' &=& \text{Rot} \cdot d \quad \text{its start stays at } C \text{; only its end swings round } C \\[3pt]
3. \;\; P' &=& C + d' \quad \text{the new end in global coordinates} \\[3pt]
P' &=& C + \text{Rot} \cdot (P - C) \\[10pt]
\rlap{\textbf{the node's global angles — read from } P' \textbf{, output only}} & & \\[3pt]
\begin{bmatrix} R' \\ \Phi' \\ \Theta' \end{bmatrix} &=& \textbf{cart_to_polar}(P' - O) \\[3pt]
& & \text{these are not } a_{\varphi},\, a_{\theta} \text{: the node turned about } C \text{, so seen from } O \text{ it sweeps a different angle,} \\[3pt]
& & \text{and its distance } R' \text{ from } O \text{ changes; they match only when } C = O \\[3pt]
& & \text{the scene index is these rounded (the global card, step 4); never read back into the rotation} \\[10pt]
\rlap{\textbf{example — } a_{\theta} = 90^{\circ} \textbf{,  } O = (0, 0, 0) \textbf{,  } C = (1, 0, 0) \textbf{,  } P = (2, 0, 0)} & & \\[3pt]
P' &=& (1, 0, 0) + \text{Rot} \cdot (1, 0, 0) = (1, 0, 0) + (0, 0, 1) = (1, 0, 1) \\[3pt]
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
