import { PageLayout } from "@/components/layout";
import { ProjectUpload } from "./ProjectUpload";
import { ProjectList } from "./ProjectList";
import { useProjectStore } from "../stores/project.store";
import { useNavigate } from "react-router";

export function ProjectPage() {
  const setCurrentProject = useProjectStore((s) => s.setCurrentProject);
  const navigate = useNavigate();

  const handleSelectProject = (projectId: string) => {
    setCurrentProject(projectId);
    void navigate(`/projects/${projectId}/diagram`);
  };

  return (
    <PageLayout>
      <div className="mx-auto max-w-3xl space-y-8 py-8">
        <div>
          <h1 className="text-2xl font-bold text-text">Analyze a Project</h1>
          <p className="mt-1 text-sm text-text-muted">
            Provide a Git repository URL or upload a project archive to generate
            C4 diagrams.
          </p>
        </div>

        <div className="rounded-xl border border-gray-200 bg-surface p-6">
          <ProjectUpload />
        </div>

        <div>
          <h2 className="mb-4 text-lg font-semibold text-text">
            Analyzed Projects
          </h2>
          <ProjectList onSelectProject={handleSelectProject} />
        </div>
      </div>
    </PageLayout>
  );
}
