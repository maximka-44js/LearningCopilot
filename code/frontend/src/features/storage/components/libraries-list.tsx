"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import type { GithubComMaximka44jsLcInternalStorageEntityLibraryAccess as LibraryAccess } from "@/generated/model/githubComMaximka44jsLcInternalStorageEntityLibraryAccess";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type LibrariesListItem = {
  id: string;
  title: string;
  description: string;
  isPublic: boolean;
  ownerId: string;
  createdAt: string | null;
  updatedAt: string | null;
  accesses: LibraryAccess[];
};

type LibrariesListProps = {
  items: LibrariesListItem[];
  className?: string;
};

type VisibilityFilter = "all" | "public" | "private";

function ellipsis(value: string, maxLength: number) {
  const normalized = value.trim();
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return `${normalized.slice(0, Math.max(0, maxLength - 1)).trimEnd()}...`;
}

function formatDate(value: string | null) {
  if (!value) {
    return "Unknown";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }

  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "short",
    day: "2-digit",
  }).format(date);
}

function formatAccesses(accesses: LibraryAccess[]) {
  if (accesses.length === 0) {
    return "No collaborators";
  }

  const readers = accesses.filter((access) => access.role === "reader").length;
  const writers = accesses.filter((access) => access.role === "writer").length;

  return `${writers} writer${writers === 1 ? "" : "s"}, ${readers} reader${readers === 1 ? "" : "s"}`;
}

export default function LibrariesList({ items, className }: LibrariesListProps) {
  const [query, setQuery] = useState("");
  const [visibility, setVisibility] = useState<VisibilityFilter>("all");

  const visibleItems = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();

    return items
      .filter((item) => {
        const visibilityOk =
          visibility === "all" ||
          (visibility === "public" ? item.isPublic : !item.isPublic);

        const queryOk =
          normalizedQuery.length === 0 ||
          item.title.toLowerCase().includes(normalizedQuery) ||
          item.description.toLowerCase().includes(normalizedQuery);

        return visibilityOk && queryOk;
      })
      .sort((left, right) => {
        const leftTime = left.updatedAt ? Date.parse(left.updatedAt) : 0;
        const rightTime = right.updatedAt ? Date.parse(right.updatedAt) : 0;

        return rightTime - leftTime;
      });
  }, [items, query, visibility]);

  return (
    <section className={cn("space-y-6", className)}>
      <div className="flex flex-wrap items-center gap-4">
        <label className="flex min-w-[16rem] flex-1 flex-col gap-2 text-[0.8rem] uppercase tracking-[0.2em] text-slate-400">
          Search
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Find library by title or description"
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] normal-case tracking-normal text-slate-100 placeholder:text-slate-500"
          />
        </label>

        <label className="flex flex-col gap-2 text-[0.8rem] uppercase tracking-[0.2em] text-slate-400">
          Visibility
          <select
            value={visibility}
            onChange={(event) => setVisibility(event.target.value as VisibilityFilter)}
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] normal-case tracking-normal text-slate-100"
          >
            <option value="all">All</option>
            <option value="public">Public</option>
            <option value="private">Private</option>
          </select>
        </label>
      </div>

      <div className="grid gap-4">
        {visibleItems.length === 0 ? (
          <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
            No libraries found for current filters.
          </div>
        ) : (
          visibleItems.map((library) => (
            <Card key={library.id} className="transition hover:border-slate-700/80">
              <CardHeader className="space-y-3">
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <div className="space-y-2">
                    <CardTitle className="flex flex-wrap items-center gap-2 text-[1.2rem]">
                      <span
                        title={library.title}
                        className="block max-w-136 truncate"
                      >
                        {ellipsis(library.title, 56)}
                      </span>
                      <span
                        className={cn(
                          "rounded-full border px-2 py-0.5 text-[0.7rem] uppercase tracking-[0.18em]",
                          library.isPublic
                            ? "border-cyan-500/30 bg-cyan-500/10 text-cyan-200"
                            : "border-slate-700 bg-slate-900/60 text-slate-300"
                        )}
                      >
                        {library.isPublic ? "Public" : "Private"}
                      </span>
                    </CardTitle>
                    <CardDescription className="max-w-3xl text-base leading-relaxed">
                      <span
                        title={library.description}
                        className="block max-w-208 truncate"
                      >
                        {ellipsis(library.description, 120)}
                      </span>
                    </CardDescription>
                  </div>
                  <Link
                    href={`/storage/libraries/${library.id}`}
                    className="inline-flex h-9 items-center rounded-md border border-slate-700 px-3 text-sm font-medium text-slate-100 transition hover:bg-slate-900/40"
                  >
                    Open library
                  </Link>
                </div>
              </CardHeader>

              <CardContent>
                <div className="grid gap-3 text-sm text-slate-300 sm:grid-cols-2 lg:grid-cols-4">
                  <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                    <p className="text-xs uppercase tracking-[0.15em] text-slate-500">
                      Owner
                    </p>
                    <p className="mt-1 truncate font-medium" title={library.ownerId}>
                      {ellipsis(library.ownerId, 36)}
                    </p>
                  </div>
                  <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                    <p className="text-xs uppercase tracking-[0.15em] text-slate-500">
                      Collaborators
                    </p>
                    <p className="mt-1 font-medium">
                      {formatAccesses(library.accesses)}
                    </p>
                  </div>
                  <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                    <p className="text-xs uppercase tracking-[0.15em] text-slate-500">
                      Created
                    </p>
                    <p className="mt-1 font-medium">{formatDate(library.createdAt)}</p>
                  </div>
                  <div className="rounded-lg border border-slate-800/70 bg-slate-900/50 px-3 py-2">
                    <p className="text-xs uppercase tracking-[0.15em] text-slate-500">
                      Updated
                    </p>
                    <p className="mt-1 font-medium">{formatDate(library.updatedAt)}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </section>
  );
}