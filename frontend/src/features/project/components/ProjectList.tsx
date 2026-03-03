import { LoadingSpinner } from "@/components/common";
import { useProjects } from "../hooks/useProjects";
import type { AnalysisStatus } from "../types/project.types";

const statusColors: Record<AnalysisStatus, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  analyzing: "bg-blue-100 text-blue-800",
  completed: "bg-green-100 text-green-800",
  failed: "bg-red-100 text-red-800",
};

interface ProjectListProps {
  onSelectProject?: (projectId: string) => void;
}

export function ProjectList({ onSelectProject }: ProjectListProps) {
  const { data, isLoading, error } = useProjects();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className="p-4 text-red-600" role="alert">
        Failed to load projects: {error.message}
      </div>
    );
  }

  const projects = data?.data ?? [];

  if (projects.length === 0) {
    return <div className="p-4 text-gray-500">No projects found.</div>;
  }

  return (
    <ul className="space-y-3">
      {projects.map((project) => (
        <li
          key={project.id}
          className="flex items-center justify-between rounded-lg border border-gray-200 p-4 transition-shadow hover:shadow-md"
        >
          <div className="min-w-0 flex-1">
            <h3 className="truncate text-sm font-semibold text-text">
              {project.name}
            </h3>
            <p className="text-xs text-text-muted">{project.source}</p>
          </div>
          <div className="ml-4 flex items-center gap-3">
            <span
              className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${statusColors[project.status]}`}
            >
              {project.status}
            </span>
            {project.status === "completed" && onSelectProject && (
              <button
                onClick={() => onSelectProject(project.id)}
                className="rounded-md bg-primary px-3 py-1 text-xs font-medium text-white hover:bg-primary/90"
              >
                View Diagram
              </button>
            )}
          </div>
        </li>
      ))}
    </ul>
  );
}
