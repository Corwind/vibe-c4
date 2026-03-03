import { Handle, Position } from "@xyflow/react";
import type { NodeProps } from "@xyflow/react";

interface ContainerNodeData {
  label: string;
  description?: string;
  technology?: string;
  kind?: string;
  [key: string]: unknown;
}

export function ContainerNode({ data }: NodeProps) {
  const nodeData = data as ContainerNodeData;
  return (
    <div className="min-w-[180px] max-w-[250px] rounded-lg border-2 border-blue-500 bg-blue-400 px-4 py-3 text-white shadow-md">
      <Handle type="target" position={Position.Top} className="!bg-blue-600" />
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
            d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
          />
        </svg>
        <span className="text-sm font-bold">{nodeData.label}</span>
      </div>
      {nodeData.description && (
        <p className="text-xs leading-tight text-blue-50">
          {nodeData.description}
        </p>
      )}
      {nodeData.technology && (
        <p className="mt-1 text-xs italic text-blue-100">
          [{nodeData.technology}]
        </p>
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!bg-blue-600"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="!bg-blue-600"
      />
      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="!bg-blue-600"
      />
    </div>
  );
}
