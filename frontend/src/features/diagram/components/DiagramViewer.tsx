import { useCallback, useMemo } from "react";
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  BackgroundVariant,
  type Node,
  type Edge,
  type NodeMouseHandler,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { SystemNode } from "./nodes/SystemNode";
import { ContainerNode } from "./nodes/ContainerNode";
import { ComponentNode } from "./nodes/ComponentNode";
import { CodeNode } from "./nodes/CodeNode";
import { ExternalSystemNode } from "./nodes/ExternalSystemNode";
import { PersonNode } from "./nodes/PersonNode";
import { C4Edge } from "./edges/C4Edge";
import { LoadingSpinner } from "@/components/common";

const nodeTypes = {
  system: SystemNode,
  container: ContainerNode,
  component: ComponentNode,
  code: CodeNode,
  external: ExternalSystemNode,
  person: PersonNode,
};

const edgeTypes = {
  c4: C4Edge,
};

interface DiagramViewerProps {
  nodes: Node[];
  edges: Edge[];
  isLoading?: boolean;
  error?: Error | null;
  onNodeDoubleClick?: (nodeId: string, nodeLabel: string) => void;
}

export function DiagramViewer({
  nodes,
  edges,
  isLoading,
  error,
  onNodeDoubleClick,
}: DiagramViewerProps) {
  const handleNodeDoubleClick: NodeMouseHandler = useCallback(
    (_event, node) => {
      if (onNodeDoubleClick) {
        const label =
          (node.data as { label?: string }).label ?? node.id;
        onNodeDoubleClick(node.id, label);
      }
    },
    [onNodeDoubleClick],
  );

  const defaultEdgeOptions = useMemo(
    () => ({
      type: "c4" as const,
      animated: false,
    }),
    [],
  );

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-full items-center justify-center" role="alert">
        <div className="text-center">
          <p className="text-lg font-semibold text-red-600">
            Failed to load diagram
          </p>
          <p className="mt-1 text-sm text-gray-500">{error.message}</p>
        </div>
      </div>
    );
  }

  if (nodes.length === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-gray-500">No diagram data available.</p>
      </div>
    );
  }

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      defaultEdgeOptions={defaultEdgeOptions}
      onNodeDoubleClick={handleNodeDoubleClick}
      fitView
      fitViewOptions={{ padding: 0.2 }}
      minZoom={0.1}
      maxZoom={2}
    >
      <Background variant={BackgroundVariant.Dots} gap={16} size={1} />
      <Controls />
      <MiniMap
        nodeStrokeWidth={3}
        zoomable
        pannable
        className="!bg-gray-50"
      />
    </ReactFlow>
  );
}
