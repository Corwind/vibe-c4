import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ReactFlowProvider } from "@xyflow/react";
import { ComponentNode } from "./ComponentNode";

vi.mock("@xyflow/react", async () => {
  const actual = await vi.importActual("@xyflow/react");
  return {
    ...actual,
    Handle: ({ position }: { position: string }) => (
      <div data-testid={`handle-${position}`} />
    ),
  };
});

function renderNode(data: Record<string, unknown>) {
  const props = {
    id: "test-node",
    data,
    type: "component",
    draggable: true,
    dragging: false,
    zIndex: 0,
    selectable: true,
    deletable: false,
    selected: false,
    isConnectable: true,
    positionAbsoluteX: 0,
    positionAbsoluteY: 0,
  } as const;
  return render(
    <ReactFlowProvider>
      <ComponentNode {...props} />
    </ReactFlowProvider>,
  );
}

describe("ComponentNode", () => {
  it("renders the label", () => {
    renderNode({ label: "UserService" });
    expect(screen.getByText("UserService")).toBeInTheDocument();
  });

  it("renders kind when provided", () => {
    renderNode({ label: "UserService", kind: "interface" });
    expect(screen.getByText("[interface]")).toBeInTheDocument();
  });

  it("renders description when provided", () => {
    renderNode({ label: "UserService", description: "Handles user logic" });
    expect(screen.getByText("Handles user logic")).toBeInTheDocument();
  });
});
