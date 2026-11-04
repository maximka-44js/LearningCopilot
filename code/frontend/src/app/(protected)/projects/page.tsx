import { requireUserSession } from "@/features/auth/server/require-user-session";
import ProjectsListContainer from "@/features/projects/components/projects-list-container";

export default async function ProjectsPage() {
  await requireUserSession("/projects");

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10">
        <header className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
            Проекты
          </p>
          <h1 className="text-3xl font-semibold">Проекты</h1>
          <p className="max-w-2xl text-sm text-slate-400">
            Отслеживайте готовность, охват и риски по текущим планам.
          </p>
        </header>

        <ProjectsListContainer />
      </div>
    </main>
  );
}
