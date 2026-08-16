import { useState } from "react";
import { postGoRecord } from "../../../vscode-api";
import {
  encodeNodeSelfDragPhiToggle,
  encodeNodeSelfDragMaxTheta,
  encodeNodeSelfDragActiveToggle,
} from "../../../../schema/input/input-encode";
import { type NodeRuleRow } from "../flags/node-rules";
import { ComponentLine } from "./NodeRuleBlocks";

export function SelfThetaLine({ rule }: { rule: NodeRuleRow }) {
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
            if (Number.isFinite(deg)) postGoRecord(encodeNodeSelfDragMaxTheta(rule.row, deg));
            setDraft(null);
          }}
        />
        <span className="node-rules-unit">°</span>
      </ComponentLine>
    );
  }

  if (rule.selfMaxThetaDeg === null) {
    return (
      <ComponentLine glyph="θ" free>
        <button className="node-rules-edit" onClick={() => setDraft("90")}>
          free
        </button>
      </ComponentLine>
    );
  }

  const deg = Math.round(rule.selfMaxThetaDeg * 10) / 10;
  return (
    <ComponentLine glyph="θ">
      <button className="node-rules-edit" onClick={() => setDraft(String(deg))}>
        ∈ [−{deg}°, +{deg}°]
      </button>
    </ComponentLine>
  );
}

export function SelfBlock({ rule }: { rule: NodeRuleRow }) {
  return (
    <div className={rule.selfActive ? "node-rules-holder" : "node-rules-holder node-rules-holder--inactive"}>
      <div className="node-rules-holder-name">
        <input
          type="checkbox"
          title="trim this node's own drag, about the scene centre"
          checked={rule.selfActive}
          onChange={() => postGoRecord(encodeNodeSelfDragActiveToggle(rule.row))}
        />{" "}
        {rule.label} <span className="node-rules-node">itself</span>
      </div>
      <div className="node-rules-components">
        <ComponentLine glyph="r" free={!rule.hasSelfRule}>
          {rule.hasSelfRule ? "r held" : "free"}
        </ComponentLine>
        <ComponentLine glyph="φ" free={!rule.selfPhiLocked}>
          <button
            className="node-rules-edit"
            onClick={() => postGoRecord(encodeNodeSelfDragPhiToggle(rule.row))}
          >
            {rule.selfPhiLocked ? "locked" : "free"}
          </button>
        </ComponentLine>
        <SelfThetaLine rule={rule} />
      </div>
    </div>
  );
}
