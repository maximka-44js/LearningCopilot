"use client";

import ProjectsList, {
  type ProjectsListItem,
} from "@/features/projects/components/projects-list";
import { useListProjects } from "@/generated/api/projects/projects";

export default function ProjectsListContainer() {
  const { data, isLoading, isError } = useListProjects();

  if (isLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Загрузка проектов...
      </div>
    );
  }

  const rawProjects = Array.isArray(data)
    ? data
    : data?.status === 200
      ? data.data
      : null;

  if (isError || !rawProjects) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Не удалось загрузить проекты. Попробуйте ещё раз позже.
      </div>
    );
  }

  const items: ProjectsListItem[] = rawProjects.map((project) => ({
    project: {
      name: project.name,
      description: project.description,
      deadline_at: project.deadline_at,
      status: project.status,
      priority: project.priority,
      readiness: project.readiness,
      coverage: project.coverage,
      risk: project.risk,
      metadata: project.metadata,
      created_at: project.created_at,
      updated_at: project.updated_at,
      archived_at: project.archived_at,
    },
    href: `/projects/${project.id ?? ""}`,
  }));

  return <ProjectsList items={items} />;
}
