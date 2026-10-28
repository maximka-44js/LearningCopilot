"use client";

import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import CreateStageForm from "@/features/projects/components/create-stage-form";
import StageCard from "@/features/projects/components/stage-card";
import { unwrapOrvalData } from "@/features/projects/lib/orval";
import { useListStagesByProject } from "@/generated/api/stages/stages";
import type { GithubComMaximka44jsLcInternalProjectsEntityStage as Stage } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityStage";

type StatusFilter = "all" | "active" | "paused" | "archived";
type RiskFilter = "all" | "low" | "medium" | "high" | "critical";

const orderValue = (stage: Stage) => {
  if (typeof stage.order === "number") return stage.order;
  return Number.MAX_SAFE_INTEGER;
};

const normalizeText = (value: string) => value.trim().toLowerCase();

export default function StagesManageContainer({ projectId }: { projectId: string }) {
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [riskFilter, setRiskFilter] = useState<RiskFilter>("all");
  const [createOpen, setCreateOpen] = useState(false);

  const { data, isLoading, isError } = useListStagesByProject(projectId);

  const rawStages = useMemo(() => {
    const payload = unwrapOrvalData<Stage[] | { status: number; data: Stage[] }>(data);
    return Array.isArray(payload) ? payload : null;
  }, [data]);

  const suggestedOrder = useMemo(() => {
    if (!rawStages || rawStages.length === 0) return 1;
    const maxOrder = rawStages.reduce((max, stage) => {
      const current = typeof stage.order === "number" ? stage.order : 0;
      return Math.max(max, current);
    }, 0);
    return Math.max(1, maxOrder + 1);
  }, [rawStages]);

  const filteredStages = useMemo(() => {
    if (!rawStages) return null;

    const query = normalizeText(search);

    return [...rawStages]
      .filter((stage) => {
        if (!stage) return false;
        if (statusFilter !== "all" && stage.status !== statusFilter) return false;
        if (riskFilter !== "all" && stage.risk !== riskFilter) return false;

        if (!query) return true;

        const haystack = normalizeText(
          `${stage.name ?? ""} ${stage.goal ?? ""}`
        );
        return haystack.includes(query);
      })
      .sort((a, b) => orderValue(a) - orderValue(b));
  }, [rawStages, search, statusFilter, riskFilter]);

  if (isLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading stages...
      </div>
    );
  }

  if (isError || !filteredStages) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Unable to load stages right now.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 rounded-xl border border-slate-800 bg-slate-950/40 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-1 flex-col gap-3 sm:flex-row sm:items-center">
          <div className="flex-1">
            <label className="sr-only" htmlFor="stage-search">
              Search stages
            </label>
            <input
              id="stage-search"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search stages (name or goal)"
              className="h-10 w-full rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 placeholder:text-slate-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
            />
          </div>

          <div className="flex flex-wrap gap-2">
            <label className="sr-only" htmlFor="stage-status">
              Filter by status
            </label>
            <select
              id="stage-status"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
              className="h-10 rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
            >
              <option value="all">All statuses</option>
              <option value="active">Active</option>
              <option value="paused">Paused</option>
              <option value="archived">Archived</option>
            </select>

            <label className="sr-only" htmlFor="stage-risk">
              Filter by risk
            </label>
            <select
              id="stage-risk"
              value={riskFilter}
              onChange={(e) => setRiskFilter(e.target.value as RiskFilter)}
              className="h-10 rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
            >
              <option value="all">All risks</option>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => setCreateOpen(true)}
          >
            New stage
          </Button>
        </div>
      </div>

      <CreateStageForm
        projectId={projectId}
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        suggestedOrder={suggestedOrder}
      />

      <div className="grid gap-6">
        {filteredStages.length === 0 ? (
          <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
            No stages match the current filters.
          </div>
        ) : (
          filteredStages.map((stage, index) => (
            <StageCard
              key={stage.id ?? `${stage.name}-${index}`}
              href={`/projects/${projectId}/stages/${stage.id ?? ""}`}
              stage={stage}
            />
          ))
        )}
      </div>
    </div>
  );
}
