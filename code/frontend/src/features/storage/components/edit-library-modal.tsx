"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { useToast } from "@/components/providers/toast-provider";
import { ENTITY_TITLE_MAX_LENGTH } from "@/lib/validation-limits";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getGetLibraryByIdQueryKey,
  getListLibrariesQueryKey,
  useUpdateLibrary,
} from "@/generated/api/libraries/libraries";
import type { LibrariesListItem } from "@/features/storage/components/libraries-list";

type EditLibraryModalProps = {
  open: boolean;
  library: LibrariesListItem | null;
  onClose: () => void;
};

type AccessFormRow = {
  id: string;
  role: "reader" | "writer";
  userID: string;
};

type FormState = {
  title: string;
  description: string;
  isPublic: boolean;
  accesses: AccessFormRow[];
};

const STORAGE_NAME_PATTERN = /^[\p{L}\p{N} _-]+$/u;

function isValidStorageName(value: string) {
  return value !== "." && value !== ".." && STORAGE_NAME_PATTERN.test(value);
}

const createAccessRow = (userID = "", role: "reader" | "writer" = "reader"): AccessFormRow => ({
  id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`,
  role,
  userID,
});

const formSchema = z.object({
  title: z
    .string()
    .trim()
    .min(1, "Введите название.")
    .max(ENTITY_TITLE_MAX_LENGTH, `Название должно содержать не более ${ENTITY_TITLE_MAX_LENGTH} символов.`)
    .refine(isValidStorageName, {
      message:
        "Название может содержать только буквы, цифры, пробелы, «-» и «_». «.» и «..» недопустимы.",
    }),
  description: z.string().trim().max(1000).optional().or(z.literal("")),
  isPublic: z.boolean(),
  accesses: z.array(
    z.object({
      role: z.enum(["reader", "writer"]),
      userID: z.string().trim(),
    })
  ),
});

const defaultFormState: FormState = {
  title: "",
  description: "",
  isPublic: false,
  accesses: [createAccessRow()],
};

function normalizeLibraryId(value: string | undefined) {
  return value?.trim() ?? "";
}

export default function EditLibraryModal({
  open,
  library,
  onClose,
}: EditLibraryModalProps) {
  const [isMounted, setIsMounted] = useState(false);
  const [form, setForm] = useState<FormState>(defaultFormState);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const { toast } = useToast();
  const queryClient = useQueryClient();

  const libraryId = normalizeLibraryId(library?.id);

  const { mutateAsync, isPending } = useUpdateLibrary({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getListLibrariesQueryKey(),
        });

        if (libraryId) {
          queryClient.invalidateQueries({
            queryKey: getGetLibraryByIdQueryKey(libraryId),
          });
        }
      },
    },
  });

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!open || !library) {
      return;
    }

    const mappedAccesses =
      library.accesses.length > 0
        ? library.accesses.map((access) =>
            createAccessRow(
              access.userID?.trim() ?? "",
              access.role === "writer" ? "writer" : "reader"
            )
          )
        : [createAccessRow()];

    setForm({
      title: library.title,
      description: library.description === "Нет описания" ? "" : library.description,
      isPublic: library.isPublic,
      accesses: mappedAccesses,
    });
    setErrorMessage(null);
  }, [library, open]);

  const updateAccess = (id: string, key: "role" | "userID", value: string) => {
    setForm((current) => ({
      ...current,
      accesses: current.accesses.map((access) =>
        access.id === id
          ? {
              ...access,
              [key]: key === "role" ? (value as "reader" | "writer") : value,
            }
          : access
      ),
    }));
  };

  const addAccess = () => {
    setForm((current) => ({
      ...current,
      accesses: [...current.accesses, createAccessRow()],
    }));
  };

  const removeAccess = (id: string) => {
    setForm((current) => {
      const next = current.accesses.filter((access) => access.id !== id);
      return {
        ...current,
        accesses: next.length > 0 ? next : [createAccessRow()],
      };
    });
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage(null);

    if (!library || !libraryId) {
      const message = "Library ID is empty. Reload and try again.";
      setErrorMessage(message);
      toast({
        title: "Библиотека не обновлена",
        description: message,
        variant: "error",
      });
      return;
    }

    const parsed = formSchema.safeParse({
      title: form.title,
      description: form.description,
      isPublic: form.isPublic,
      accesses: form.accesses.map((access) => ({
        role: access.role,
        userID: access.userID,
      })),
    });

    if (!parsed.success) {
      const message = parsed.error.issues[0]?.message ?? "Проверьте значения в форме.";
      setErrorMessage(message);
      toast({
        title: "Проверьте значения",
        description: message,
        variant: "error",
      });
      return;
    }

    try {
      const response = await mutateAsync({
        id: libraryId,
        data: {
          title: parsed.data.title.trim(),
          description: parsed.data.description?.trim() || undefined,
          public: parsed.data.isPublic,
          accesses: parsed.data.accesses
            .filter((access) => access.userID.trim().length > 0)
            .map((access) => ({
              ...access,
              role: parsed.data.isPublic ? "writer" as const : access.role,
            })),
        },
      });

      const status =
        typeof (response as { status?: unknown }).status === "number"
          ? (response as { status: number }).status
          : undefined;

      if (typeof status === "number" && status >= 400) {
        const message = "Сервер отклонил обновление библиотеки.";
        setErrorMessage(message);
        toast({
          title: "Библиотека не обновлена",
          description: message,
          variant: "error",
        });
        return;
      }

      toast({
        title: "Библиотека обновлена",
        description: "Данные библиотеки сохранены.",
      });
      onClose();
    } catch {
      const message = "При обновлении библиотеки произошла ошибка.";
      setErrorMessage(message);
      toast({
        title: "Библиотека не обновлена",
        description: message,
        variant: "error",
      });
    }
  };

  if (!open || !isMounted || !library) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-9999 grid place-items-center bg-slate-950/70 p-4 backdrop-blur-sm">
      <Card className="w-full max-w-2xl">
        <CardHeader className="space-y-2">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="space-y-1">
              <CardTitle className="text-[1.2rem]">Изменить библиотеку</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Измените название, видимость и доступ.
              </CardDescription>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={onClose}>
              Закрыть
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Название</label>
              <input
                value={form.title}
                onChange={(event) => setForm((current) => ({ ...current, title: event.target.value }))}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="Моя библиотека"
                maxLength={ENTITY_TITLE_MAX_LENGTH}
                required
              />
            </div>

            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Описание</label>
              <textarea
                value={form.description}
                onChange={(event) =>
                  setForm((current) => ({ ...current, description: event.target.value }))
                }
                className="min-h-24 rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100"
                placeholder="Описание библиотеки"
                maxLength={1000}
              />
            </div>

            <label className="flex items-center gap-3 text-sm text-slate-200">
              <input
                type="checkbox"
                checked={form.isPublic}
                onChange={(event) =>
                  setForm((current) => ({ ...current, isPublic: event.target.checked }))
                }
                className="h-4 w-4 rounded border-slate-700 bg-slate-900"
              />
              Публичная библиотека
            </label>

            {form.isPublic ? (
              <p className="text-sm text-slate-400">
                Читать публичную библиотеку могут все пользователи. Добавляйте
                сюда только пользователей с правом редактирования.
              </p>
            ) : null}

            <div className="space-y-3 rounded-lg border border-slate-800/80 bg-slate-900/30 p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-medium text-slate-200">
                  {form.isPublic ? "Доступ на редактирование" : "Доступ"}
                </p>
                <Button type="button" variant="outline" size="sm" onClick={addAccess}>
                  Добавить доступ
                </Button>
              </div>
              <div className="grid gap-3">
                {form.accesses.map((access) => (
                  <div key={access.id} className="grid gap-2 rounded-md border border-slate-800/80 p-3 md:grid-cols-[1fr_10rem_auto]">
                    <input
                      value={access.userID}
                      onChange={(event) => updateAccess(access.id, "userID", event.target.value)}
                      className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                      placeholder="Идентификатор пользователя"
                    />
                    {form.isPublic ? (
                      <div className="flex h-10 items-center px-2 text-sm text-slate-400">
                        Редактирование
                      </div>
                    ) : (
                      <select
                        value={access.role}
                        onChange={(event) => updateAccess(access.id, "role", event.target.value)}
                        className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                      >
                        <option value="reader">Чтение</option>
                        <option value="writer">Запись</option>
                      </select>
                    )}
                    <Button type="button" variant="outline" className="border-rose-400/40 text-rose-200 hover:bg-rose-500/20" onClick={() => removeAccess(access.id)}>
                      Удалить
                    </Button>
                  </div>
                ))}
              </div>
            </div>

            {errorMessage ? <p className="text-sm text-rose-200">{errorMessage}</p> : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Отмена
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Сохранение..." : "Сохранить изменения"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
