import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi, beforeAll } from "vitest";
import { DiagramViewer } from "./DiagramViewer";
import type { Node, Edge } from "@xyflow/react";

// React Flow needs ResizeObserver in jsdom
beforeAll(() => {
  global.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
  }));
});

// Mock React Flow since it needs browser APIs not available in jsdom
vi.mock("@xyflow/react", () => ({
  ReactFlow: ({
    nodes,
    children,
  }: {
    nodes: Node[];
    children: React.ReactNode;
  }) => (
    <div data-testid="react-flow">
      <div data-testid="node-count">{nodes.length}</div>
      {children}
    </div>
  ),
  Background: () => <div data-testid="background" />,
  Controls: () => <div data-testid="controls" />,
  MiniMap: () => <div data-testid="minimap" />,
  BackgroundVariant: { Dots: "dots" },
  ReactFlowProvider: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
}));

describe("DiagramViewer", () => {
  const mockNodes: Node[] = [
    {
      id: "1",
      type: "system",
      position: { x: 0, y: 0 },
      data: { label: "Test System" },
    },
    {
      id: "2",
      type: "container",
      position: { x: 200, y: 0 },
      data: { label: "Test Container" },
    },
  ];

  const mockEdges: Edge[] = [
    { id: "e1-2", source: "1", target: "2", label: "uses" },
  ];

  it("renders the ReactFlow canvas with nodes", () => {
    render(<DiagramViewer nodes={mockNodes} edges={mockEdges} />);

    expect(screen.getByTestId("react-flow")).toBeInTheDocument();
    expect(screen.getByTestId("node-count")).toHaveTextContent("2");
  });

  it("renders MiniMap, Controls, and Background", () => {
    render(<DiagramViewer nodes={mockNodes} edges={mockEdges} />);

    expect(screen.getByTestId("background")).toBeInTheDocument();
    expect(screen.getByTestId("controls")).toBeInTheDocument();
    expect(screen.getByTestId("minimap")).toBeInTheDocument();
  });

  it("shows loading spinner when isLoading is true", () => {
    render(<DiagramViewer nodes={[]} edges={[]} isLoading />);

    expect(
      screen.getByRole("status", { name: /loading/i }),
    ).toBeInTheDocument();
  });

  it("shows error message when error is present", () => {
    render(
      <DiagramViewer
        nodes={[]}
        edges={[]}
        error={new Error("Network error")}
      />,
    );

    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText("Failed to load diagram")).toBeInTheDocument();
    expect(screen.getByText("Network error")).toBeInTheDocument();
  });

  it("shows empty state when no nodes are present", () => {
    render(<DiagramViewer nodes={[]} edges={[]} />);

    expect(
      screen.getByText("No diagram data available."),
    ).toBeInTheDocument();
  });
});
