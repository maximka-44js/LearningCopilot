import Link from "next/link";

import type { GithubComMaximka44jsLcInternalProjectsEntityTask as Task } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityTask";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  badgeTone,
  formatDate,
  formatLabel,
  formatPercent,
  Metric,
  metricTone,
  taskStatusDotClass,
} from "@/features/projects/components/card-metrics";
import { cn } from "@/lib/utils";

type TaskCardProps = {
  href: string;
  task: Task;
  description?: string;
  className?: string;
};

export default function TaskCard({ href, task, description, className }: TaskCardProps) {
  const title = task.outcome?.trim() || "Task";

  return (
    <Card
      className={cn(
        "border-dashed border-slate-700/80 bg-slate-950/60 transition hover:border-slate-700/80 sm:border-slate-800/70 sm:bg-slate-950/40",
        className
      )}
    >
      <Link
        href={href}
        className="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
      >
        <CardHeader className="space-y-3 px-4 py-4 sm:px-6 sm:py-6">
          <div className="flex flex-wrap items-start justify-between gap-3 sm:gap-4">
            <div className="space-y-2">
              <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
                Task
              </p>
              <CardTitle className="text-[1.05rem] sm:text-[1.2rem]">
                {title}
              </CardTitle>
              <CardDescription className="max-w-2xl text-sm leading-relaxed sm:text-base">
                {description?.trim() || ""}
              </CardDescription>
            </div>
            <div className="flex flex-col items-start gap-2 text-[0.85rem] sm:text-[0.95rem]">
              <span className="inline-flex items-center gap-2 rounded-md border border-slate-800/70 bg-slate-950/40 px-2 py-1 font-medium text-slate-200">
                <span
                  className={cn(
                    "h-2 w-2 rounded-full",
                    taskStatusDotClass(task.status)
                  )}
                />
                Status: {formatLabel(task.status)}
              </span>
              <div className="flex flex-wrap gap-2">
                <span
                  className={cn(
                    "inline-flex items-center rounded-md border px-2 py-1 font-medium",
                    badgeTone(task.risk)
                  )}
                >
                  Risk: {formatLabel(task.risk)}
                </span>
                <span className="rounded-md border border-slate-800/70 bg-slate-900/60 px-2 py-1 text-slate-200">
                  Priority: {formatLabel(task.priority)}
                </span>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4 px-4 pb-4 sm:px-6 sm:pb-6">
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <Metric
              label="Readiness"
              value={formatPercent(task.readiness)}
              tone={metricTone(task.readiness)}
              progress={task.readiness}
              compactOnMobile
              highlightOnMobile
            />
            <Metric
              label="Coverage"
              value={formatPercent(task.coverage)}
              tone={metricTone(task.coverage)}
              progress={task.coverage}
              compactOnMobile
              highlightOnMobile
            />
            <Metric
              label="Deadline"
              value={formatDate(task.deadline_at)}
              tone="slate"
              compactOnMobile
            />
          </div>
          <div className="hidden flex-wrap gap-4 text-sm text-slate-400 sm:flex sm:gap-6 sm:text-[0.95rem]">
            <span>Created: {formatDate(task.created_at)}</span>
            <span>Updated: {formatDate(task.updated_at)}</span>
            <span>Archived: {formatDate(task.archived_at)}</span>
          </div>
        </CardContent>
      </Link>
    </Card>
  );
}
