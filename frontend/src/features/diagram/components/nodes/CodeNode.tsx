import { Handle, Position } from "@xyflow/react";
import type { NodeProps } from "@xyflow/react";

interface CodeNodeData {
  label: string;
  description?: string;
  kind?: string;
  [key: string]: unknown;
}

export function CodeNode({ data }: NodeProps) {
  const nodeData = data as CodeNodeData;
  return (
    <div className="min-w-[140px] max-w-[200px] rounded border border-gray-400 bg-gray-100 px-3 py-2 text-gray-800 shadow-sm">
      <Handle type="target" position={Position.Top} className="!bg-gray-500" />
      <div className="flex items-center gap-1.5">
        <svg
          className="h-3.5 w-3.5 shrink-0 text-gray-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
        <span className="truncate font-mono text-xs font-medium">
          {nodeData.label}
        </span>
      </div>
      {nodeData.kind && (
        <p className="mt-0.5 text-xs text-gray-500">{nodeData.kind}</p>
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!bg-gray-500"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="!bg-gray-500"
      />
      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="!bg-gray-500"
      />
    </div>
  );
}
