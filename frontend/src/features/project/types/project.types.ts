export type AnalysisStatus =
  | "pending"
  | "analyzing"
  | "completed"
  | "failed";

export interface Project {
  id: string;
  name: string;
  source: string;
  status: AnalysisStatus;
  createdAt: string;
  updatedAt: string;
}

export type DiagramLevel = "context" | "container" | "component" | "code";

export interface C4Node {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: {
    label: string;
    description?: string;
    technology?: string;
    kind?: string;
  };
}

export interface C4Edge {
  id: string;
  source: string;
  target: string;
  label?: string;
}

export interface DiagramData {
  nodes: C4Node[];
  edges: C4Edge[];
}

export interface AnalyzeProjectInput {
  gitUrl?: string;
  file?: File;
}
