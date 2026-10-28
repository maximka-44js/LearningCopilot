"use client";

import StagesList from "@/features/projects/components/stages-list";
import { useListStagesByProject } from "@/generated/api/stages/stages";

export default function StagesListContainer({ projectId }: { projectId: string }) {
  const { data, isLoading, isError } = useListStagesByProject(projectId);

  if (isLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading stages...
      </div>
    );
  }

  const rawStages = Array.isArray(data)
    ? data
    : data?.status === 200
      ? data.data
      : null;

  if (isError || !rawStages) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Unable to load stages right now.
      </div>
    );
  }

  return <StagesList projectId={projectId} stages={rawStages} />;
}
