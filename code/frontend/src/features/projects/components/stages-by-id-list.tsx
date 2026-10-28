"use client";

import { useMemo } from "react";
import Link from "next/link";

import StageCard from "@/features/projects/components/stage-card";
import { useListStagesByProject } from "@/generated/api/stages/stages";
import type { GithubComMaximka44jsLcInternalProjectsEntityStage as Stage } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityStage";
import { cn } from "@/lib/utils";

type StagesByIdListProps = {
  projectId: string;
  stageIds: string[];
  className?: string;
};

type OrvalWrappedResponse = {
  status: number;
  data: unknown;
};

function unwrapOrvalPayload(response: unknown): unknown | null {
  if (!response) return null;
  if (Array.isArray(response)) return response;
  if (typeof response !== "object") return null;

  if ("status" in response && "data" in response) {
    const wrapped = response as OrvalWrappedResponse;
    if (wrapped.status === 200) return wrapped.data;
    return null;
  }

  return response;
}

const stageOrderValue = (stage: Stage) => {
  return typeof stage?.order === "number" ? stage.order : Number.MAX_SAFE_INTEGER;
};

export default function StagesByIdList({
  projectId,
  stageIds,
  className,
}: StagesByIdListProps) {
  const uniqueStageIds = useMemo(() => {
    return Array.from(new Set(stageIds)).filter(Boolean);
  }, [stageIds]);

  const { data, isLoading, isError } = useListStagesByProject(projectId, {
    query: {
      enabled: Boolean(projectId) && uniqueStageIds.length > 0,
    },
  });

  const stagesPayload = unwrapOrvalPayload(data);
  const stages = Array.isArray(stagesPayload) ? (stagesPayload as Stage[]) : null;

  const visibleStages = useMemo(() => {
    if (!stages) return [];
    const allowed = new Set(uniqueStageIds);
    return stages
      .filter((stage) => stage?.id && allowed.has(stage.id) && stage?.status === "active")
      .sort((a, b) => stageOrderValue(a) - stageOrderValue(b));
  }, [stages, uniqueStageIds]);

  return (
    <section className={cn("space-y-6 lg:flex lg:h-[42rem] lg:flex-col", className)}>
      <div className="flex flex-col gap-1">
        <Link
          href={`/projects/${projectId}/stages`}
          className="w-fit focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
        >
          <h2 className="inline-flex items-center gap-2 text-xl font-semibold text-slate-100 underline-offset-4 transition hover:underline">
            Stages
            <img
              src="/icons/external-link.svg"
              alt=""
              aria-hidden="true"
              className="h-4 w-4 opacity-60 brightness-0 invert"
            />
            <span className="sr-only">(opens stages)</span>
          </h2>
        </Link>
        <p className="text-sm text-slate-400">
          Stage status, goal, readiness, coverage, and risk.
        </p>
      </div>

      <div className="lc-scrollbar grid gap-6 lg:flex-1 lg:overflow-auto lg:pr-2">
        {uniqueStageIds.length === 0 ? (
          <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
            No stages returned yet.
          </div>
        ) : isLoading ? (
          <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
            Loading stages...
          </div>
        ) : isError || !stages ? (
          <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
            Unable to load stages right now.
          </div>
        ) : visibleStages.length === 0 ? (
          <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
            No matching stages found.
          </div>
        ) : (
          visibleStages.map((stage, index) => (
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
