"use client";

import Link from "next/link";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Metric,
  badgeTone,
  formatDate,
  formatLabel,
  formatPercent,
  metricTone,
  riskTone,
  statusDotClass,
} from "@/features/projects/components/card-metrics";
import CriticalToDeadlineTasks from "@/features/projects/components/critical-to-deadline-tasks";
import StagesByIdList from "@/features/projects/components/stages-by-id-list";
import { useGetProjectDashboard } from "@/generated/api/projects-read-model/projects-read-model";
import type { GithubComMaximka44jsLcInternalProjectsEntityProject } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityProject";
import type { GithubComMaximka44jsLcInternalProjectsUsecaseCriticalItem } from "@/generated/model/githubComMaximka44jsLcInternalProjectsUsecaseCriticalItem";
import type { InternalProjectsControllerRestapiV1GetProjectDashboardResponse } from "@/generated/model/internalProjectsControllerRestapiV1GetProjectDashboardResponse";
import { cn } from "@/lib/utils";

const ValueRow = ({ label, value }: { label: string; value: string }) => {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:items-baseline sm:justify-between">
      <dt className="text-xs uppercase tracking-wide text-slate-500">
        {label}
      </dt>
      <dd className="text-sm text-slate-200">{value}</dd>
    </div>
  );
};

export default function ProjectDashboardContainer({
  projectId,
}: {
  projectId: string;
}) {
  const { data, isLoading, isError } = useGetProjectDashboard(projectId);

  if (isLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading project dashboard...
      </div>
    );
  }

  const response = data as unknown;

  const wrapped =
    response &&
    typeof response === "object" &&
    !Array.isArray(response) &&
    "status" in response &&
    "data" in response
      ? (response as { status: number; data: unknown })
      : null;

  const dashboard =
    wrapped && wrapped.status === 200
      ? (wrapped.data as InternalProjectsControllerRestapiV1GetProjectDashboardResponse)
      : !wrapped && response && typeof response === "object" && !Array.isArray(response)
        ? (response as InternalProjectsControllerRestapiV1GetProjectDashboardResponse)
        : null;

  if (isError || !dashboard) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Unable to load the project dashboard right now.
        {wrapped && wrapped.status !== 200 ? (
          <span className="mt-2 block text-xs text-slate-400">
            API returned status {wrapped.status}.
          </span>
        ) : null}
      </div>
    );
  }

  const project: GithubComMaximka44jsLcInternalProjectsEntityProject | undefined =
    dashboard.project;
  const track =
    project && typeof project === "object" && project.metadata && typeof project.metadata === "object"
      ? project.metadata.track
      : undefined;
  const criticalItems: GithubComMaximka44jsLcInternalProjectsUsecaseCriticalItem[] =
    dashboard.criticalToDeadline ?? [];
  const stageIds = Object.keys(dashboard.riskByStage ?? {});

  return (
    <div className="space-y-8">
      <div className="space-y-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
              Project
            </p>
            <h1 className="text-3xl font-semibold">
              <span className="inline-flex flex-wrap items-center gap-3">
                {project?.name || "Project dashboard"}
                {typeof track === "string" && track.trim() ? (
                  <span className="rounded-full border border-slate-800/70 bg-slate-900/60 px-2 py-0.5 text-[0.7rem] uppercase tracking-[0.2em] text-slate-300">
                    {track}
                  </span>
                ) : null}
              </span>
            </h1>
            <p className="max-w-2xl text-sm text-slate-400">
              {project?.description?.trim() ||
                "Readiness, coverage, and critical items before the deadline."}
            </p>
          </div>

          <nav className="flex flex-wrap gap-2 text-sm">
            <Link
              href={`/projects/${projectId}/settings`}
              className="rounded-md border border-slate-800 bg-slate-950/40 px-3 py-2 text-slate-200 transition hover:border-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
            >
              Project settings
            </Link>
          </nav>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <Metric
            label="Coverage"
            value={formatPercent(dashboard.coverage)}
            tone={metricTone(dashboard.coverage)}
            progress={dashboard.coverage}
          />
          <Metric
            label="Readiness now"
            value={formatPercent(dashboard.readinessNow)}
            tone={metricTone(dashboard.readinessNow)}
            progress={dashboard.readinessNow}
          />
          <Metric
            label="Project risk"
            value={formatLabel(project?.risk)}
            tone={riskTone(project?.risk)}
          />
          <Metric
            label="Deadline"
            value={formatDate(project?.deadline_at)}
            tone="slate"
          />
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="lg:flex lg:h-[42rem] lg:flex-col">
          <CardHeader className="space-y-2">
            <CardTitle>Project details</CardTitle>
            <CardDescription>
              Core fields returned by the dashboard read-model.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 lg:lc-scrollbar lg:flex-1 lg:overflow-auto">
            <dl className="space-y-3">
              <ValueRow label="Status" value={formatLabel(project?.status)} />
              <ValueRow
                label="Priority"
                value={formatLabel(project?.priority)}
              />
              <ValueRow
                label="Track"
                value={typeof track === "string" && track.trim() ? track : "N/A"}
              />
              <ValueRow
                label="Readiness"
                value={formatPercent(project?.readiness)}
              />
              <ValueRow
                label="Coverage"
                value={formatPercent(project?.coverage)}
              />
              <ValueRow label="Risk" value={formatLabel(project?.risk)} />
              <ValueRow
                label="Created"
                value={formatDate(project?.created_at)}
              />
              <ValueRow
                label="Updated"
                value={formatDate(project?.updated_at)}
              />
              <ValueRow
                label="Archived"
                value={formatDate(project?.archived_at)}
              />
            </dl>

            {project ? (
              <div className="flex flex-wrap gap-2 text-sm">
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
            ) : null}
          </CardContent>
        </Card>
        <StagesByIdList projectId={projectId} stageIds={stageIds} />
      </div>

      <Card>
        <CardHeader className="space-y-2">
          <CardTitle>Critical to deadline</CardTitle>
          <CardDescription>
            Items that could block hitting the project deadline.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <CriticalToDeadlineTasks projectId={projectId} items={criticalItems} />
        </CardContent>
      </Card>
    </div>
  );
}
