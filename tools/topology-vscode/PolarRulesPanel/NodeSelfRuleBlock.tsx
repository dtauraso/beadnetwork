import { useState } from "react";
import { postGoRecord } from "../src/webview/vscode-api";
import {
  encodeNodeSelfDragPhiToggle,
  encodeNodeSelfDragMaxTheta,
  encodeNodeSelfDragActiveToggle,
  encodeNodeSelfDragRToggle,
} from "../src/schema/input/input-encode";
import { type NodeRuleRow } from "./node-rules";
import { ComponentLine } from "./NodeRuleBlocks";
import { formatPi, formatPiDraft, parsePiDraft } from "../src/webview/three/controls/panels/pi-fraction";

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
            const turns = parsePiDraft(draft);
            if (Number.isFinite(turns)) postGoRecord(encodeNodeSelfDragMaxTheta(rule.row, turns));
            setDraft(null);
          }}
        />
        <span className="node-rules-unit">π</span>
      </ComponentLine>
    );
  }

  if (rule.selfMaxThetaPi === null) {
    return (
      <ComponentLine glyph="θ" free>
        <button className="node-rules-edit" onClick={() => setDraft("1/2")}>
          free
        </button>
      </ComponentLine>
    );
  }

  const span = formatPi(rule.selfMaxThetaPi);
  return (
    <ComponentLine glyph="θ">
      <button className="node-rules-edit" onClick={() => setDraft(formatPiDraft(rule.selfMaxThetaPi!))}>
        ∈ [−{span}, +{span}]
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
        <ComponentLine glyph="r" free={!rule.selfRLocked}>
          <button
            className="node-rules-edit"
            onClick={() => postGoRecord(encodeNodeSelfDragRToggle(rule.row))}
          >
            {rule.selfRLocked ? "held" : "free"}
          </button>
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
