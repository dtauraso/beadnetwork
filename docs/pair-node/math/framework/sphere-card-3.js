const SPHERE_CARD_3_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
s &=& \text{the tick mark scale} \\[3pt]
s &=& 1 \\[3pt]
12s &=& \text{1 full turn } \varphi \\[3pt]
12s &=& \text{1 full turn } \theta \\[3pt]
3s &=& \text{1 quarter turn } \varphi \\[3pt]
3s &=& \text{1 quarter turn } \theta \\[3pt]
P &=& \text{the pole numbers} \\[3pt]
P &=& \{0,\, 6s,\, 12s\} \\[3pt]
qt &=& \text{the quarter turn mark between two pole numbers} \\[3pt]
qt &\in& \{3s,\, 9s\} \\[3pt]
J &=& \text{the jump constant} \\[3pt]
J &=& \text{integer} > 0 \\[3pt]
L_{j} &=& 1 \text{ if a link reaches node } j \text{, } 0 \text{ if not} \\[3pt]
\text{pole_offset}_{j\,\varphi},\, \text{pole_offset}_{j\,\theta} &\in& \{0, 1s, 2s, 3s\} \\[3pt]
p &\in& P \\[3pt]
k_{j} &\in& \{0, 1\} \\[3pt]
\leftarrow &:& \text{the node receives the vector on a link} \\[3pt]
\rightarrow &:& \text{the node sends the vector on a link} \\[10pt]
\textbf{node 1} & & \\[3pt]
\begin{bmatrix} \text{start}_{2\,\varphi} \\ \text{start}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{start}_{3\,\varphi} \\ \text{start}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{2\,\varphi} \\ \text{pole_offset}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2s \\ 2s \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{3\,\varphi} \\ \text{pole_offset}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3s \\ 3s \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{2} \\ k_{3} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 1 \\ 1 \end{bmatrix} \\[10pt]
\textbf{node 2} & & \\[3pt]
\begin{bmatrix} \text{start}_{1\,\varphi} \\ \text{start}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{start}_{3\,\varphi} \\ \text{start}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{1\,\varphi} \\ \text{pole_offset}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3s \\ 3s \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{3\,\varphi} \\ \text{pole_offset}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1s \\ 1s \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{1} \\ k_{3} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 0 \\ 1 \end{bmatrix} \\[10pt]
\textbf{node 3} & & \\[3pt]
\begin{bmatrix} \text{start}_{1\,\varphi} \\ \text{start}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{start}_{2\,\varphi} \\ \text{start}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{1\,\varphi} \\ \text{pole_offset}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2s \\ 2s \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{pole_offset}_{2\,\varphi} \\ \text{pole_offset}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1s \\ 1s \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{1} \\ k_{2} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \\ 0 \end{bmatrix} \\[10pt]
\rlap{\textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{\varphi} \\ \text{local_arrival}_{\theta} \end{bmatrix},\; p,\; qt\right)} & & \\[3pt]
\begin{bmatrix} \text{pole}_{\varphi} \\ \text{pole}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} p + \text{pole_offset}_{\varphi} \\[6pt]
                      p + \text{pole_offset}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{direction}_{\varphi} \\ \text{direction}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} -J & \text{if } \text{local_arrival}_{\varphi} \in (\text{pole}_{\varphi},\, qt) \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -J & \text{if } \text{local_arrival}_{\theta} \in (\text{pole}_{\theta},\, qt) \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\rlap{\textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{\varphi} \\ \text{local_arrival}_{\theta} \end{bmatrix},\; qt,\; p\right)} & & \\[3pt]
\begin{bmatrix} \text{pole}_{\varphi} \\ \text{pole}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} p - \text{pole_offset}_{\varphi} \\[6pt]
                      p - \text{pole_offset}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{direction}_{\varphi} \\ \text{direction}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} J & \text{if } \text{local_arrival}_{\varphi} \in (qt,\, \text{pole}_{\varphi}) \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      J & \text{if } \text{local_arrival}_{\theta} \in (qt,\, \text{pole}_{\theta}) \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\rlap{\textbf{pick_one}\left(k_{1},\; k_{2},\; \begin{bmatrix} \text{arrival}_{1\,\varphi} \\ \text{arrival}_{1\,\theta} \end{bmatrix},\; \begin{bmatrix} \text{arrival}_{2\,\varphi} \\ \text{arrival}_{2\,\theta} \end{bmatrix}\right)} & & \\[3pt]
