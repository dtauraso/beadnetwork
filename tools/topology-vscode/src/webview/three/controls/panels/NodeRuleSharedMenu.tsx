import { postGoRecord } from "../../../vscode-api";
import { encodeNodeDragActiveToggle } from "../../../../schema/input/input-encode";
import type { NodeRuleRow } from "../flags/node-rules";

export function NodeRuleSharedMenu({
  rule,
  members,
  anchor,
  onClose,
}: {
  rule: NodeRuleRow;
  members: NodeRuleRow[];
  anchor: DOMRect;
  onClose: () => void;
}) {
  const allOn = members.every((m) => m.active);

  return (
    <div
      className="node-rules-shared-menu"
      style={{ left: Math.round(anchor.right + 8), top: Math.round(anchor.top) }}
    >
      <div className="node-rules-shared-head">
        <span>shared by {members.length}</span>
        <button className="node-rules-shared-close" onClick={onClose}>
          ×
        </button>
      </div>
      <label className="node-rules-shared-row node-rules-shared-row--all">
        <input
          type="checkbox"
          checked={allOn}
          onChange={() => {
            for (const m of members) {
              if (m.active !== !allOn) postGoRecord(encodeNodeDragActiveToggle(m.row));
            }
          }}
        />
        <span>all nodes</span>
      </label>
      {members.map((m) => (
        <label
          key={m.row}
          className={
            m.row === rule.row
              ? "node-rules-shared-row node-rules-shared-row--self"
              : "node-rules-shared-row"
          }
        >
          <input
            type="checkbox"
            checked={m.active}
            onChange={() => postGoRecord(encodeNodeDragActiveToggle(m.row))}
          />
          <span>{m.label}</span>
        </label>
      ))}
    </div>
  );
}
