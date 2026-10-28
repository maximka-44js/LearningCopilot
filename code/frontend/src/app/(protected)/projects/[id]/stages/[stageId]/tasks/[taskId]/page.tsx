import { requireUserSession } from "@/features/auth/server/require-user-session";
import TaskReadinessContainer from "@/features/projects/components/task-readiness-container";

export default async function TaskReadinessPage({
  params,
}: {
  params: Promise<{ id: string; stageId: string; taskId: string }>;
}) {
  const { id, stageId, taskId } = await params;

  await requireUserSession(`/projects/${id}/stages/${stageId}/tasks/${taskId}`);

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto w-full max-w-7xl">
        <TaskReadinessContainer
          projectId={id}
          stageId={stageId}
          taskId={taskId}
        />
      </div>
    </main>
  );
}
