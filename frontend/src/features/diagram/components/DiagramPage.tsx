import { useParams } from "react-router";
import { ReactFlowProvider } from "@xyflow/react";
import { DiagramViewer } from "./DiagramViewer";
import { Breadcrumbs } from "./Breadcrumbs";
import { useDiagram, useDiagramNavigation } from "../hooks/useDiagram";
import { useProjectStore } from "@/features/project/stores/project.store";
import { useEffect } from "react";

export function DiagramPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const setCurrentProject = useProjectStore((s) => s.setCurrentProject);
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

  const handleNodeDoubleClick = (nodeId: string, nodeLabel: string) => {
    if (canDrillDown) {
      drillDown(nodeId, nodeLabel);
    }
  };

  return (
    <div className="flex h-[calc(100vh-65px)] flex-col">
      <div className="border-b border-gray-200 bg-white px-4 py-2">
        <Breadcrumbs items={navigationHistory} onNavigate={navigateTo} />
      </div>
      <div className="flex-1">
        <ReactFlowProvider>
          <DiagramViewer
            nodes={nodes}
            edges={edges}
            isLoading={isLoading}
            error={error}
            onNodeDoubleClick={handleNodeDoubleClick}
          />
        </ReactFlowProvider>
      </div>
    </div>
  );
}
