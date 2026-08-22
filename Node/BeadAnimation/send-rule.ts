export type SendRule = "consumeGated" | "fireAndForget";
export const SEND_RULES = ["consumeGated", "fireAndForget"] as const satisfies readonly SendRule[];

type MustEqual<A, B> = [A] extends [B] ? ([B] extends [A] ? true : never) : never;
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _sendRulesParity: MustEqual<SendRule, (typeof SEND_RULES)[number]> = true;
