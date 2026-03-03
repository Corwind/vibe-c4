import { create } from "zustand";
import type { DiagramLevel } from "../types/project.types";

interface BreadcrumbEntry {
  level: DiagramLevel;
  label: string;
  name?: string;
}

interface ProjectStore {
  currentProjectId: string | null;
  currentLevel: DiagramLevel;
  navigationHistory: BreadcrumbEntry[];
  setCurrentProject: (id: string | null) => void;
  setCurrentLevel: (level: DiagramLevel) => void;
  pushNavigation: (entry: BreadcrumbEntry) => void;
  navigateTo: (index: number) => void;
  resetNavigation: () => void;
}

export const useProjectStore = create<ProjectStore>()((set) => ({
  currentProjectId: null,
  currentLevel: "context",
  navigationHistory: [{ level: "context", label: "System Context" }],

  setCurrentProject: (id) =>
    set({
      currentProjectId: id,
      currentLevel: "context",
      navigationHistory: [{ level: "context", label: "System Context" }],
    }),

  setCurrentLevel: (level) => set({ currentLevel: level }),

  pushNavigation: (entry) =>
    set((state) => ({
      currentLevel: entry.level,
      navigationHistory: [...state.navigationHistory, entry],
    })),

  navigateTo: (index) =>
    set((state) => {
      const entry = state.navigationHistory[index];
      if (!entry) return state;
      return {
        currentLevel: entry.level,
        navigationHistory: state.navigationHistory.slice(0, index + 1),
      };
    }),

  resetNavigation: () =>
    set({
      currentLevel: "context",
      navigationHistory: [{ level: "context", label: "System Context" }],
    }),
}));
