import { requireUserSession } from "@/features/auth/server/require-user-session";
import LibraryTreeContainer from "@/features/storage/components/library-tree-container";

type LibraryPageProps = {
  params: Promise<{ id: string }>;
};

export default async function LibraryPage({ params }: LibraryPageProps) {
  const { id } = await params;

  await requireUserSession(`/storage/libraries/${id}`);

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-10">
        <header className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
            Хранилище
          </p>
          <p className="max-w-2xl text-sm text-slate-400">
            Просматривайте папки и файлы этой библиотеки.
          </p>
        </header>

        <LibraryTreeContainer libraryId={id} />
      </div>
    </main>
  );
}
