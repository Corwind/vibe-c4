import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Breadcrumbs } from "./Breadcrumbs";

describe("Breadcrumbs", () => {
  const items = [
    { label: "System Context" },
    { label: "Containers" },
    { label: "Components" },
  ];

  it("renders all breadcrumb items", () => {
    render(<Breadcrumbs items={items} onNavigate={() => {}} />);

    expect(screen.getByText("System Context")).toBeInTheDocument();
    expect(screen.getByText("Containers")).toBeInTheDocument();
    expect(screen.getByText("Components")).toBeInTheDocument();
  });

  it("renders the last item as non-clickable text", () => {
    render(<Breadcrumbs items={items} onNavigate={() => {}} />);

    const lastItem = screen.getByText("Components");
    expect(lastItem.tagName).toBe("SPAN");
    expect(lastItem).toHaveAttribute("aria-current", "page");
  });

  it("renders non-last items as buttons", () => {
    render(<Breadcrumbs items={items} onNavigate={() => {}} />);

    expect(
      screen.getByRole("button", { name: "System Context" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Containers" }),
    ).toBeInTheDocument();
  });

  it("calls onNavigate with correct index when a breadcrumb is clicked", async () => {
    const onNavigate = vi.fn();
    const user = userEvent.setup();
    render(<Breadcrumbs items={items} onNavigate={onNavigate} />);

    await user.click(
      screen.getByRole("button", { name: "System Context" }),
    );

    expect(onNavigate).toHaveBeenCalledWith(0);
  });

  it("renders separators between items", () => {
    const { container } = render(
      <Breadcrumbs items={items} onNavigate={() => {}} />,
    );

    const separators = container.querySelectorAll('[aria-hidden="true"]');
    expect(separators).toHaveLength(2);
  });

  it("renders single item without separators", () => {
    const { container } = render(
      <Breadcrumbs
        items={[{ label: "System Context" }]}
        onNavigate={() => {}}
      />,
    );

    const separators = container.querySelectorAll('[aria-hidden="true"]');
    expect(separators).toHaveLength(0);
  });
});
