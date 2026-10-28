import { requireUserSession } from "@/features/auth/server/require-user-session";
import LibrariesListContainer from "@/features/storage/components/libraries-list-container";

export default async function StorageLibrariesPage() {
  await requireUserSession("/storage/libraries");

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10">
        <header className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
            Storage
          </p>
          <h1 className="text-3xl font-semibold">Libraries</h1>
          <p className="max-w-2xl text-sm text-slate-400">
            Browse your document libraries, access level, and visibility in one
            place.
          </p>
        </header>

        <LibrariesListContainer />
      </div>
    </main>
  );
}