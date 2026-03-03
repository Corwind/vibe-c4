import { PageLayout } from "@/components/layout";
import { env } from "@/config/env";

const features = [
  {
    title: "React 19",
    description:
      "The latest version of React with concurrent features and improved performance.",
  },
  {
    title: "TypeScript",
    description:
      "Full type safety across the entire codebase with strict mode enabled.",
  },
  {
    title: "Tailwind CSS v4",
    description:
      "Utility-first CSS framework with the new CSS-based configuration.",
  },
  {
    title: "TanStack Query",
    description:
      "Powerful data fetching and caching for seamless server state management.",
  },
  {
    title: "Zustand",
    description:
      "Lightweight and flexible client state management with minimal boilerplate.",
  },
  {
    title: "Vitest",
    description:
      "Fast, Vite-native testing framework with built-in coverage and mocking.",
  },
];

export function HomePage() {
  return (
    <PageLayout>
      <div className="py-12 text-center">
        <h1 className="text-5xl font-extrabold tracking-tight text-text">
          {env.appTitle}
        </h1>
        <p className="mx-auto mt-4 max-w-2xl text-lg text-text-muted">
          A modern React starter built with TypeScript, Tailwind CSS v4,
          TanStack Query, Zustand, and Vitest. Everything you need to ship fast.
        </p>
      </div>

      <div className="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((feature) => (
          <div
            key={feature.title}
            className="rounded-xl border border-gray-200 bg-surface p-6 shadow-sm transition-shadow hover:shadow-md"
          >
            <h3 className="text-lg font-semibold text-text">{feature.title}</h3>
            <p className="mt-2 text-sm leading-relaxed text-text-muted">
              {feature.description}
            </p>
          </div>
        ))}
      </div>
    </PageLayout>
  );
}
