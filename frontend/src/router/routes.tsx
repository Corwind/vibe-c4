import type { RouteObject } from "react-router";
import { RootLayout } from "@/components/layout";
import { HomePage } from "@/pages/HomePage";
import { ProjectsPage } from "@/pages/ProjectsPage";
import { DiagramPage } from "@/pages/DiagramPage";
import { NotFoundPage } from "@/pages/NotFoundPage";

export const routes: RouteObject[] = [
  {
    path: "/",
    element: <RootLayout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: "projects", element: <ProjectsPage /> },
      { path: "projects/:projectId/diagram", element: <DiagramPage /> },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
];
