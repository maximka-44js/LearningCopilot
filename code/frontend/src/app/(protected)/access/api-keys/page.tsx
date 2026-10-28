import { requireUserSession } from "@/features/auth/server/require-user-session";
import ApiKeysManager from "@/features/api-keys/components/api-keys-manager";

export default async function ApiKeysPage() {
  await requireUserSession("/access/api-keys");

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10">
        <header className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">Access BC</p>
          <h1 className="text-3xl font-semibold">API Keys</h1>
          <p className="max-w-2xl text-sm text-slate-400">
            Manage integration keys for platform APIs with precise domain and access scopes.
          </p>
        </header>

        <ApiKeysManager />
      </div>
    </main>
  );
}