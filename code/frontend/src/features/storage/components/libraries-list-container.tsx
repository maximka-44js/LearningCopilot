"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import CreateLibraryModal from "@/features/storage/components/create-library-modal";
import LibrariesList, {
  type LibrariesListItem,
} from "@/features/storage/components/libraries-list";
import PublicLibrariesCatalog from "@/features/storage/components/public-libraries-catalog";
import { useListLibraries } from "@/generated/api/libraries/libraries";

export default function LibrariesListContainer() {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [activeView, setActiveView] = useState<"mine" | "global">("mine");
  const { data, isLoading, isError } = useListLibraries();

  const listContent = (() => {
    if (isLoading) {
      return (
        <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
          Загрузка библиотек...
        </div>
      );
    }

    const rawLibraries = Array.isArray(data)
      ? data
      : data?.status === 200
        ? data.data
        : null;

    if (isError || !rawLibraries) {
      return (
        <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
          Не удалось загрузить библиотеки. Попробуйте ещё раз позже.
        </div>
      );
    }

    const items: LibrariesListItem[] = rawLibraries.map((library) => ({
      id: library.id ?? "",
      title: library.title?.trim() || "Без названия",
      description: library.description?.trim() || "Нет описания",
      isPublic: Boolean(library.public),
      ownerId:
        library.owner_username?.trim() ||
        library.user_id ||
        "Неизвестный владелец",
      createdAt: library.created_at ?? null,
      updatedAt: library.updated_at ?? null,
      accesses: Array.isArray(library.accesses) ? library.accesses : [],
    }));

    return <LibrariesList items={items} />;
  })();

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Tabs
          value={activeView}
          onValueChange={(value) => setActiveView(value as "mine" | "global")}
        >
          <TabsList className="border border-slate-800 bg-slate-900/60">
            <TabsTrigger value="mine" className="px-4 py-2">
              Мои библиотеки
            </TabsTrigger>
            <TabsTrigger value="global" className="px-4 py-2">
              Глобальные
            </TabsTrigger>
          </TabsList>
        </Tabs>

        {activeView === "mine" ? (
          <Button type="button" onClick={() => setIsCreateOpen(true)}>
            Новая библиотека
          </Button>
        ) : null}
      </div>

      {activeView === "mine" ? listContent : <PublicLibrariesCatalog />}

      <CreateLibraryModal
        open={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
      />
    </section>
  );
}