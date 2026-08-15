import { useState } from "react";
import { postGoRecord } from "../../../vscode-api";
import { encodeNodeDragPhiToggle, encodeNodeDragMaxTheta } from "../../../../schema/input/input-encode";
import { type NodeRuleRow, type EdgePartner } from "../flags/node-rules";

export function ComponentLine({
  glyph,
  children,
  free,
}: {
  glyph: string;
  children: React.ReactNode;
  free?: boolean;
}) {
  return (
    <div className={free ? "node-rules-line node-rules-line--free" : "node-rules-line"}>
      <span className="node-rules-glyph">{glyph}</span>
      <span className="node-rules-value">{children}</span>
    </div>
  );
}

export function ThetaLine({ rule }: { rule: NodeRuleRow }) {
  const [draft, setDraft] = useState<string | null>(null);

  if (draft !== null) {
    return (
      <ComponentLine glyph="θ">
        <input
          className="node-rules-input"
          autoFocus
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={() => setDraft(null)}
          onKeyDown={(e) => {
            if (e.key === "Escape") {
              setDraft(null);
              return;
            }
            if (e.key !== "Enter") return;
            const deg = Number.parseFloat(draft);
            if (Number.isFinite(deg)) postGoRecord(encodeNodeDragMaxTheta(rule.row, deg));
            setDraft(null);
          }}
        />
        <span className="node-rules-unit">°</span>
      </ComponentLine>
    );
  }

  if (rule.maxThetaDeg === null) {
    return (
      <ComponentLine glyph="θ" free>
        <button className="node-rules-edit" onClick={() => setDraft("90")}>
          free
        </button>
      </ComponentLine>
    );
  }

  const deg = Math.round(rule.maxThetaDeg * 10) / 10;
  return (
    <ComponentLine glyph="θ">
      <button className="node-rules-edit" onClick={() => setDraft(String(deg))}>
        ∈ [−{deg}°, +{deg}°]
      </button>
    </ComponentLine>
  );
}

export function FreeNodeBlock() {
  return (
    <div className="node-rules-holder">
      <div className="node-rules-holder-name node-rules-line--free">drags free</div>
      <div className="node-rules-components">
        <ComponentLine glyph="r" free>
          free
        </ComponentLine>
        <ComponentLine glyph="φ" free>
          free
        </ComponentLine>
        <ComponentLine glyph="θ" free>
          free
        </ComponentLine>
      </div>
    </div>
  );
}

const KIND_AXIS_RULES: Record<string, { glyph: string; text: string }[]> = {
  Input: [{ glyph: "θ", text: "snaps to half-turns" }],
};

const KIND_SPANNING_RULES: Record<string, string[]> = {
  Input: ["out-lengths held equal"],
};

export function EdgeBlock({ rule, partner }: { rule: NodeRuleRow; partner: EdgePartner }) {
  const holder = rule.holders.find((h) => h.holderRow === partner.otherRow);
  return (
    <div className="node-rules-holder">
      <div className="node-rules-holder-name">
        {rule.label} <span className="node-rules-node">↔ {partner.otherLabel}</span>
      </div>
      <div className="node-rules-components">
        {holder ? (
          <>
            <ComponentLine glyph="r">r held {Math.round(holder.r * 10) / 10}</ComponentLine>
            <ComponentLine glyph="φ" free={!rule.phiLocked}>
              <button
                className="node-rules-edit"
                onClick={() => postGoRecord(encodeNodeDragPhiToggle(rule.row))}
              >
                {rule.phiLocked ? "locked" : "free"}
              </button>
            </ComponentLine>
            <ThetaLine rule={rule} />
          </>
        ) : (
          <ComponentLine glyph="—" free>
            no rule of its own
          </ComponentLine>
        )}
      </div>
    </div>
  );
}

export function SpanningBlock({ rule }: { rule: NodeRuleRow }) {
  const axis = KIND_AXIS_RULES[rule.kind] ?? [];
  const spanning = KIND_SPANNING_RULES[rule.kind] ?? [];
  if (axis.length === 0 && spanning.length === 0) return null;
  const count = rule.partners.length;
  return (
    <div className="node-rules-holder">
      <div className="node-rules-holder-name">
        {count === 2 ? "both edges" : `all ${count} edges`}
      </div>
      <div className="node-rules-components">
        {axis.map((a) => (
          <ComponentLine key={a.glyph + a.text} glyph={a.glyph}>
            {a.text}
          </ComponentLine>
        ))}
        {spanning.map((s) => (
          <div className="node-rules-kind-rule" key={s}>
            ⤷ {s}
          </div>
        ))}
      </div>
    </div>
  );
}
