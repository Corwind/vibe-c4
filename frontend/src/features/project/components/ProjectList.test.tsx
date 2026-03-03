import { render, screen, waitFor } from "@/test/test-utils";
import { http, HttpResponse } from "msw";
import { describe, expect, it, vi } from "vitest";
import { server } from "@/test/mocks/server";
import { ProjectList } from "./ProjectList";

describe("ProjectList", () => {
  it("shows loading spinner initially", () => {
    render(<ProjectList />);
    expect(
      screen.getByRole("status", { name: /loading/i }),
    ).toBeInTheDocument();
  });

  it("renders list of projects after loading", async () => {
    render(<ProjectList />);

    await waitFor(() => {
      expect(screen.getByText("my-go-project")).toBeInTheDocument();
    });
    expect(screen.getByText("another-project")).toBeInTheDocument();
  });

  it("displays project status badges", async () => {
    render(<ProjectList />);

    await waitFor(() => {
      expect(screen.getByText("completed")).toBeInTheDocument();
    });
    expect(screen.getByText("analyzing")).toBeInTheDocument();
  });

  it("shows View Diagram button for completed projects", async () => {
    render(<ProjectList onSelectProject={() => {}} />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /view diagram/i }),
      ).toBeInTheDocument();
    });
  });

  it("calls onSelectProject when View Diagram is clicked", async () => {
    const onSelect = vi.fn();
    render(<ProjectList onSelectProject={onSelect} />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /view diagram/i }),
      ).toBeInTheDocument();
    });

    screen.getByRole("button", { name: /view diagram/i }).click();
    expect(onSelect).toHaveBeenCalledWith("proj-1");
  });

  it("handles error state", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/projects", () => {
        return HttpResponse.json(
          { message: "Internal Server Error" },
          { status: 500 },
        );
      }),
    );

    render(<ProjectList />);

    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });
  });

  it("shows empty state when no projects exist", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/projects", () => {
        return HttpResponse.json({ data: [] });
      }),
    );

    render(<ProjectList />);

    await waitFor(() => {
      expect(screen.getByText(/no projects found/i)).toBeInTheDocument();
    });
  });
});
