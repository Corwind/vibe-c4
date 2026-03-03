import { useParams } from "react-router";
import { ReactFlowProvider } from "@xyflow/react";
import { DiagramViewer } from "./DiagramViewer";
import { Breadcrumbs } from "./Breadcrumbs";
import { DetailPanel } from "./DetailPanel";
import { useDiagram, useDiagramNavigation } from "../hooks/useDiagram";
import { useProjectStore } from "@/features/project/stores/project.store";
import { useEffect, useCallback } from "react";
import type { C4NodeData } from "@/features/project/types/project.types";

export function DiagramPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const setCurrentProject = useProjectStore((s) => s.setCurrentProject);
  const selectedNodeId = useProjectStore((s) => s.selectedNodeId);
  const selectedNodeData = useProjectStore((s) => s.selectedNodeData);
  const selectedNodeType = useProjectStore((s) => s.selectedNodeType);
  const setSelectedNode = useProjectStore((s) => s.setSelectedNode);
  const clearSelectedNode = useProjectStore((s) => s.clearSelectedNode);
  const { currentLevel, navigationHistory, drillDown, navigateTo, canDrillDown } =
    useDiagramNavigation();

  useEffect(() => {
    if (projectId) {
      setCurrentProject(projectId);
    }
  }, [projectId, setCurrentProject]);

  const parentName =
    navigationHistory.length > 1
      ? navigationHistory[navigationHistory.length - 1]?.name
      : undefined;

  const { data, isLoading, error } = useDiagram(
    projectId ?? "",
    currentLevel,
    parentName,
  );

  const nodes = data?.data?.nodes ?? [];
  const edges = data?.data?.edges ?? [];

  const handleNodeClick = useCallback(
    (nodeId: string, nodeData: C4NodeData, nodeType: string) => {
      setSelectedNode(nodeId, nodeData, nodeType);
    },
    [setSelectedNode],
  );

  const handleNodeDoubleClick = useCallback(
    (nodeId: string, nodeLabel: string) => {
      if (canDrillDown) {
        drillDown(nodeId, nodeLabel);
      }
    },
    [canDrillDown, drillDown],
  );

  return (
    <div className="flex h-[calc(100vh-65px)] flex-col">
      <div className="border-b border-gray-200 bg-white px-4 py-2">
        <Breadcrumbs items={navigationHistory} onNavigate={navigateTo} />
      </div>
      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1">
          <ReactFlowProvider>
            <DiagramViewer
              nodes={nodes}
              edges={edges}
              isLoading={isLoading}
              error={error}
              selectedNodeId={selectedNodeId}
              onNodeClick={handleNodeClick}
              onNodeDoubleClick={handleNodeDoubleClick}
              onPaneClick={clearSelectedNode}
            />
          </ReactFlowProvider>
        </div>
        <DetailPanel
          nodeData={selectedNodeData}
          nodeType={selectedNodeType}
          onClose={clearSelectedNode}
        />
      </div>
    </div>
  );
}
