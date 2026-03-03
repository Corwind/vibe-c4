import { describe, expect, it, beforeEach } from "vitest";
import { useProjectStore } from "./project.store";

describe("useProjectStore", () => {
  beforeEach(() => {
    useProjectStore.setState({
      currentProjectId: null,
      currentLevel: "context",
      navigationHistory: [{ level: "context", label: "System Context" }],
    });
  });

  it("sets current project and resets navigation", () => {
    useProjectStore.getState().setCurrentProject("proj-1");

    const state = useProjectStore.getState();
    expect(state.currentProjectId).toBe("proj-1");
    expect(state.currentLevel).toBe("context");
    expect(state.navigationHistory).toEqual([
      { level: "context", label: "System Context" },
    ]);
  });

  it("pushes navigation entry", () => {
    useProjectStore.getState().pushNavigation({
      level: "container",
      label: "Containers",
      name: "main-system",
    });

    const state = useProjectStore.getState();
    expect(state.currentLevel).toBe("container");
    expect(state.navigationHistory).toHaveLength(2);
    expect(state.navigationHistory[1]).toEqual({
      level: "container",
      label: "Containers",
      name: "main-system",
    });
  });

  it("navigates to a specific breadcrumb index", () => {
    useProjectStore.getState().pushNavigation({
      level: "container",
      label: "Containers",
    });
    useProjectStore.getState().pushNavigation({
      level: "component",
      label: "Components",
    });

    useProjectStore.getState().navigateTo(1);

    const state = useProjectStore.getState();
    expect(state.currentLevel).toBe("container");
    expect(state.navigationHistory).toHaveLength(2);
  });

  it("resets navigation back to context level", () => {
    useProjectStore.getState().pushNavigation({
      level: "container",
      label: "Containers",
    });
    useProjectStore.getState().pushNavigation({
      level: "component",
      label: "Components",
    });

    useProjectStore.getState().resetNavigation();

    const state = useProjectStore.getState();
    expect(state.currentLevel).toBe("context");
    expect(state.navigationHistory).toHaveLength(1);
  });

  it("does not change state when navigating to invalid index", () => {
    const before = useProjectStore.getState();
    useProjectStore.getState().navigateTo(99);
    const after = useProjectStore.getState();

    expect(after.currentLevel).toBe(before.currentLevel);
    expect(after.navigationHistory).toEqual(before.navigationHistory);
  });

  it("clears current project", () => {
    useProjectStore.getState().setCurrentProject("proj-1");
    useProjectStore.getState().setCurrentProject(null);

    expect(useProjectStore.getState().currentProjectId).toBeNull();
  });
});
