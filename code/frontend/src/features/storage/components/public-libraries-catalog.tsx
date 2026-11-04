"use client";

import { useEffect, useRef, useState } from "react";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { Star } from "lucide-react";

import { useToast } from "@/components/providers/toast-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getListLibrariesQueryKey,
  listPublicLibraries,
  useDeleteStorageLibrariesIdStar,
  useAddPublicLibrary,
  usePostStorageLibrariesIdStar,
} from "@/generated/api/libraries/libraries";
import type { GithubComMaximka44jsLcInternalStorageEntityPublicLibrariesPage } from "@/generated/model";
import { getApiErrorMessage } from "@/lib/api-error-message";

const PAGE_SIZE = 20;

type PublicLibrariesPage = GithubComMaximka44jsLcInternalStorageEntityPublicLibrariesPage;

function formatDate(value?: string) {
  if (!value) {
    return "Неизвестно";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Неизвестно";
  }
  return new Intl.DateTimeFormat("ru-RU", {
    year: "numeric",
    month: "short",
    day: "2-digit",
  }).format(date);
}

function getPublicLibrariesPage(
  response: Awaited<ReturnType<typeof listPublicLibraries>>
): PublicLibrariesPage | null {
  if ("libraries" in response) {
    return response as PublicLibrariesPage;
  }
  return response.status === 200 &&
    "libraries" in response.data
    ? response.data
    : null;
}

export default function PublicLibrariesCatalog() {
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [sort, setSort] = useState<"newest" | "popular">("newest");
  const loadMoreRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const {
    data,
    isLoading,
    isError,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["/storage/libraries/public", debouncedQuery, sort],
    initialPageParam: 0,
    queryFn: ({ pageParam, signal }) =>
      listPublicLibraries(
        {
          search: debouncedQuery || undefined,
          sort,
          limit: PAGE_SIZE,
          offset: pageParam,
        },
        { signal }
      ),
    getNextPageParam: (lastPage, allPages) => {
      const page = getPublicLibrariesPage(lastPage);
      return page?.has_more ? allPages.length * PAGE_SIZE : undefined;
    },
  });
  const { mutateAsync, isPending } = useAddPublicLibrary();
  const { mutateAsync: starLibrary, isPending: isStarring } = usePostStorageLibrariesIdStar();
  const { mutateAsync: unstarLibrary, isPending: isUnstarring } =
    useDeleteStorageLibrariesIdStar();

  useEffect(() => {
    const timeout = window.setTimeout(() => setDebouncedQuery(query.trim()), 300);
    return () => window.clearTimeout(timeout);
  }, [query]);

  useEffect(() => {
    const sentinel = loadMoreRef.current;
    if (!sentinel || !hasNextPage) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && !isFetchingNextPage) {
          void fetchNextPage();
        }
      },
      { rootMargin: "240px" }
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [fetchNextPage, hasNextPage, isFetchingNextPage]);

  const libraries =
    data?.pages.flatMap((page) => getPublicLibrariesPage(page)?.libraries ?? []) ?? [];

  const addLibrary = async (id: string, title: string) => {
    try {
      await mutateAsync({ id });
      await queryClient.invalidateQueries({ queryKey: getListLibrariesQueryKey() });
      toast({
        title: "Библиотека добавлена",
        description: `«${title}» появилась в ваших библиотеках.`,
      });
    } catch (error) {
      toast({
        title: "Библиотека не добавлена",
        description: getApiErrorMessage(error),
        variant: "error",
      });
    }
  };

  const toggleStar = async (id: string, starred: boolean) => {
    try {
      if (starred) {
        await unstarLibrary({ id });
      } else {
        await starLibrary({ id });
      }
      await queryClient.invalidateQueries({ queryKey: ["/storage/libraries/public"] });
    } catch (error) {
      toast({
        title: "Не удалось изменить оценку",
        description: getApiErrorMessage(error),
        variant: "error",
      });
    }
  };

  return (
    <section className="space-y-5">
      <div className="space-y-2">
        <h2 className="text-lg font-medium text-slate-100">Глобальные публичные библиотеки</h2>
        <p className="text-sm text-slate-400">
          Найдите открытую библиотеку и добавьте её в свой список. Чтение доступно всем,
          а права на редактирование выдаёт владелец.
        </p>
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Поиск по названию, описанию или владельцу"
          className="w-full rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500"
        />
        <select
          value={sort}
          onChange={(event) => setSort(event.target.value as "newest" | "popular")}
          className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-300"
          aria-label="Сортировка библиотек"
        >
          <option value="newest">Сначала новые</option>
          <option value="popular">По популярности</option>
        </select>
      </div>

      {isLoading ? (
        <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
          Загрузка каталога...
        </div>
      ) : isError && libraries.length === 0 ? (
        <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-rose-200">
          Не удалось загрузить публичные библиотеки.
        </div>
      ) : libraries.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
          Публичные библиотеки не найдены.
        </div>
      ) : (
        <div className="grid gap-4">
          {libraries.map((library) => {
            const id = library.id ?? "";
            const title = library.title?.trim() || "Без названия";
            return (
              <Card key={id}>
                <CardHeader className="space-y-2">
                  <CardTitle>{title}</CardTitle>
                  <CardDescription>{library.description?.trim() || "Нет описания"}</CardDescription>
                </CardHeader>
                <CardContent className="flex items-center justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <div className="space-y-1">
                      <p className="text-xs text-slate-500">
                        Владелец: {library.owner_username || "Неизвестно"}
                      </p>
                      <p className="text-xs text-slate-500">
                        Создана: {formatDate(library.created_at)} · Изменена:{" "}
                        {formatDate(library.updated_at)}
                      </p>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      disabled={!id || isStarring || isUnstarring}
                      onClick={() => toggleStar(id, Boolean(library.starred))}
                      aria-label={library.starred ? "Убрать звезду" : "Поставить звезду"}
                      title={library.starred ? "Убрать звезду" : "Поставить звезду"}
                      className="gap-1 text-slate-400 hover:text-amber-300"
                    >
                      <Star
                        className={library.starred ? "fill-amber-300 text-amber-300" : ""}
                        size={16}
                      />
                      {library.star_count ?? 0}
                    </Button>
                  </div>
                  <Button type="button" size="sm" disabled={!id || isPending} onClick={() => addLibrary(id, title)}>
                    Добавить к себе
                  </Button>
                </CardContent>
              </Card>
            );
          })}
          <div ref={loadMoreRef} aria-hidden="true" className="h-1" />
          {isFetchingNextPage ? (
            <div className="py-4 text-center text-sm text-slate-400">
              Загрузка библиотек...
            </div>
          ) : isError ? (
            <div className="py-4 text-center text-sm text-rose-200">
              Не удалось загрузить следующую страницу библиотек.
            </div>
          ) : null}
        </div>
      )}
    </section>
  );
}
