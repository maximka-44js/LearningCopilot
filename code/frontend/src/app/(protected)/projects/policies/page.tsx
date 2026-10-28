import { requireUserSession } from "@/features/auth/server/require-user-session";
import PoliciesSettingsPanel from "@/features/projects/components/policies-settings-panel";

export default async function ProjectPoliciesPage() {
  await requireUserSession("/projects/policies");

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-8">
        <header className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
            Projects
          </p>
          <h1 className="text-3xl font-semibold">Policy settings</h1>
          <p className="max-w-2xl text-sm text-slate-400">
            Global defaults for aggregation, calibration, recommendation, and risk.
          </p>
        </header>

        <PoliciesSettingsPanel />
      </div>
    </main>
  );
}
