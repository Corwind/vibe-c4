import { Handle, Position } from "@xyflow/react";
import type { NodeProps } from "@xyflow/react";

interface SystemNodeData {
  label: string;
  description?: string;
  technology?: string;
  kind?: string;
  [key: string]: unknown;
}

export function SystemNode({ data }: NodeProps) {
  const nodeData = data as SystemNodeData;
  return (
    <div className="min-w-[200px] max-w-[280px] rounded-lg border-2 border-blue-700 bg-blue-600 px-4 py-3 text-white shadow-lg">
      <Handle type="target" position={Position.Top} className="!bg-blue-800" />
      <div className="mb-1 flex items-center gap-2">
        <svg
          className="h-5 w-5 shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
          />
        </svg>
        <span className="text-sm font-bold">{nodeData.label}</span>
      </div>
      {nodeData.description && (
        <p className="text-xs leading-tight text-blue-100">
          {nodeData.description}
        </p>
      )}
      {nodeData.technology && (
        <p className="mt-1 text-xs italic text-blue-200">
          [{nodeData.technology}]
        </p>
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!bg-blue-800"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="!bg-blue-800"
      />
      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="!bg-blue-800"
      />
    </div>
  );
}
