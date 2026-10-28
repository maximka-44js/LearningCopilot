import Link from "next/link";

import type { GithubComMaximka44jsLcInternalProjectsEntityProject as Project } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityProject";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  badgeTone,
  formatDate,
  formatLabel,
  formatPercent,
  Metric,
  metricTone,
  riskTone,
  statusDotClass,
} from "@/features/projects/components/card-metrics";
import { cn } from "@/lib/utils";

export type ProjectCardData = Omit<Project, "id" | "owner_id">;

type ProjectCardProps = {
  href: string;
  project: ProjectCardData;
  className?: string;
};

export default function ProjectCard({ href, project, className }: ProjectCardProps) {
  const description = project.description?.trim() || "No description";

  return (
    <Card className={cn("transition hover:border-slate-700/80", className)}>
      <Link
        href={href}
        className="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
      >
        <CardHeader className="space-y-3">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="space-y-2">
              <CardTitle className="flex flex-wrap items-center gap-2 text-[1.4rem]">
                {project.name || "Untitled project"}
                {project.metadata?.track ? (
                  <span className="rounded-full border border-slate-800/70 bg-slate-900/60 px-2 py-0.5 text-[0.7rem] uppercase tracking-[0.2em] text-slate-300">
                    {project.metadata.track}
                  </span>
                ) : null}
              </CardTitle>
              <CardDescription className="max-w-2xl text-base leading-relaxed">
                {description}
              </CardDescription>
            </div>
            <div className="flex flex-col items-start gap-2 text-[0.95rem]">
              <span
                className={cn(
                  "inline-flex items-center gap-2 rounded-md border px-2 py-1 font-medium",
                  badgeTone(project.status)
                )}
              >
                <span
                  className={cn(
                    "h-2 w-2 rounded-full",
                    statusDotClass(project.status)
                  )}
                />
                Status: {formatLabel(project.status)}
              </span>
              <span className="rounded-md border border-slate-800/70 bg-slate-900/60 px-2 py-1 text-slate-200">
                Priority: {formatLabel(project.priority)}
              </span>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <Metric
              label="Readiness"
              value={formatPercent(project.readiness)}
              tone={metricTone(project.readiness)}
              progress={project.readiness}
            />
            <Metric
              label="Coverage"
              value={formatPercent(project.coverage)}
              tone={metricTone(project.coverage)}
              progress={project.coverage}
            />
            <Metric
              label="Risk"
              value={formatLabel(project.risk)}
              tone={riskTone(project.risk)}
            />
            <Metric
              label="Deadline"
              value={formatDate(project.deadline_at)}
              tone="slate"
            />
          </div>
          <div className="flex flex-wrap gap-6 text-[0.95rem] text-slate-400">
            <span>Created: {formatDate(project.created_at)}</span>
            <span>Updated: {formatDate(project.updated_at)}</span>
            <span>Archived: {formatDate(project.archived_at)}</span>
          </div>
        </CardContent>
      </Link>
    </Card>
  );
}
