const SPHERE_CARD_3_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_3_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\textbf{shared} & & \textbf{— the three nodes together} \\[3pt]
q_{j} &=& \text{quadrant } j \text{, clockwise, } j = 1 \ldots 4 \\[3pt]
q &=& \text{the quadrant number} \\[3pt]
a &=& \text{arrival 1's quadrant} \\[3pt]
b &=& \text{arrival 2's quadrant} \\[3pt]
s &=& \text{the tick mark scale} \\[3pt]
s &=& 1 \\[3pt]
12s &=& \text{1 full turn } \varphi \\[3pt]
12s &=& \text{1 full turn } \theta \\[3pt]
3s &=& \text{1 quarter turn } \varphi \\[3pt]
3s &=& \text{1 quarter turn } \theta \\[3pt]
P &=& \text{the pole numbers} \\[3pt]
P &=& \{0,\, 6,\, 12\} \\[3pt]
qt &=& \text{the quarter turn mark between two pole numbers} \\[3pt]
qt &\in& \{3,\, 9\} \\[3pt]
L_{j} &=& 1 \text{ if a link reaches node } j \text{, } 0 \text{ if not} \\[10pt]
\textbf{node 1} & & \\[3pt]
\begin{bmatrix} \text{offset}_{0\,\varphi} \\ \text{offset}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{1\,\varphi} \\ \text{offset}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 6 \\ 6 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{2\,\varphi} \\ \text{offset}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 6 \\ 6 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{3\,\varphi} \\ \text{offset}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 12 \\ 12 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{1}\text{start}_{1}q_{1\,\varphi} \\ n_{1}\text{start}_{1}q_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{1}\text{start}_{2}q_{2\,\varphi} \\ n_{1}\text{start}_{2}q_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3 \\ 3 \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{1} \\ k_{2} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 1 \\ 1 \end{bmatrix} \\[10pt]
\textbf{node 2} & & \\[3pt]
\begin{bmatrix} \text{offset}_{0\,\varphi} \\ \text{offset}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{1\,\varphi} \\ \text{offset}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{2\,\varphi} \\ \text{offset}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{3\,\varphi} \\ \text{offset}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{2}\text{start}_{1}q_{4\,\varphi} \\ n_{2}\text{start}_{1}q_{4\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 3 \\ 3 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{2}\text{start}_{2}q_{1\,\varphi} \\ n_{2}\text{start}_{2}q_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{1} \\ k_{2} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 0 \\ 1 \end{bmatrix} \\[10pt]
\textbf{node 3} & & \\[3pt]
\begin{bmatrix} \text{offset}_{0\,\varphi} \\ \text{offset}_{0\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{1\,\varphi} \\ \text{offset}_{1\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{2\,\varphi} \\ \text{offset}_{2\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{offset}_{3\,\varphi} \\ \text{offset}_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{3}\text{start}_{1}q_{3\,\varphi} \\ n_{3}\text{start}_{1}q_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 2 \\ 2 \end{bmatrix} \\[6pt]
\begin{bmatrix} n_{3}\text{start}_{2}q_{3\,\varphi} \\ n_{3}\text{start}_{2}q_{3\,\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \end{bmatrix} \\[6pt]
\begin{bmatrix} k_{1} \\ k_{2} \end{bmatrix}
  &=& \begin{bmatrix} 0 \\ 0 \end{bmatrix} \\[6pt]
\begin{bmatrix} L_{1} \\ L_{2} \\ L_{3} \end{bmatrix}
  &=& \begin{bmatrix} 1 \\ 1 \\ 0 \end{bmatrix} \\[10pt]
\text{offset}_{\varphi},\, \text{offset}_{\theta} &\in& \{0, 1, 2, 3\} \\[6pt]
p &\in& P \\[6pt]
\rlap{\textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{\varphi} \\ \text{local_arrival}_{\theta} \end{bmatrix},\; qt,\; p\right)} & & \\[3pt]
\begin{bmatrix} \text{pole}_{\varphi} \\ \text{pole}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} p + \text{offset}_{\varphi} \\[6pt]
                      p + \text{offset}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \Delta_{\varphi} \\ \Delta_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{local_arrival}_{\varphi} - \text{pole}_{\varphi} \\[6pt]
                      \text{local_arrival}_{\theta} - \text{pole}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{direction}_{\varphi} \\ \text{direction}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} -1 & \text{if } \Delta_{\varphi} \in (qt,\, \text{pole}_{\varphi}) \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      -1 & \text{if } \Delta_{\theta} \in (qt,\, \text{pole}_{\theta}) \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\rlap{\textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{\varphi} \\ \text{local_arrival}_{\theta} \end{bmatrix},\; qt,\; p\right)} & & \\[3pt]
\begin{bmatrix} \text{pole}_{\varphi} \\ \text{pole}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} p - \text{offset}_{\varphi} \\[6pt]
                      p - \text{offset}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \Delta_{\varphi} \\ \Delta_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} \text{pole}_{\varphi} - \text{local_arrival}_{\varphi} \\[6pt]
                      \text{pole}_{\theta} - \text{local_arrival}_{\theta} \end{bmatrix} \\[6pt]
\begin{bmatrix} \text{direction}_{\varphi} \\ \text{direction}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 1 & \text{if } \Delta_{\varphi} \in (qt,\, \text{pole}_{\varphi}) \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      1 & \text{if } \Delta_{\theta} \in (qt,\, \text{pole}_{\theta}) \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\rlap{\textbf{pick_one}\left(k_{1},\; k_{2},\; \begin{bmatrix} \text{arrival}_{1\,\varphi} \\ \text{arrival}_{1\,\theta} \end{bmatrix},\; \begin{bmatrix} \text{arrival}_{2\,\varphi} \\ \text{arrival}_{2\,\theta} \end{bmatrix}\right)} & & \\[3pt]
\begin{bmatrix} \text{arrival}_{\varphi} \\ \text{arrival}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} k_{1}\,\text{arrival}_{1\,\varphi} + k_{2}\,\text{arrival}_{2\,\varphi} & \text{if } k_{1} \oplus k_{2} \\[3pt]
                      0 & \text{otherwise} \\[6pt]
                      k_{1}\,\text{arrival}_{1\,\theta} + k_{2}\,\text{arrival}_{2\,\theta} & \text{if } k_{1} \oplus k_{2} \\[3pt]
                      0 & \text{otherwise} \end{bmatrix} \\[10pt]