\begin{bmatrix} \text{arrival}_{\varphi} \\ \text{arrival}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} k_{1}\,\text{arrival}_{1\,\varphi} + k_{2}\,\text{arrival}_{2\,\varphi} & \text{if } k_{1} \oplus k_{2} \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      k_{1}\,\text{arrival}_{1\,\theta} + k_{2}\,\text{arrival}_{2\,\theta} & \text{if } k_{1} \oplus k_{2} \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\textbf{one node} & & \\[3pt]
\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix}
  &\leftarrow& \text{link } 1 \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}
  &\leftarrow& \text{link } 2 \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix}
  &=& \textbf{pick_one}\left(k_{1},\; k_{2},\; \begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; \begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}
  &=& \textbf{pick_one}\left(k_{1},\; k_{2},\; \begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; \begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}\right) \\[10pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{0\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{0\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 0,\; 3s\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{1\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{1\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 3s,\; 6s\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{2\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{2\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 6s,\; 9s\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{3\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{3\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 9s,\; 12s\right) \\[10pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{0\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{0\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 0,\; 3s\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{1\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{1\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 3s,\; 6s\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{2\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{2\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 6s,\; 9s\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{3\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{3\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 9s,\; 12s\right) \\[10pt]
\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix}
      \begin{array}{@{}r@{\;}l@{}}
        & \text{local_arrival}_{1}\text{direction}_{0\,\varphi} \\
        + & \text{local_arrival}_{1}\text{direction}_{1\,\varphi} \\
        + & \text{local_arrival}_{1}\text{direction}_{2\,\varphi} \\
        + & \text{local_arrival}_{1}\text{direction}_{3\,\varphi}
      \end{array} \\[10pt]
      \begin{array}{@{}r@{\;}l@{}}
        & \text{local_arrival}_{1}\text{direction}_{0\,\theta} \\
        + & \text{local_arrival}_{1}\text{direction}_{1\,\theta} \\
        + & \text{local_arrival}_{1}\text{direction}_{2\,\theta} \\
        + & \text{local_arrival}_{1}\text{direction}_{3\,\theta}
      \end{array}
      \end{bmatrix} \\[10pt]
\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix}
      \begin{array}{@{}r@{\;}l@{}}
        & \text{local_arrival}_{2}\text{direction}_{0\,\varphi} \\
        + & \text{local_arrival}_{2}\text{direction}_{1\,\varphi} \\
        + & \text{local_arrival}_{2}\text{direction}_{2\,\varphi} \\
        + & \text{local_arrival}_{2}\text{direction}_{3\,\varphi}
      \end{array} \\[10pt]
      \begin{array}{@{}r@{\;}l@{}}
        & \text{local_arrival}_{2}\text{direction}_{0\,\theta} \\
        + & \text{local_arrival}_{2}\text{direction}_{1\,\theta} \\
        + & \text{local_arrival}_{2}\text{direction}_{2\,\theta} \\
        + & \text{local_arrival}_{2}\text{direction}_{3\,\theta}
      \end{array}
      \end{bmatrix} \\[10pt]
\begin{bmatrix} k_{1}\,\text{local_arrival}_{1\,\varphi} \\ k_{1}\,\text{local_arrival}_{1\,\theta} \end{bmatrix}
  &\rightarrow& \text{link } 1 \\[6pt]
\begin{bmatrix} k_{2}\,\text{local_arrival}_{2\,\varphi} \\ k_{2}\,\text{local_arrival}_{2\,\theta} \end{bmatrix}
  &\rightarrow& \text{link } 2
\end{array}
\]`;

function sphereCard3Math(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3]')) {
  host.appendChild(sphereCard3Math(SPHERE_CARD_3_FORMULAS));
}
