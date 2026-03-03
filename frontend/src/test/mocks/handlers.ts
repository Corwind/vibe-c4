import { http, HttpResponse } from "msw";

const mockExamples = [
  {
    id: "1",
    title: "First Example",
    description: "This is the first example",
    createdAt: "2024-01-01T00:00:00Z",
  },
  {
    id: "2",
    title: "Second Example",
    description: "This is the second example",
    createdAt: "2024-01-02T00:00:00Z",
  },
];

const mockProjects = [
  {
    id: "proj-1",
    name: "my-go-project",
    source: "https://github.com/user/my-go-project",
    status: "completed" as const,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:01:00Z",
  },
  {
    id: "proj-2",
    name: "another-project",
    source: "file://uploaded.zip",
    status: "analyzing" as const,
    createdAt: "2024-01-02T00:00:00Z",
    updatedAt: "2024-01-02T00:00:30Z",
  },
];

export const handlers = [
  // Example handlers
  http.get("http://localhost:8080/api/examples", () => {
    return HttpResponse.json({ data: mockExamples });
  }),
  http.get("http://localhost:8080/api/examples/:id", ({ params }) => {
    const example = mockExamples.find((e) => e.id === params.id);
    if (!example) {
      return HttpResponse.json({ message: "Not found" }, { status: 404 });
    }
    return HttpResponse.json({ data: example });
  }),
  http.post("http://localhost:8080/api/examples", async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    const newExample = {
      id: String(mockExamples.length + 1),
      title: body.title as string,
      description: body.description as string,
      createdAt: new Date().toISOString(),
    };
    return HttpResponse.json({ data: newExample }, { status: 201 });
  }),

  // Project handlers
  http.get("http://localhost:8080/api/v1/projects", () => {
    return HttpResponse.json({ data: mockProjects });
  }),
  http.post("http://localhost:8080/api/v1/projects/analyze", () => {
    const newProject = {
      id: "proj-new",
      name: "new-project",
      source: "https://github.com/user/new-project",
      status: "pending" as const,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    return HttpResponse.json({ data: newProject }, { status: 201 });
  }),
];
