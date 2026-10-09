const SPHERE_CARD_3_ROTATION_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{shared — one card change, turned about a centre the user picks}} & & \\[3pt]
O &=& \text{the scene centre, the zero point of global coordinates} \\[3pt]
S &=& \text{the nodes the user selects} \\[3pt]
C &=& \text{the rotation centre: the average of the centres of the nodes in } S \text{, computed for each rotation} \\[3pt]
P &=& \text{the rotated node's centre} \\[3pt]
\Delta\varphi,\, \Delta\theta &=& \text{the local card's change for link } j \text{, in ticks: start}_{j} + t_{j} \text{ on start, the direction sum each round} \\[3pt]
a_{\varphi} &=& \Delta\varphi \cdot (T_{\varphi} / 12s) \cdot c_{\varphi} \\[3pt]
a_{\theta} &=& \Delta\theta \cdot (T_{\theta} / 12s) \cdot c_{\theta} \\[3pt]
\text{Rot} &=& \text{the turn by } a_{\varphi},\, a_{\theta} \text{: the identity when both are } 0 \text{, so a } 0 \text{ change moves nothing} \\[10pt]
\rlap{\textbf{Rot}_{\theta}\textbf{ — a turn by } a_{\theta} \textbf{ about the vertical axis}} & & \\[3pt]
\text{Rot}_{\theta} &=& \begin{bmatrix} \cos a_{\theta} & 0 & -\sin a_{\theta} \\ 0 & 1 & 0 \\ \sin a_{\theta} & 0 & \cos a_{\theta} \end{bmatrix}
      \quad y \text{ stays; } x, z \text{ swing on a circle; the length stays} \\[10pt]
\rlap{\textbf{rotate the local vector — about } C} & & \\[3pt]
1. \;\; d &=& P - C \quad \text{the local vector: it starts at } C \text{ and ends at the node} \\[3pt]
2. \;\; d' &=& \text{Rot} \cdot d \quad \text{its start stays at } C \text{; only its end swings round } C \\[3pt]
3. \;\; P' &=& C + d' \quad \text{the new end in global coordinates} \\[3pt]
P' &=& C + \text{Rot} \cdot (P - C) \\[10pt]
\rlap{\textbf{not the global vector — about } O} & & \\[3pt]
P'_{O} &=& \text{Rot} \cdot P \quad \text{the same turn, about } O \\[3pt]
P' - P'_{O} &=& C + \text{Rot} \cdot P - \text{Rot} \cdot C - \text{Rot} \cdot P \\[3pt]
&=& C - \text{Rot} \cdot C \quad \text{how far } C \text{ itself would move under the turn about } O \\[3pt]
& & \text{zero only if the turn leaves } C \text{ where it is: } C = O \text{, or } C \text{ on the axis through } O \\[3pt]
& & \text{same direction, different place: the turn must be about } C \\[10pt]
\rlap{\textbf{example — } a_{\theta} = 90^{\circ} \textbf{,  } C = (1, 0, 0) \textbf{,  } P = (2, 0, 0)} & & \\[3pt]
P'_{O} &=& \text{Rot} \cdot (2, 0, 0) = (0, 0, 2) \\[3pt]
P' &=& (1, 0, 0) + \text{Rot} \cdot (1, 0, 0) = (1, 0, 0) + (0, 0, 1) = (1, 0, 1) \\[3pt]
P' - P'_{O} &=& (1, 0, -1) = (1, 0, 0) - (0, 0, 1) = C - \text{Rot} \cdot C \\[10pt]
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
