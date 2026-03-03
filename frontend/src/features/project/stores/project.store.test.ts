import { describe, expect, it, beforeEach } from "vitest";
import { useProjectStore } from "./project.store";

describe("useProjectStore", () => {
  beforeEach(() => {
    useProjectStore.setState({
      currentProjectId: null,
      currentLevel: "context",
      navigationHistory: [{ level: "context", label: "System Context" }],
      selectedNodeId: null,
      selectedNodeData: null,
      selectedNodeType: null,
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
      id: "main-system",
    });

    const state = useProjectStore.getState();
    expect(state.currentLevel).toBe("container");
    expect(state.navigationHistory).toHaveLength(2);
    expect(state.navigationHistory[1]).toEqual({
      level: "container",
      label: "Containers",
      id: "main-system",
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

  it("sets selected node", () => {
    const nodeData = { label: "TestNode", kind: "struct" };
    useProjectStore.getState().setSelectedNode("node-1", nodeData, "component");

    const state = useProjectStore.getState();
    expect(state.selectedNodeId).toBe("node-1");
    expect(state.selectedNodeData).toEqual(nodeData);
    expect(state.selectedNodeType).toBe("component");
  });

  it("clears selected node", () => {
    useProjectStore
      .getState()
      .setSelectedNode("node-1", { label: "Test" }, "system");
    useProjectStore.getState().clearSelectedNode();

    const state = useProjectStore.getState();
    expect(state.selectedNodeId).toBeNull();
    expect(state.selectedNodeData).toBeNull();
    expect(state.selectedNodeType).toBeNull();
  });

  it("clears selection when navigating to a different level", () => {
    useProjectStore
      .getState()
      .setSelectedNode("node-1", { label: "Test" }, "system");
    useProjectStore.getState().pushNavigation({
      level: "container",
      label: "Containers",
    });

    const state = useProjectStore.getState();
    expect(state.selectedNodeId).toBeNull();
    expect(state.selectedNodeData).toBeNull();
  });

  it("clears selection when setting a new project", () => {
    useProjectStore
      .getState()
      .setSelectedNode("node-1", { label: "Test" }, "system");
    useProjectStore.getState().setCurrentProject("proj-2");

    const state = useProjectStore.getState();
    expect(state.selectedNodeId).toBeNull();
    expect(state.selectedNodeData).toBeNull();
  });

  it("clears selection when navigating via breadcrumbs", () => {
    useProjectStore.getState().pushNavigation({
      level: "container",
      label: "Containers",
    });
    useProjectStore
      .getState()
      .setSelectedNode("node-1", { label: "Test" }, "container");
    useProjectStore.getState().navigateTo(0);

    const state = useProjectStore.getState();
    expect(state.selectedNodeId).toBeNull();
    expect(state.selectedNodeData).toBeNull();
  });

  it("clears selection on reset navigation", () => {
    useProjectStore
      .getState()
      .setSelectedNode("node-1", { label: "Test" }, "system");
    useProjectStore.getState().resetNavigation();

    const state = useProjectStore.getState();
    expect(state.selectedNodeId).toBeNull();
    expect(state.selectedNodeData).toBeNull();
  });
});
