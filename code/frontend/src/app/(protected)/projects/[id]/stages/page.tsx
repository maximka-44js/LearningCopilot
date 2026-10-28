import { requireUserSession } from "@/features/auth/server/require-user-session";
import StagesManageContainer from "@/features/projects/components/stages-manage-container";

export default async function ProjectStagesPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  await requireUserSession(`/projects/${id}/stages`);

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10">
        <header className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
            Project
          </p>
          <h1 className="text-3xl font-semibold">Stages</h1>
          <p className="max-w-2xl text-sm text-slate-400">
            Review and manage the stages in this project.
          </p>
        </header>

        <StagesManageContainer projectId={id} />
      </div>
    </main>
  );
}
