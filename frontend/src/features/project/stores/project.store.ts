import { create } from "zustand";
import type { DiagramLevel, C4NodeData } from "../types/project.types";

interface BreadcrumbEntry {
  level: DiagramLevel;
  label: string;
  id?: string;
}

interface ProjectStore {
  currentProjectId: string | null;
  currentLevel: DiagramLevel;
  navigationHistory: BreadcrumbEntry[];
  selectedNodeId: string | null;
  selectedNodeData: C4NodeData | null;
  selectedNodeType: string | null;
  setCurrentProject: (id: string | null) => void;
  setCurrentLevel: (level: DiagramLevel) => void;
  pushNavigation: (entry: BreadcrumbEntry) => void;
  navigateTo: (index: number) => void;
  resetNavigation: () => void;
  setSelectedNode: (
    nodeId: string,
    nodeData: C4NodeData,
    nodeType: string,
  ) => void;
  clearSelectedNode: () => void;
}

export const useProjectStore = create<ProjectStore>()((set) => ({
  currentProjectId: null,
  currentLevel: "context",
  navigationHistory: [{ level: "context", label: "System Context" }],
  selectedNodeId: null,
  selectedNodeData: null,
  selectedNodeType: null,

  setCurrentProject: (id) =>
    set({
      currentProjectId: id,
      currentLevel: "context",
      navigationHistory: [{ level: "context", label: "System Context" }],
      selectedNodeId: null,
      selectedNodeData: null,
      selectedNodeType: null,
    }),

  setCurrentLevel: (level) => set({ currentLevel: level }),

  pushNavigation: (entry) =>
    set((state) => ({
      currentLevel: entry.level,
      navigationHistory: [...state.navigationHistory, entry],
      selectedNodeId: null,
      selectedNodeData: null,
      selectedNodeType: null,
    })),

  navigateTo: (index) =>
    set((state) => {
      const entry = state.navigationHistory[index];
      if (!entry) return state;
      return {
        currentLevel: entry.level,
        navigationHistory: state.navigationHistory.slice(0, index + 1),
        selectedNodeId: null,
        selectedNodeData: null,
        selectedNodeType: null,
      };
    }),

  resetNavigation: () =>
    set({
      currentLevel: "context",
      navigationHistory: [{ level: "context", label: "System Context" }],
      selectedNodeId: null,
      selectedNodeData: null,
      selectedNodeType: null,
    }),

  setSelectedNode: (nodeId, nodeData, nodeType) =>
    set({
      selectedNodeId: nodeId,
      selectedNodeData: nodeData,
      selectedNodeType: nodeType,
    }),

  clearSelectedNode: () =>
    set({
      selectedNodeId: null,
      selectedNodeData: null,
      selectedNodeType: null,
    }),
}));
