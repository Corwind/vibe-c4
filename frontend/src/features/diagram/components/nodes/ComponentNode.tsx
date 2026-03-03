import { Handle, Position } from "@xyflow/react";
import type { NodeProps } from "@xyflow/react";

interface ComponentNodeData {
  label: string;
  description?: string;
  technology?: string;
  kind?: string;
  [key: string]: unknown;
}

export function ComponentNode({ data }: NodeProps) {
  const nodeData = data as ComponentNodeData;
  return (
    <div className="min-w-[160px] max-w-[220px] rounded-lg border-2 border-blue-300 bg-blue-200 px-3 py-2 text-blue-900 shadow">
      <Handle type="target" position={Position.Top} className="!bg-blue-400" />
      <div className="mb-1 flex items-center gap-2">
        <svg
          className="h-4 w-4 shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
          />
        </svg>
        <span className="text-xs font-bold">{nodeData.label}</span>
      </div>
      {nodeData.description && (
        <p className="text-xs leading-tight text-blue-700">
          {nodeData.description}
        </p>
      )}
      {nodeData.kind && (
        <p className="mt-1 text-xs italic text-blue-500">[{nodeData.kind}]</p>
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!bg-blue-400"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="!bg-blue-400"
      />
      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="!bg-blue-400"
      />
    </div>
  );
}
