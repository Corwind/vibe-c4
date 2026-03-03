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
  componentsDiagram: (id: string, containerName: string) =>
    ["projects", id, "diagram", "containers", containerName, "components"] as const,
  codeDiagram: (id: string, componentName: string) =>
    ["projects", id, "diagram", "components", componentName, "code"] as const,
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
  containerName: string,
) {
  return useQuery({
    queryKey: projectKeys.componentsDiagram(projectId, containerName),
    queryFn: () => fetchComponentsDiagram(projectId, containerName),
    enabled: !!projectId && !!containerName,
  });
}

export function useCodeDiagram(projectId: string, componentName: string) {
  return useQuery({
    queryKey: projectKeys.codeDiagram(projectId, componentName),
    queryFn: () => fetchCodeDiagram(projectId, componentName),
    enabled: !!projectId && !!componentName,
  });
}
