import { useCallback } from "react";
import {
  useContextDiagram,
  useContainersDiagram,
  useComponentsDiagram,
  useCodeDiagram,
} from "@/features/project/hooks/useProjects";
import { useProjectStore } from "@/features/project/stores/project.store";
import type { DiagramLevel } from "@/features/project/types/project.types";

export function useDiagram(
  projectId: string,
  level: DiagramLevel,
  parentId?: string,
) {
  const contextQuery = useContextDiagram(
    level === "context" ? projectId : "",
  );
  const containerQuery = useContainersDiagram(
    level === "container" ? projectId : "",
  );
  const componentQuery = useComponentsDiagram(
    level === "component" ? projectId : "",
    level === "component" ? (parentId ?? "") : "",
  );
  const codeQuery = useCodeDiagram(
    level === "code" ? projectId : "",
    level === "code" ? (parentId ?? "") : "",
  );

  switch (level) {
    case "context":
      return contextQuery;
    case "container":
      return containerQuery;
    case "component":
      return componentQuery;
    case "code":
      return codeQuery;
  }
}

export function useDiagramNavigation() {
  const {
    currentLevel,
    navigationHistory,
    pushNavigation,
    navigateTo,
    resetNavigation,
  } = useProjectStore();

  const drillDown = useCallback(
    (nodeId: string, nodeLabel: string) => {
      const nextLevel = getNextLevel(currentLevel);
      if (!nextLevel) return;

      pushNavigation({
        level: nextLevel,
        label: nodeLabel,
        id: nodeId,
      });
    },
    [currentLevel, pushNavigation],
  );

  return {
    currentLevel,
    navigationHistory,
    drillDown,
    navigateTo,
    resetNavigation,
    canDrillDown: currentLevel !== "code",
  };
}

function getNextLevel(current: DiagramLevel): DiagramLevel | null {
  switch (current) {
    case "context":
      return "container";
    case "container":
      return "component";
    case "component":
      return "code";
    case "code":
      return null;
  }
}
