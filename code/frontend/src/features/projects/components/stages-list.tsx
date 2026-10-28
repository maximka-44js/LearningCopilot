"use client";

import { useMemo } from "react";

import type { GithubComMaximka44jsLcInternalProjectsEntityStage as Stage } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityStage";
import StageCard from "@/features/projects/components/stage-card";
import { cn } from "@/lib/utils";

type StagesListProps = {
  projectId: string;
  stages: Stage[];
  className?: string;
};

const orderValue = (stage: Stage) => {
  if (typeof stage.order === "number") return stage.order;
  return Number.MAX_SAFE_INTEGER;
};

export default function StagesList({ projectId, stages, className }: StagesListProps) {
  const orderedStages = useMemo(() => {
    return [...stages].sort((a, b) => orderValue(a) - orderValue(b));
  }, [stages]);

  return (
    <section className={cn("space-y-6", className)}>
      <div className="flex flex-col gap-1">
        <h2 className="text-xl font-semibold text-slate-100">Stages</h2>
        <p className="text-sm text-slate-400">
          Ordered by the stage sequence in your project.
        </p>
      </div>
      <div className="grid gap-6">
        {orderedStages.length === 0 ? (
          <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
            No stages yet.
          </div>
        ) : (
          orderedStages.map((stage, index) => (
            <StageCard
              key={stage.id ?? `${stage.name}-${index}`}
              href={`/projects/${projectId}/stages/${stage.id ?? ""}`}
              stage={stage}
            />
          ))
        )}
      </div>
    </section>
  );
}
