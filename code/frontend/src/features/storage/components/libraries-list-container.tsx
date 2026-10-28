"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import CreateLibraryModal from "@/features/storage/components/create-library-modal";
import LibrariesList, {
  type LibrariesListItem,
} from "@/features/storage/components/libraries-list";
import { useListLibraries } from "@/generated/api/libraries/libraries";

export default function LibrariesListContainer() {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const { data, isLoading, isError } = useListLibraries();

  const listContent = (() => {
    if (isLoading) {
      return (
        <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
          Loading libraries...
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
          Unable to load libraries right now.
        </div>
      );
    }

    const items: LibrariesListItem[] = rawLibraries.map((library) => ({
      id: library.id ?? "",
      title: library.title?.trim() || "Untitled library",
      description: library.description?.trim() || "No description",
      isPublic: Boolean(library.public),
      ownerId: library.user_id ?? "Unknown owner",
      createdAt: library.created_at ?? null,
      updatedAt: library.updated_at ?? null,
      accesses: Array.isArray(library.accesses) ? library.accesses : [],
    }));

    return <LibrariesList items={items} />;
  })();

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-lg font-medium text-slate-100">Your libraries</h2>
        <Button type="button" onClick={() => setIsCreateOpen(true)}>
          New library
        </Button>
      </div>

      {listContent}

      <CreateLibraryModal
        open={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
      />
    </section>
  );
}