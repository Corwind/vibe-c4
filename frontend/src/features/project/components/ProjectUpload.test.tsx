import { render, screen, waitFor } from "@/test/test-utils";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "@/test/mocks/server";
import { ProjectUpload } from "./ProjectUpload";

describe("ProjectUpload", () => {
  it("renders the git URL input and submit button", () => {
    render(<ProjectUpload />);

    expect(
      screen.getByLabelText(/git repository url/i),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /analyze project/i }),
    ).toBeInTheDocument();
  });

  it("renders the drop zone", () => {
    render(<ProjectUpload />);

    expect(
      screen.getByText(/drag & drop a project archive/i),
    ).toBeInTheDocument();
  });

  it("disables submit when no input is provided", () => {
    render(<ProjectUpload />);

    expect(
      screen.getByRole("button", { name: /analyze project/i }),
    ).toBeDisabled();
  });

  it("enables submit when a git URL is entered", async () => {
    const user = userEvent.setup();
    render(<ProjectUpload />);

    const input = screen.getByLabelText(/git repository url/i);
    await user.type(input, "https://github.com/user/repo");

    expect(
      screen.getByRole("button", { name: /analyze project/i }),
    ).toBeEnabled();
  });

  it("submits a git URL for analysis", async () => {
    const user = userEvent.setup();
    render(<ProjectUpload />);

    const input = screen.getByLabelText(/git repository url/i);
    await user.type(input, "https://github.com/user/repo");
    await user.click(
      screen.getByRole("button", { name: /analyze project/i }),
    );

    await waitFor(() => {
      expect(
        screen.getByText(/project analysis started successfully/i),
      ).toBeInTheDocument();
    });
  });

  it("displays selected file name after file selection", async () => {
    const user = userEvent.setup();
    render(<ProjectUpload />);

    const file = new File(["content"], "project.zip", {
      type: "application/zip",
    });
    const fileInput = screen.getByTestId("file-input");
    await user.upload(fileInput, file);

    expect(screen.getByText(/project\.zip/i)).toBeInTheDocument();
  });

  it("shows error message when analysis fails", async () => {
    server.use(
      http.post("http://localhost:8080/api/v1/projects/analyze", () => {
        return HttpResponse.json(
          { message: "Internal Server Error" },
          { status: 500 },
        );
      }),
    );

    const user = userEvent.setup();
    render(<ProjectUpload />);

    const input = screen.getByLabelText(/git repository url/i);
    await user.type(input, "https://github.com/user/repo");
    await user.click(
      screen.getByRole("button", { name: /analyze project/i }),
    );

    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });
  });

  it("shows loading state during submission", async () => {
    server.use(
      http.post("http://localhost:8080/api/v1/projects/analyze", async () => {
        await new Promise((resolve) => setTimeout(resolve, 200));
        return HttpResponse.json(
          {
            data: {
              id: "proj-new",
              name: "new-project",
              source: "https://github.com/user/repo",
              status: "pending",
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            },
          },
          { status: 201 },
        );
      }),
    );

    const user = userEvent.setup();
    render(<ProjectUpload />);

    const input = screen.getByLabelText(/git repository url/i);
    await user.type(input, "https://github.com/user/repo");
    await user.click(
      screen.getByRole("button", { name: /analyze project/i }),
    );

    await waitFor(() => {
      expect(screen.getByText(/analyzing/i)).toBeInTheDocument();
    });
  });
});
