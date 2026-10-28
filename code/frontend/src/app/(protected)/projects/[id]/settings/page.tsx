import { requireUserSession } from "@/features/auth/server/require-user-session";
import ProjectSettingsContainer from "@/features/projects/components/project-settings-container";

export default async function ProjectSettingsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  await requireUserSession(`/projects/${id}/settings`);

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10">
        <ProjectSettingsContainer projectId={id} />
      </div>
    </main>
  );
}
