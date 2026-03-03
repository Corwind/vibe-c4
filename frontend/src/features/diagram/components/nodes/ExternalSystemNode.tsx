import { Handle, Position } from "@xyflow/react";
import type { NodeProps } from "@xyflow/react";

interface ExternalSystemNodeData {
  label: string;
  description?: string;
  technology?: string;
  [key: string]: unknown;
}

export function ExternalSystemNode({ data }: NodeProps) {
  const nodeData = data as ExternalSystemNodeData;
  return (
    <div className="min-w-[180px] max-w-[250px] rounded-lg border-2 border-dashed border-gray-400 bg-gray-50 px-4 py-3 text-gray-700 shadow-sm">
      <Handle type="target" position={Position.Top} className="!bg-gray-400" />
      <div className="mb-1 flex items-center gap-2">
        <svg
          className="h-4 w-4 shrink-0 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
          />
        </svg>
        <span className="text-sm font-bold text-gray-600">
          {nodeData.label}
        </span>
      </div>
      <p className="text-xs text-gray-500">[External System]</p>
      {nodeData.description && (
        <p className="mt-1 text-xs leading-tight text-gray-500">
          {nodeData.description}
        </p>
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!bg-gray-400"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="!bg-gray-400"
      />
      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="!bg-gray-400"
      />
    </div>
  );
}
