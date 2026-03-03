import { Handle, Position } from "@xyflow/react";
import type { NodeProps } from "@xyflow/react";

interface PersonNodeData {
  label: string;
  description?: string;
  [key: string]: unknown;
}

export function PersonNode({ data }: NodeProps) {
  const nodeData = data as PersonNodeData;
  return (
    <div className="flex min-w-[120px] max-w-[180px] flex-col items-center rounded-lg border-2 border-green-800 bg-green-700 px-4 py-3 text-white shadow-lg">
      <Handle type="target" position={Position.Top} className="!bg-green-900" />
      <svg
        className="mb-1 h-8 w-8"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
        />
      </svg>
      <span className="text-sm font-bold">{nodeData.label}</span>
      {nodeData.description && (
        <p className="mt-1 text-center text-xs text-green-200">
          {nodeData.description}
        </p>
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!bg-green-900"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="!bg-green-900"
      />
      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="!bg-green-900"
      />
    </div>
  );
}
