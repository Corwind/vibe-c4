import { Link } from "react-router";
import { PageLayout } from "@/components/layout";

const c4Levels = [
  {
    title: "Context",
    color: "bg-blue-600",
    description: "System landscape and external actors",
  },
  {
    title: "Container",
    color: "bg-blue-400",
    description: "Services and data stores",
  },
  {
    title: "Component",
    color: "bg-blue-200",
    description: "Structs, interfaces, and packages",
  },
  {
    title: "Code",
    color: "bg-gray-400",
    description: "Functions and methods",
  },
];

export function HomePage() {
  return (
    <PageLayout>
      <div className="py-16 text-center">
        <h1 className="text-5xl font-extrabold tracking-tight text-text">
          Vibe C4
        </h1>
        <p className="mx-auto mt-4 max-w-2xl text-lg text-text-muted">
          Upload a Go project, get interactive C4 architecture diagrams.
          Automatically analyze your codebase and visualize its structure at
          every level of abstraction.
        </p>
        <div className="mt-8">
          <Link
            to="/projects"
            className="inline-flex items-center rounded-lg bg-primary px-6 py-3 text-base font-medium text-white shadow-sm transition-colors hover:bg-primary/90"
          >
            Get Started
          </Link>
        </div>
      </div>

      <div className="mt-12">
        <h2 className="mb-8 text-center text-2xl font-bold text-text">
          Four Levels of Detail
        </h2>
        <div className="flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
          {c4Levels.map((level, index) => (
            <div key={level.title} className="flex items-center gap-4">
              <div className="flex w-40 flex-col items-center rounded-xl border border-gray-200 bg-surface p-5 shadow-sm">
                <div
                  className={`mb-3 h-3 w-3 rounded-full ${level.color}`}
                />
                <span className="text-lg font-semibold text-text">
                  {level.title}
                </span>
                <span className="mt-1 text-center text-xs text-text-muted">
                  {level.description}
                </span>
              </div>
              {index < c4Levels.length - 1 && (
                <span
                  className="hidden text-2xl font-light text-gray-400 sm:block"
                  aria-hidden="true"
                >
                  {"\u2192"}
                </span>
              )}
            </div>
          ))}
        </div>
      </div>
    </PageLayout>
  );
}
