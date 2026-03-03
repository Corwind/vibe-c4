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
import type { C4NodeData } from "@/features/project/types/project.types";

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
  selectedNodeId?: string | null;
  onNodeClick?: (nodeId: string, nodeData: C4NodeData, nodeType: string) => void;
  onNodeDoubleClick?: (nodeId: string, nodeLabel: string) => void;
  onPaneClick?: () => void;
}

export function DiagramViewer({
  nodes,
  edges,
  isLoading,
  error,
  selectedNodeId,
  onNodeClick,
  onNodeDoubleClick,
  onPaneClick,
}: DiagramViewerProps) {
  const handleNodeClick: NodeMouseHandler = useCallback(
    (_event, node) => {
      if (onNodeClick) {
        const nodeData = node.data as C4NodeData;
        onNodeClick(node.id, nodeData, node.type ?? "system");
      }
    },
    [onNodeClick],
  );

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

  const styledNodes = useMemo(() => {
    if (!selectedNodeId) return nodes;
    return nodes.map((node) => ({
      ...node,
      style: {
        ...node.style,
        opacity: node.id === selectedNodeId ? 1 : 0.5,
      },
    }));
  }, [nodes, selectedNodeId]);

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
      nodes={styledNodes}
      edges={edges}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      defaultEdgeOptions={defaultEdgeOptions}
      onNodeClick={handleNodeClick}
      onNodeDoubleClick={handleNodeDoubleClick}
      onPaneClick={onPaneClick}
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
