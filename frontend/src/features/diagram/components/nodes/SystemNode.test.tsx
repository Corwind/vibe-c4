import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ReactFlowProvider } from "@xyflow/react";
import { SystemNode } from "./SystemNode";

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
    type: "system",
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
      <SystemNode {...props} />
    </ReactFlowProvider>,
  );
}

describe("SystemNode", () => {
  it("renders the label", () => {
    renderNode({ label: "My System" });
    expect(screen.getByText("My System")).toBeInTheDocument();
  });

  it("renders the description when provided", () => {
    renderNode({ label: "My System", description: "A great system" });
    expect(screen.getByText("A great system")).toBeInTheDocument();
  });

  it("renders technology when provided", () => {
    renderNode({ label: "My System", technology: "Go" });
    expect(screen.getByText("[Go]")).toBeInTheDocument();
  });

  it("does not render description when not provided", () => {
    const { container } = renderNode({ label: "My System" });
    expect(container.querySelector("p")).toBeNull();
  });

  it("renders connection handles", () => {
    renderNode({ label: "My System" });
    expect(screen.getByTestId("handle-top")).toBeInTheDocument();
    expect(screen.getByTestId("handle-bottom")).toBeInTheDocument();
    expect(screen.getByTestId("handle-left")).toBeInTheDocument();
    expect(screen.getByTestId("handle-right")).toBeInTheDocument();
  });
});
