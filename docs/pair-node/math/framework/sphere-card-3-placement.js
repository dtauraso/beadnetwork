const SPHERE_CARD_3_PLACEMENT_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\rlap{\textbf{shared — each node's geometry goroutine, not its card goroutine}} & & \\[3pt]
C_{n} &=& \text{node } n\text{'s centre, a point in the scene} \\[3pt]
I_{n} &=& \text{node } n\text{'s scene index (base + drag)} \\[3pt]
V_{j} &=& \text{the vector } n \text{ sends along the path to } j\text{: } k_{j} \cdot \text{start}_{j} \text{ before round 1, then } k_{j} \cdot \text{local_arrival}_{j} \text{ unless it is 0 (0 is no change: } V_{j} \text{ keeps the last non-zero one)} \\[3pt]
\text{chan}_{n \to j} &=& \text{one buffered channel per ordered pair } n \to j \\[3pt]
\text{move} &=& (o,\; q)\text{: the node } o \text{ that started it and } o\text{'s count } q \text{ of moves it has started} \\[3pt]
\text{moved}_{j}[o] &=& \text{the last } q \text{ from } o \text{ that node } j \text{ moved in — one lookup, however long the chain} \\[3pt]
\text{in}_{j} &=& \text{the latest tip node } j \text{ holds from each partner sending to it} \\[3pt]
\leftarrow &:& \text{the node receives on a placement channel} \\[3pt]
\rightarrow &:& \text{the node sends on a placement channel} \\[10pt]
\rlap{\textbf{when node } n \textbf{ sends}} & & \\[3pt]
& & \text{at load and on the start button: node 1 only — one walk lays out the chain; nodes 2 and 3 hold their links and pass it on} \\[3pt]
& & \text{after every round} \\[3pt]
& & \text{when } n\text{'s start, } t\text{, } k \text{ or } s \text{ changes} \\[3pt]
& & \text{when } C_{n} \text{ moves} \\[10pt]
\rlap{\textbf{send}(n,\; \text{move})} & & \\[3pt]
\text{move} &=& (n,\; q_{n} + 1) \quad \text{if } n \text{ starts it (no partner's tip moved } n\text{)} \\[3pt]
& & \text{send nothing} \quad \text{if } o = n \text{ (the move came back to the node that started it)} \\[3pt]
\text{moved}_{n}[o] &=& q \\[3pt]
\text{for each partner } j &:& k_{j} = 1 \\[3pt]
\text{tip}_{j} &=& C_{n} + \textbf{polar_to_cart}(V_{j}) \\[3pt]
(\text{tip}_{j},\; \text{move}) &\rightarrow& \text{chan}_{n \to j} \\[6pt]
\text{for each partner } j &:& k_{j} \text{ went from } 1 \text{ to } 0 \\[3pt]
\text{release} &\rightarrow& \text{chan}_{n \to j} \\[10pt]
\rlap{\textbf{receive}(j)} & & \\[3pt]
(\text{tip},\; (o,\, q)) &\leftarrow& \text{chan}_{n \to j} \quad \text{in}_{j}[n] = \text{tip} \\[3pt]
& & \quad \text{ignored if } o \ne j \text{ and } \text{moved}_{j}[o] \ge q \text{ (} j \text{ already moved in this move)} \\[3pt]
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
4. & & \textbf{send}(j,\; \text{the picked tip's move}) \quad \text{if } C_{j} \text{ moved} \\[10pt]
\rlap{\textbf{loop — } k \textbf{ links that return to where they started}} & & \\[3pt]
& & \text{a move reaches a node again only at } o\text{, which takes the tip and sends nothing further} \\[3pt]
& & \text{so each node moves once per move, and } o \text{ ends at its predecessor's tip} \\[3pt]
& & \text{each hop is } O(1)\text{: one lookup in } \text{moved}_{j}\text{, nothing carried that grows} \\[3pt]
\sum V &=& 0 \text{ around the loop: every link meets} \\[3pt]
\sum V &\ne& 0 \text{: one link stays open, } o\text{'s own outgoing vector} \\[10pt]
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
