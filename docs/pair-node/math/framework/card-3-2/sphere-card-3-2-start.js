const SPHERE_CARD_3_2_START = String.raw`\textbf{one node} & & \\[3pt]
\begin{bmatrix} \text{start_direction}_{0\,\varphi} \\ \text{start_direction}_{0\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 0,\; 3s\right) \\[6pt]
\begin{bmatrix} \text{start_direction}_{1\,\varphi} \\ \text{start_direction}_{1\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 3s,\; 6s\right) \\[6pt]
\begin{bmatrix} \text{start_direction}_{2\,\varphi} \\ \text{start_direction}_{2\,\theta} \end{bmatrix}
  &=& \textbf{dir_down}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 6s,\; 9s\right) \\[6pt]
\begin{bmatrix} \text{start_direction}_{3\,\varphi} \\ \text{start_direction}_{3\,\theta} \end{bmatrix}
  &=& \textbf{dir_up}\left(\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix},\; 9s,\; 12s\right) \\[10pt]
\begin{bmatrix} \text{start}_{\varphi} \\ \text{start}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix}
      \begin{array}{@{}r@{\;}l@{}}
        & \text{start}_{\varphi} \\
        + & \text{start_direction}_{0\,\varphi} \\
        + & \text{start_direction}_{1\,\varphi} \\
        + & \text{start_direction}_{2\,\varphi} \\
        + & \text{start_direction}_{3\,\varphi}
      \end{array} \\[10pt]
      \begin{array}{@{}r@{\;}l@{}}
        & \text{start}_{\theta} \\
        + & \text{start_direction}_{0\,\theta} \\
        + & \text{start_direction}_{1\,\theta} \\
        + & \text{start_direction}_{2\,\theta} \\
        + & \text{start_direction}_{3\,\theta}
      \end{array}
      \end{bmatrix}`;
