import { render, screen } from "@/test/test-utils";
import { describe, expect, it } from "vitest";
import { HomePage } from "./HomePage";

describe("HomePage", () => {
  it("renders the hero heading", () => {
    render(<HomePage />);

    expect(
      screen.getByRole("heading", { level: 1, name: /vibe c4/i }),
    ).toBeInTheDocument();
  });

  it("renders the tagline describing the product", () => {
    render(<HomePage />);

    expect(
      screen.getByText(/upload a go project.*c4 architecture diagrams/i),
    ).toBeInTheDocument();
  });

  it("renders a CTA link to the projects page", () => {
    render(<HomePage />);

    const ctaLink = screen.getByRole("link", { name: /get started/i });
    expect(ctaLink).toBeInTheDocument();
    expect(ctaLink).toHaveAttribute("href", "/projects");
  });

  it("renders the four C4 levels", () => {
    render(<HomePage />);

    expect(screen.getByText("Context")).toBeInTheDocument();
    expect(screen.getByText("Container")).toBeInTheDocument();
    expect(screen.getByText("Component")).toBeInTheDocument();
    expect(screen.getByText("Code")).toBeInTheDocument();
  });

  it("renders arrows between C4 levels", () => {
    render(<HomePage />);

    const arrows = screen.getAllByText("\u2192");
    expect(arrows).toHaveLength(3);
  });

  it("renders descriptions for each C4 level", () => {
    render(<HomePage />);

    expect(screen.getByText(/system landscape/i)).toBeInTheDocument();
    expect(screen.getByText(/services and data stores/i)).toBeInTheDocument();
    expect(screen.getByText(/structs, interfaces/i)).toBeInTheDocument();
    expect(screen.getByText(/functions and methods/i)).toBeInTheDocument();
  });
});
