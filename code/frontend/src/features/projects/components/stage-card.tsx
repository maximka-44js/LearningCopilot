import Link from "next/link";

import type { GithubComMaximka44jsLcInternalProjectsEntityStage as Stage } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityStage";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  badgeTone,
  formatDate,
  formatLabel,
  formatPercent,
  Metric,
  metricTone,
  statusDotClass,
} from "@/features/projects/components/card-metrics";
import { cn } from "@/lib/utils";

type StageCardProps = {
  href: string;
  stage: Stage;
  className?: string;
};

export default function StageCard({ href, stage, className }: StageCardProps) {
  const goal = stage.goal?.trim();

  return (
    <Card className={cn("transition hover:border-slate-700/80", className)}>
      <Link
        href={href}
        className="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
      >
        <CardHeader className="space-y-3">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="space-y-2">
              <p className="text-sm uppercase tracking-[0.3em] text-slate-500">
                Stage
                {typeof stage.order === "number" ? ` · ${stage.order}` : ""}
              </p>
              <CardTitle className="flex flex-wrap items-center gap-2 text-[1.3rem]">
                {stage.name || "Untitled stage"}
              </CardTitle>
              <CardDescription className="max-w-2xl text-base leading-relaxed">
                {goal || "No goal defined"}
              </CardDescription>
            </div>
            <div className="flex flex-col items-start gap-2 text-[0.95rem]">
              <span
                className={cn(
                  "inline-flex items-center gap-2 rounded-md border px-2 py-1 font-medium",
                  "border-slate-800/70 bg-slate-950/40 text-slate-200"
                )}
              >
                <span
                  className={cn("h-2 w-2 rounded-full", statusDotClass(stage.status))}
                />
                Status: {formatLabel(stage.status)}
              </span>
              <span
                className={cn(
                  "inline-flex items-center rounded-md border px-2 py-1 font-medium",
                  badgeTone(stage.risk)
                )}
              >
                Risk: {formatLabel(stage.risk)}
              </span>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <Metric
              label="Readiness"
              value={formatPercent(stage.readiness)}
              tone={metricTone(stage.readiness)}
              progress={stage.readiness}
            />
            <Metric
              label="Coverage"
              value={formatPercent(stage.coverage)}
              tone={metricTone(stage.coverage)}
              progress={stage.coverage}
            />
            <Metric
              label="Deadline"
              value={formatDate(stage.deadline_at)}
              tone="slate"
            />
          </div>
          <div className="flex flex-wrap gap-6 text-[0.95rem] text-slate-400">
            <span>Created: {formatDate(stage.created_at)}</span>
            <span>Updated: {formatDate(stage.updated_at)}</span>
            <span>Archived: {formatDate(stage.archived_at)}</span>
          </div>
        </CardContent>
      </Link>
    </Card>
  );
}
