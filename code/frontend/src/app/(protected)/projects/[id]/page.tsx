import { requireUserSession } from "@/features/auth/server/require-user-session";
import ProjectDashboardContainer from "@/features/projects/components/project-dashboard-container";

export default async function ProjectPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  await requireUserSession(`/projects/${id}`);

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto w-full max-w-7xl">
        <ProjectDashboardContainer projectId={id} />
      </div>
    </main>
  );
}
