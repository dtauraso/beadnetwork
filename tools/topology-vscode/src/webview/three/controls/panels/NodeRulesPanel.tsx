import { useState } from "react";
import { createPortal } from "react-dom";
import { postGoRecord } from "../../../vscode-api";
import { encodeNodeDragActiveToggle } from "../../../../schema/input/input-encode";
import { useNodeRuleRows, type NodeRuleRow } from "../flags/node-rules";
import { firePanelToggle, usePanelOpen } from "../pills/panel-toggle";
import { NodeRuleSharedMenu } from "./NodeRuleSharedMenu";
import { FreeNodeBlock, EdgeBlock, SpanningBlock } from "./NodeRuleBlocks";

function NodeBlock({ rule, members }: { rule: NodeRuleRow; members: NodeRuleRow[] }) {
  const [menuAnchor, setMenuAnchor] = useState<DOMRect | null>(null);
  const shared = rule.groupSize > 1;
  const cls = ["node-rules-node-block", rule.hasRule && !rule.active ? "node-rules-node-block--inactive" : ""]
    .filter(Boolean)
    .join(" ");

  return (
    <div className={cls}>
      <div className="node-rules-node-head">
        {(rule.hasRule || rule.hasKindRule) && (
          <input
            type="checkbox"
            title="trim drags"
            checked={rule.active}
            onChange={() => postGoRecord(encodeNodeDragActiveToggle(rule.row))}
          />
        )}
        <span className="node-rules-node">{rule.label}</span>
        {rule.kind && <span className="node-rules-kind">· {rule.kind}</span>}
        {shared && (
          <button
            className="node-rules-shared-pill"
            title={`This rule is shared by ${rule.groupSize} nodes`}
            onClick={(e) => {
              const rect = e.currentTarget.getBoundingClientRect();
              setMenuAnchor((open) => (open ? null : rect));
            }}
          >
            ⇄ shared ×{rule.groupSize}
          </button>
        )}
      </div>
      {shared && menuAnchor && (
        <NodeRuleSharedMenu
          rule={rule}
          members={members}
          anchor={menuAnchor}
          onClose={() => setMenuAnchor(null)}
        />
      )}
      {!rule.hasRule && <FreeNodeBlock />}
      {rule.partners
        .filter((p) => !p.incoming)
        .map((p) => (
          <EdgeBlock key={p.otherRow} rule={rule} partner={p} />
        ))}
      {rule.hasKindRule && <SpanningBlock rule={rule} />}
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
            <NodeBlock
              key={rule.row}
              rule={rule}
              members={(rows ?? []).filter((r) => r.groupId === rule.groupId)}
            />
          ))}
        </div>
      )}
    </div>,
    mount,
  );
}
