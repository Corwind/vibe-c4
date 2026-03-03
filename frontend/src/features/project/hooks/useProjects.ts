import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  analyzeProject,
  fetchProjects,
  fetchContextDiagram,
  fetchContainersDiagram,
  fetchComponentsDiagram,
  fetchCodeDiagram,
} from "../services/project.api";
import type { AnalyzeProjectInput } from "../types/project.types";

export const projectKeys = {
  all: ["projects"] as const,
  detail: (id: string) => ["projects", id] as const,
  contextDiagram: (id: string) => ["projects", id, "diagram", "context"] as const,
  containersDiagram: (id: string) =>
    ["projects", id, "diagram", "containers"] as const,
  componentsDiagram: (id: string, containerId: string) =>
    ["projects", id, "diagram", "containers", containerId, "components"] as const,
  codeDiagram: (id: string, componentId: string) =>
    ["projects", id, "diagram", "components", componentId, "code"] as const,
};

export function useProjects() {
  return useQuery({
    queryKey: projectKeys.all,
    queryFn: fetchProjects,
  });
}

export function useAnalyzeProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: AnalyzeProjectInput) => analyzeProject(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: projectKeys.all });
    },
  });
}

export function useContextDiagram(projectId: string) {
  return useQuery({
    queryKey: projectKeys.contextDiagram(projectId),
    queryFn: () => fetchContextDiagram(projectId),
    enabled: !!projectId,
  });
}

export function useContainersDiagram(projectId: string) {
  return useQuery({
    queryKey: projectKeys.containersDiagram(projectId),
    queryFn: () => fetchContainersDiagram(projectId),
    enabled: !!projectId,
  });
}

export function useComponentsDiagram(
  projectId: string,
  containerId: string,
) {
  return useQuery({
    queryKey: projectKeys.componentsDiagram(projectId, containerId),
    queryFn: () => fetchComponentsDiagram(projectId, containerId),
    enabled: !!projectId && !!containerId,
  });
}

export function useCodeDiagram(projectId: string, componentId: string) {
  return useQuery({
    queryKey: projectKeys.codeDiagram(projectId, componentId),
    queryFn: () => fetchCodeDiagram(projectId, componentId),
    enabled: !!projectId && !!componentId,
  });
}