\textbf{one node} & & \\[3pt]
\leftarrow &:& \text{the node receives the vector on a link} \\[3pt]
\rightarrow &:& \text{the node sends the vector on a link} \\[6pt]
k_{1},\, k_{2} &\in& \{0, 1\} \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix}
  &\leftarrow& \text{in on link } 1 \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}
  &\leftarrow& \text{in on link } 2 \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix}
  &=& \textbf{pick_one}\left(k_{1},\; k_{2},\; \begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; \begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}
  &=& \textbf{pick_one}\left(k_{1},\; k_{2},\; \begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; \begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix}\right) \\[10pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{0\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{0\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 3,\; 0\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{1\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{1\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 3,\; 6\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{2\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{2\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 9,\; 6\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{1}\text{direction}_{3\,\varphi} \\ \text{local_arrival}_{1}\text{direction}_{3\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{1\,\varphi} \\ \text{local_arrival}_{1\,\theta} \end{bmatrix},\; 9,\; 12\right) \\[10pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{0\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{0\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 3,\; 0\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{1\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{1\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 3,\; 6\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{2\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{2\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 9,\; 6\right) \\[6pt]
\begin{bmatrix} \text{local_arrival}_{2}\text{direction}_{3\,\varphi} \\ \text{local_arrival}_{2}\text{direction}_{3\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{local_arrival}_{2\,\varphi} \\ \text{local_arrival}_{2\,\theta} \end{bmatrix},\; 9,\; 12\right) \\[10pt]
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
  &\rightarrow& \text{out on link } 1 \\[6pt]
\begin{bmatrix} k_{2}\,\text{local_arrival}_{2\,\varphi} \\ k_{2}\,\text{local_arrival}_{2\,\theta} \end{bmatrix}
  &\rightarrow& \text{out on link } 2
\end{array}
\]`;

function sphereCard3Figure(spec) {
  const fig = document.createElement('div');
  fig.className = 'keysphere';
  const svgEl = sphere(spec);
  fig.appendChild(svgEl);
  fig.appendChild(sphereCardToggles(svgEl));
  fig.appendChild(sphereEndControls(svgEl));
  return fig;
}

function sphereCard3Math(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3]')) {
  host.appendChild(sphereCard3Figure(SPHERE_CARD_3_SPEC));
  host.appendChild(sphereCard3Math(SPHERE_CARD_3_FORMULAS));
}
