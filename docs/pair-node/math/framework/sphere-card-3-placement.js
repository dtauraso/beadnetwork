const SPHERE_CARD_3_PLACEMENT_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{shared — each node's geometry goroutine, not its card goroutine}} & & \\[3pt]
C_{n} &=& \text{node } n\text{'s centre, a point in the scene} \\[3pt]
I_{n} &=& \text{node } n\text{'s scene index (base + drag)} \\[3pt]
V_{j} &=& k_{j} \cdot \text{start}_{j} \text{ as a local vector (the global card, step 1)} \\[3pt]
\text{chan}_{n \to j} &=& \text{one buffered channel per ordered pair } n \to j \\[3pt]
\text{path} &=& \text{the nodes a move has passed through, in order} \\[3pt]
\text{in}_{j} &=& \text{the latest tip node } j \text{ holds from each partner sending to it} \\[3pt]
\leftarrow &:& \text{the node receives on a placement channel} \\[3pt]
\rightarrow &:& \text{the node sends on a placement channel} \\[10pt]
\rlap{\textbf{when node } n \textbf{ sends}} & & \\[3pt]
& & \text{when } n \text{ starts} \\[3pt]
& & \text{when } n\text{'s start, } t\text{, } k \text{ or } s \text{ changes} \\[3pt]
& & \text{when } C_{n} \text{ moves} \\[10pt]
\rlap{\textbf{send}(n,\; \text{path})} & & \\[3pt]
\text{path}' &=& \text{path} + n \\[3pt]
\text{for each partner } j &:& k_{j} = 1 \text{ and } (j \notin \text{path} \text{ or } j = \text{path}_{0}) \\[3pt]
\text{tip}_{j} &=& C_{n} + \textbf{polar_to_cart}(V_{j}) \\[3pt]
(\text{tip}_{j},\; \text{path}') &\rightarrow& \text{chan}_{n \to j} \\[6pt]
\text{for each partner } j &:& k_{j} \text{ went from } 1 \text{ to } 0 \\[3pt]
\text{release} &\rightarrow& \text{chan}_{n \to j} \\[10pt]
\rlap{\textbf{receive}(j)} & & \\[3pt]
(\text{tip},\; \text{path}) &\leftarrow& \text{chan}_{n \to j} \quad \text{in}_{j}[n] = \text{tip} \\[3pt]
\text{release} &\leftarrow& \text{chan}_{n \to j} \quad \text{in}_{j}[n] \text{ removed} \\[10pt]
\rlap{\textbf{pick}(j)} & & \\[3pt]
\text{tip} &=& \text{in}_{j}[n] \quad \text{if } n \text{ is the only partner in } \text{in}_{j} \\[3pt]
\text{tip} &=& \text{in}_{j}[a] \quad \text{if both are in } \text{in}_{j} \text{ and } j\text{'s single } k = 1 \text{ names } a \\[3pt]
\text{tip} &=& \text{none} \quad \text{otherwise} \\[10pt]
\rlap{\textbf{one node } j} & & \\[3pt]
1. & & \textbf{receive}(j) \\[3pt]
2. & & \textbf{pick}(j) \\[3pt]
3. \;\; C_{j} &=& \text{tip} \quad \text{if a tip was picked and } C_{j} \ne \text{tip} \\[3pt]
C_{j} &=& \text{from } I_{j} \text{ (the global card, step 2)} \quad \text{if } \text{in}_{j} \text{ is empty} \\[3pt]
I_{j} &=& \text{tip rounded to the scene index — output only, never read back into } C_{j} \\[3pt]
4. & & \textbf{send}(j,\; \text{path}) \quad \text{if } C_{j} \text{ moved} \\[10pt]
\rlap{\textbf{loop — } k \textbf{ links that return to where they started}} & & \\[3pt]
& & \text{a move returns only to } \text{path}_{0}\text{, which takes the tip and sends nothing further} \\[3pt]
& & \text{so each node moves once per move, and } \text{path}_{0} \text{ ends at its predecessor's tip} \\[3pt]
\sum V &=& 0 \text{ around the loop: every link meets} \\[3pt]
\sum V &\ne& 0 \text{: one link stays open, } \text{path}_{0}\text{'s own outgoing vector} \\[10pt]
\rlap{\textbf{pointer drag of } j} & & \\[3pt]
C_{j} &=& \text{from } I_{j} \text{ (the global card, step 2)} \quad \text{until a partner sends to } j \text{ again}
\end{array}
\]`;

function sphereCard3PlacementMath(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card-3-placement]')) {
  host.appendChild(sphereCard3PlacementMath(SPHERE_CARD_3_PLACEMENT_FORMULAS));
}
