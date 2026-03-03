import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ReactFlowProvider } from "@xyflow/react";
import { ContainerNode } from "./ContainerNode";

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
    type: "container",
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
      <ContainerNode {...props} />
    </ReactFlowProvider>,
  );
}

describe("ContainerNode", () => {
  it("renders the label", () => {
    renderNode({ label: "API Gateway" });
    expect(screen.getByText("API Gateway")).toBeInTheDocument();
  });

  it("renders the description when provided", () => {
    renderNode({ label: "API Gateway", description: "Handles routing" });
    expect(screen.getByText("Handles routing")).toBeInTheDocument();
  });

  it("renders technology when provided", () => {
    renderNode({ label: "API Gateway", technology: "Go/Chi" });
    expect(screen.getByText("[Go/Chi]")).toBeInTheDocument();
  });

  it("renders connection handles", () => {
    renderNode({ label: "API Gateway" });
    expect(screen.getByTestId("handle-top")).toBeInTheDocument();
    expect(screen.getByTestId("handle-bottom")).toBeInTheDocument();
  });
});
