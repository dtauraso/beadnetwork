import { type NodeRuleRow, type RuleHolder, type EdgePartner } from "./node-rules";

function holdersEqual(a: RuleHolder[], b: RuleHolder[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (!ai || !bi) return false;
    if (ai.holderRow !== bi.holderRow || ai.holderLabel !== bi.holderLabel || ai.r !== bi.r) {
      return false;
    }
  }
  return true;
}

function partnersEqual(a: EdgePartner[], b: EdgePartner[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (!ai || !bi) return false;
    if (
      ai.otherRow !== bi.otherRow ||
      ai.otherLabel !== bi.otherLabel ||
      ai.incoming !== bi.incoming ||
      ai.r !== bi.r ||
      ai.edgeRow !== bi.edgeRow ||
      ai.active !== bi.active
    ) {
      return false;
    }
  }
  return true;
}

export function ruleRowsEqual(a: NodeRuleRow[], b: NodeRuleRow[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (!ai || !bi) return false;
    if (
      ai.row !== bi.row ||
      ai.label !== bi.label ||
      ai.kind !== bi.kind ||
      ai.hasRule !== bi.hasRule ||
      ai.hasKindRule !== bi.hasKindRule ||
      ai.kindActive !== bi.kindActive ||
      ai.active !== bi.active ||
      ai.phiLocked !== bi.phiLocked ||
      ai.maxThetaDeg !== bi.maxThetaDeg ||
      ai.hasSelfRule !== bi.hasSelfRule ||
      ai.selfActive !== bi.selfActive ||
      ai.selfPhiLocked !== bi.selfPhiLocked ||
      ai.selfMaxThetaDeg !== bi.selfMaxThetaDeg ||
      ai.groupId !== bi.groupId ||
      ai.groupSize !== bi.groupSize ||
      !holdersEqual(ai.holders, bi.holders) ||
      !partnersEqual(ai.partners, bi.partners)
    ) {
      return false;
    }
  }
  return true;
}
