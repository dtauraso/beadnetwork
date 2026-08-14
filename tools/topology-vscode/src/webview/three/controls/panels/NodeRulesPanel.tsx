import { useState } from "react";
import { createPortal } from "react-dom";
import { postGoRecord } from "../../../vscode-api";
import {
  encodeNodeOrbitPhiToggle,
  encodeNodeOrbitMaxTheta,
  encodeNodeOrbitActiveToggle,
} from "../../../../schema/input/input-encode";
import { useNodeRuleRows, type NodeRuleRow, type RuleHolder } from "../flags/node-rules";
import { firePanelToggle, usePanelOpen } from "../pills/panel-toggle";

function ComponentLine({
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

function ThetaLine({ rule }: { rule: NodeRuleRow }) {
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
            if (Number.isFinite(deg)) postGoRecord(encodeNodeOrbitMaxTheta(rule.row, deg));
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

function HolderBlock({ rule, holder }: { rule: NodeRuleRow; holder: RuleHolder }) {
  return (
    <div className="node-rules-holder">
      <div className="node-rules-holder-name">
        orbits <span className="node-rules-node">{holder.holderLabel}</span>
      </div>
      <div className="node-rules-components">
        <ComponentLine glyph="r">fixed {Math.round(holder.r * 10) / 10}</ComponentLine>
        <ComponentLine glyph="φ" free={!rule.phiLocked}>
          <button
            className="node-rules-edit"
            onClick={() => postGoRecord(encodeNodeOrbitPhiToggle(rule.row))}
          >
            {rule.phiLocked ? "locked" : "free"}
          </button>
        </ComponentLine>
        <ThetaLine rule={rule} />
      </div>
    </div>
  );
}

function FreeNodeBlock() {
  return (
    <div className="node-rules-holder">
      <div className="node-rules-holder-name node-rules-line--free">no orbit rule</div>
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

function NodeBlock({ rule }: { rule: NodeRuleRow }) {
  const cls = ["node-rules-node-block", rule.hasRule && !rule.active ? "node-rules-node-block--inactive" : ""]
    .filter(Boolean)
    .join(" ");

  return (
    <div className={cls}>
      <div className="node-rules-node-head">
        {rule.hasRule && (
          <input
            type="checkbox"
            title="rule enforced"
            checked={rule.active}
            onChange={() => postGoRecord(encodeNodeOrbitActiveToggle(rule.row))}
          />
        )}
        <span className="node-rules-node">{rule.label}</span>
      </div>
      {!rule.hasRule || rule.holders.length === 0 ? (
        <FreeNodeBlock />
      ) : (
        rule.holders.map((h) => <HolderBlock key={h.holderRow} rule={rule} holder={h} />)
      )}
    </div>
  );
}

export function NodeRulesPanel() {
  const open = usePanelOpen("nodeRules");
  const rows = useNodeRuleRows();

  const mount = document.getElementById("node-rules-mount");
  if (!mount) return null;

  return createPortal(
    <div className="node-rules-panel">
      <button className="node-rules-toggle" onClick={() => firePanelToggle("nodeRules", open)}>
        {open ? "▾ polar rules" : "▸ polar rules"}
      </button>
      {open && (
        <div className="node-rules-body">
          {(rows ?? []).map((rule) => (
            <NodeBlock key={rule.row} rule={rule} />
          ))}
        </div>
      )}
    </div>,
    mount,
  );
}
