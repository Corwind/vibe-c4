import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { DetailPanel } from "./DetailPanel";
import type { C4NodeData } from "@/features/project/types/project.types";

describe("DetailPanel", () => {
  const onClose = vi.fn();

  it("renders nothing when no node is selected", () => {
    const { container } = render(
      <DetailPanel
        nodeData={null}
        nodeType={null}
        onClose={onClose}
      />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders the close button", () => {
    const data: C4NodeData = { label: "Test System" };
    render(
      <DetailPanel nodeData={data} nodeType="system" onClose={onClose} />,
    );
    expect(
      screen.getByRole("button", { name: /close/i }),
    ).toBeInTheDocument();
  });

  it("calls onClose when close button is clicked", async () => {
    const user = userEvent.setup();
    const data: C4NodeData = { label: "Test System" };
    render(
      <DetailPanel nodeData={data} nodeType="system" onClose={onClose} />,
    );

    await user.click(screen.getByRole("button", { name: /close/i }));
    expect(onClose).toHaveBeenCalledOnce();
  });

  describe("System node", () => {
    it("renders system name and description", () => {
      const data: C4NodeData = {
        label: "My System",
        description: "Main backend system",
      };
      render(
        <DetailPanel nodeData={data} nodeType="system" onClose={onClose} />,
      );

      expect(screen.getByText("My System")).toBeInTheDocument();
      expect(screen.getByText("Main backend system")).toBeInTheDocument();
      expect(screen.getByText("System")).toBeInTheDocument();
    });

    it("renders dependencies list", () => {
      const data: C4NodeData = {
        label: "My System",
        dependencies: ["Database", "Auth Service"],
      };
      render(
        <DetailPanel nodeData={data} nodeType="system" onClose={onClose} />,
      );

      expect(screen.getByText("Dependencies")).toBeInTheDocument();
      expect(screen.getByText("Database")).toBeInTheDocument();
      expect(screen.getByText("Auth Service")).toBeInTheDocument();
    });
  });

  describe("Container node", () => {
    it("renders container name, technology, and role", () => {
      const data: C4NodeData = {
        label: "API Server",
        technology: "Go/Chi",
        packagePath: "cmd/api",
        role: "cmd",
      };
      render(
        <DetailPanel
          nodeData={data}
          nodeType="container"
          onClose={onClose}
        />,
      );

      expect(screen.getByText("API Server")).toBeInTheDocument();
      expect(screen.getByText("Go/Chi")).toBeInTheDocument();
      expect(screen.getByText("cmd/api")).toBeInTheDocument();
      expect(screen.getByText("cmd")).toBeInTheDocument();
      expect(screen.getByText("Container")).toBeInTheDocument();
    });
  });

  describe("Component node", () => {
    it("renders component name, kind, methods, and fields", () => {
      const data: C4NodeData = {
        label: "UserService",
        kind: "struct",
        methods: ["GetUser(id string)", "CreateUser(input UserInput)"],
        fields: ["db *sql.DB", "logger *log.Logger"],
      };
      render(
        <DetailPanel
          nodeData={data}
          nodeType="component"
          onClose={onClose}
        />,
      );

      expect(screen.getByText("UserService")).toBeInTheDocument();
      expect(screen.getByText("struct")).toBeInTheDocument();
      expect(screen.getByText("Methods")).toBeInTheDocument();
      expect(screen.getByText("GetUser(id string)")).toBeInTheDocument();
      expect(
        screen.getByText("CreateUser(input UserInput)"),
      ).toBeInTheDocument();
      expect(screen.getByText("Fields")).toBeInTheDocument();
      expect(screen.getByText("db *sql.DB")).toBeInTheDocument();
      expect(screen.getByText("logger *log.Logger")).toBeInTheDocument();
    });

    it("does not render Methods heading when no methods", () => {
      const data: C4NodeData = {
        label: "Config",
        kind: "struct",
        fields: ["port int"],
      };
      render(
        <DetailPanel
          nodeData={data}
          nodeType="component"
          onClose={onClose}
        />,
      );

      expect(screen.queryByText("Methods")).not.toBeInTheDocument();
      expect(screen.getByText("Fields")).toBeInTheDocument();
    });
  });

  describe("Code node", () => {
    it("renders function signature, parameters, and return types", () => {
      const data: C4NodeData = {
        label: "HandleRequest",
        signature: "func HandleRequest(w http.ResponseWriter, r *http.Request)",
        parameters: ["w http.ResponseWriter", "r *http.Request"],
        returnTypes: ["error"],
      };
      render(
        <DetailPanel nodeData={data} nodeType="code" onClose={onClose} />,
      );

      expect(screen.getByText("HandleRequest")).toBeInTheDocument();
      expect(screen.getByText("Signature")).toBeInTheDocument();
      expect(
        screen.getByText(
          "func HandleRequest(w http.ResponseWriter, r *http.Request)",
        ),
      ).toBeInTheDocument();
      expect(screen.getByText("Parameters")).toBeInTheDocument();
      expect(
        screen.getByText("w http.ResponseWriter"),
      ).toBeInTheDocument();
      expect(screen.getByText("Returns")).toBeInTheDocument();
      expect(screen.getByText("error")).toBeInTheDocument();
    });
  });

  describe("External system node", () => {
    it("renders external system info", () => {
      const data: C4NodeData = {
        label: "Payment Gateway",
        description: "Third-party payment processor",
      };
      render(
        <DetailPanel
          nodeData={data}
          nodeType="external"
          onClose={onClose}
        />,
      );

      expect(screen.getByText("Payment Gateway")).toBeInTheDocument();
      expect(
        screen.getByText("Third-party payment processor"),
      ).toBeInTheDocument();
      expect(screen.getByText("External System")).toBeInTheDocument();
    });
  });

  describe("Person node", () => {
    it("renders person info", () => {
      const data: C4NodeData = {
        label: "Admin User",
        description: "System administrator",
      };
      render(
        <DetailPanel nodeData={data} nodeType="person" onClose={onClose} />,
      );

      expect(screen.getByText("Admin User")).toBeInTheDocument();
      expect(screen.getByText("System administrator")).toBeInTheDocument();
      expect(screen.getByText("Person")).toBeInTheDocument();
    });
  });
});
