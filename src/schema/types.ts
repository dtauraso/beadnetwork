export type EdgeKind =
  | "chain"
  | "signal"
  | "release"
  | "streak"
  | "pointer"
  | "and-out"
  | "edge-connection"
  | "inhibit-in"
  | "any";

export const EDGE_KINDS = [
  "chain", "signal", "release", "streak",
  "pointer", "and-out", "edge-connection", "inhibit-in", "any",
] as const satisfies readonly EdgeKind[];

type MustEqual<A, B> = [A] extends [B] ? ([B] extends [A] ? true : never) : never;
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _edgeKindsParity: MustEqual<EdgeKind, (typeof EDGE_KINDS)[number]> = true;

export const DEFAULT_EDGE_KIND: EdgeKind = "signal";

export type Port = {
  name: string;
  kind: EdgeKind;
  required?: boolean;

  anchorId?: number;

  portR?: number;
};
export type StateValue = string | number;
export type SendRule = "consumeGated" | "fireAndForget";
export const SEND_RULES = ["consumeGated", "fireAndForget"] as const satisfies readonly SendRule[];
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _sendRulesParity: MustEqual<SendRule, (typeof SEND_RULES)[number]> = true;
