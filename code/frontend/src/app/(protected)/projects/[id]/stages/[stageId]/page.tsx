import { requireUserSession } from "@/features/auth/server/require-user-session";
import StageBoardContainer from "@/features/projects/components/stage-board-container";

export default async function StageBoardPage({
  params,
}: {
  params: Promise<{ id: string; stageId: string }>;
}) {
  const { id, stageId } = await params;

  await requireUserSession(`/projects/${id}/stages/${stageId}`);

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto w-full max-w-7xl">
        <StageBoardContainer projectId={id} stageId={stageId} />
      </div>
    </main>
  );
}
