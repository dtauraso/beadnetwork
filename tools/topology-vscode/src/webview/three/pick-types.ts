// pick-types.ts — the raycast pick contract, shared by every module that hands a pick
// function around.
//
// This lives in its own module because four modules need it (raw-input, scene-content,
// ThreeView, interaction-controls) and none of them owns it: the pick is performed in
// scene-content, invoked from raw-input, threaded through ThreeView, and typed into the
// interaction-controls hook signature. Homing it on any one of those made the other three
// import that module for a type alone — which put a type-only edge from raw-input back to
// interaction-controls, the one import cycle in the webview.
//
// PickFn exists so the signature is written ONCE. It previously appeared longhand in four
// files, so changing what a pick takes or returns meant four matching edits.

/** Restricts what a raycast pick is allowed to hit. Omitted entirely = pick anything. */
export interface PickOptions {
  excludeId?: string;
  nodesOnly?: boolean;
  ringOnly?: boolean;
  handholdOnly?: boolean;
  /** Restrict the pick to the buffer edge pick-halos (BUFFER_EDGE_TAG), returning the hit
   *  edge's buffer EDGE-ROW index as a decimal string. */
  edgeOnly?: boolean;
}

/** Performs one raycast at NDC coords and returns the hit entity's id, or null for a miss. */
export type PickFn = (ndcX: number, ndcY: number, opts?: PickOptions) => string | null;

/** The ref the pick function is published through: scene-content installs it, raw-input
 *  calls it. Null until scene-content has mounted and installed one. */
export type PickRef = React.MutableRefObject<PickFn | null>;
