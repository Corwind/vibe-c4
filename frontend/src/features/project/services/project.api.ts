import { apiClient } from "@/lib/api-client";
import { env } from "@/config/env";
import type { ApiResponse } from "@/types/api.types";
import type { DiagramData, Project } from "../types/project.types";

export function analyzeProject(input: {
  gitUrl?: string;
  file?: File;
}): Promise<ApiResponse<Project>> {
  if (input.file) {
    const formData = new FormData();
    formData.append("file", input.file);
    return fetch(`${env.apiBaseUrl}/v1/projects/analyze`, {
      method: "POST",
      body: formData,
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to analyze project");
      return res.json() as Promise<ApiResponse<Project>>;
    });
  }
  return apiClient.post<ApiResponse<Project>>("/v1/projects/analyze", {
    gitUrl: input.gitUrl,
  });
}

export function fetchProjects(): Promise<ApiResponse<Project[]>> {
  return apiClient.get<ApiResponse<Project[]>>("/v1/projects");
}

export function fetchContextDiagram(
  projectId: string,
): Promise<ApiResponse<DiagramData>> {
  return apiClient.get<ApiResponse<DiagramData>>(
    `/v1/projects/${projectId}/diagram/context`,
  );
}

export function fetchContainersDiagram(
  projectId: string,
): Promise<ApiResponse<DiagramData>> {
  return apiClient.get<ApiResponse<DiagramData>>(
    `/v1/projects/${projectId}/diagram/containers`,
  );
}

export function fetchComponentsDiagram(
  projectId: string,
  containerName: string,
): Promise<ApiResponse<DiagramData>> {
  return apiClient.get<ApiResponse<DiagramData>>(
    `/v1/projects/${projectId}/diagram/containers/${containerName}/components`,
  );
}

export function fetchCodeDiagram(
  projectId: string,
  componentName: string,
): Promise<ApiResponse<DiagramData>> {
  return apiClient.get<ApiResponse<DiagramData>>(
    `/v1/projects/${projectId}/diagram/components/${componentName}/code`,
  );
}
