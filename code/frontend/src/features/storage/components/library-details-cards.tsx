import Image from "next/image";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { GithubComMaximka44jsLcInternalStorageEntityLibrary as Library } from "@/generated/model/githubComMaximka44jsLcInternalStorageEntityLibrary";
import { formatDate } from "@/features/storage/lib/library-tree-utils";

type LibraryDetailsCardsProps = {
  library: Library;
  accessesCount: number;
  onEdit: () => void;
  onDelete: () => void;
};

export default function LibraryDetailsCards({
  library,
  accessesCount,
  onEdit,
  onDelete,
}: LibraryDetailsCardsProps) {
  return (
    <>
      <Card className="sm:hidden">
        <details className="group">
          <summary className="flex cursor-pointer list-none items-center justify-between gap-3 px-4 py-3">
            <span className="inline-flex items-center gap-2 text-xl font-semibold text-slate-100">
              <span>Данные библиотеки</span>
              <span className="text-xs text-slate-400 transition group-open:rotate-180">▼</span>
            </span>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  onEdit();
                }}
                className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-slate-700 bg-slate-900/40 transition hover:bg-slate-800/60"
                aria-label="Изменить библиотеку"
              >
                <Image src="/icons/edit-svgrepo-com.svg" alt="Изменить" width={16} height={16} />
              </button>
              <button
                type="button"
                onClick={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  onDelete();
                }}
                className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-rose-500/40 bg-rose-500/10 transition hover:bg-rose-500/20"
                aria-label="Удалить библиотеку"
              >
                <Image src="/icons/delete-2-svgrepo-com.svg" alt="Удалить" width={16} height={16} />
              </button>
            </div>
          </summary>
          <div className="space-y-3 border-t border-slate-800/80 px-4 py-4">
            <p className="break-all text-sm text-slate-400">{library.description?.trim() || "Нет описания"}</p>
            <div className="grid gap-3 text-sm text-slate-300">
              <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Видимость</p>
                <p className="mt-1 font-medium">{library.public ? "Публичная" : "Приватная"}</p>
              </div>
              <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Владелец</p>
                <p className="mt-1 break-all font-medium">
                  {library.owner_username?.trim() || "Неизвестно"}
                </p>
              </div>
              <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Создана</p>
                <p className="mt-1 font-medium">{formatDate(library.created_at)}</p>
              </div>
              <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Изменена</p>
                <p className="mt-1 font-medium">{formatDate(library.updated_at)}</p>
              </div>
              <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Участники</p>
                <p className="mt-1 font-medium">{accessesCount}</p>
              </div>
            </div>
          </div>
        </details>
      </Card>

      <Card className="hidden sm:block">
        <CardHeader className="space-y-2">
          <div className="flex items-center justify-between gap-3">
            <CardTitle className="text-xl">Данные библиотеки</CardTitle>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={onEdit}
                className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-slate-700 bg-slate-900/40 transition hover:bg-slate-800/60"
                aria-label="Изменить библиотеку"
              >
                <Image src="/icons/edit-svgrepo-com.svg" alt="Изменить" width={16} height={16} />
              </button>
              <button
                type="button"
                onClick={onDelete}
                className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-rose-500/40 bg-rose-500/10 transition hover:bg-rose-500/20"
                aria-label="Удалить библиотеку"
              >
                <Image src="/icons/delete-2-svgrepo-com.svg" alt="Удалить" width={16} height={16} />
              </button>
            </div>
          </div>
          <p className="break-all text-sm text-slate-400">{library.description?.trim() || "Нет описания"}</p>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 text-sm text-slate-300 sm:grid-cols-2 lg:grid-cols-5">
            <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
              <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Видимость</p>
              <p className="mt-1 font-medium">{library.public ? "Публичная" : "Приватная"}</p>
            </div>
            <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
              <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Владелец</p>
              <p className="mt-1 break-all font-medium">
                {library.owner_username?.trim() || "Неизвестно"}
              </p>
            </div>
            <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
              <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Создана</p>
              <p className="mt-1 font-medium">{formatDate(library.created_at)}</p>
            </div>
            <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
              <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Изменена</p>
              <p className="mt-1 font-medium">{formatDate(library.updated_at)}</p>
            </div>
            <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
              <p className="text-xs uppercase tracking-[0.15em] text-slate-500">Участники</p>
              <p className="mt-1 font-medium">{accessesCount}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </>
  );
}
