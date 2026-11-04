"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import {
  getListProjectsQueryKey,
  useCreateProject,
} from "@/generated/api/projects/projects";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { getApiErrorMessage } from "@/lib/api-error-message";
import { ENTITY_TITLE_MAX_LENGTH } from "@/lib/validation-limits";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const defaultForm = {
  name: "",
  description: "",
  deadline: "",
  priority: "normal",
  metadataTrack: "",
};

type CreateProjectFormState = typeof defaultForm;

type FormErrors = {
  message?: string;
};

type CreateProjectFormProps = {
  open: boolean;
  onClose: () => void;
};

const createProjectSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Введите название проекта.")
    .max(ENTITY_TITLE_MAX_LENGTH, `Название не должно превышать ${ENTITY_TITLE_MAX_LENGTH} символов.`),
  description: z.string().trim().max(4000).optional().or(z.literal("")),
  deadline: z
    .string()
    .optional()
    .refine(
      (value) => !value || !Number.isNaN(Date.parse(value)),
      "Введите корректную дату дедлайна."
    )
    .or(z.literal("")),
  priority: z.enum(["low", "normal", "high"]),
  metadataTrack: z.string().trim().max(200).optional().or(z.literal("")),
});

export default function CreateProjectForm({ open, onClose }: CreateProjectFormProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<CreateProjectFormState>(defaultForm);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isMounted, setIsMounted] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const { mutateAsync, isPending } = useCreateProject({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getListProjectsQueryKey(),
        });
      },
    },
  });

  const handleChange = (
    field: keyof CreateProjectFormState,
    value: string
  ) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrors({});

    const parsed = createProjectSchema.safeParse(form);
    if (!parsed.success) {
      const firstIssue = parsed.error.issues[0];
      const message = firstIssue?.message ?? "Проверьте значения в форме.";
      setErrors({ message });
      toast({
        title: "Проверьте форму",
        description: message,
        variant: "error",
      });
      return;
    }

    try {
      const deadline_at = parsed.data.deadline
        ? new Date(parsed.data.deadline).toISOString()
        : undefined;
      const metadataTrack = parsed.data.metadataTrack?.trim();
      const metadata = metadataTrack
        ? {
            source: "manual",
            track: metadataTrack,
          }
        : undefined;

      const response = await mutateAsync({
        data: {
          name: parsed.data.name.trim(),
          description: parsed.data.description?.trim() || undefined,
          deadline_at,
          priority: parsed.data.priority,
          metadata,
        },
      });

      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 201) {
          setErrors({ message: "Не удалось создать проект." });
          toast({
            title: "Проект не создан",
            description: "Сервер отклонил запрос.",
            variant: "error",
          });
          return;
        }
      }

      setErrors({});
      setForm(defaultForm);
      toast({
        title: "Проект создан",
        description: "Проект появился в списке.",
      });
      onClose();
    } catch (error) {
      const message = getApiErrorMessage(error);
      setErrors({ message });
      toast({
        title: "Проект не создан",
        description: message,
        variant: "error",
      });
    }
  };

  if (!open || !isMounted) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-[9999] grid place-items-center bg-slate-950/70 p-4 backdrop-blur-sm">
      <Card className="w-full max-w-2xl">
        <CardHeader className="space-y-2">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="space-y-1">
              <CardTitle className="text-[1.2rem]">Создать проект</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Укажите приоритет, дедлайн и краткое описание.
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
                value={form.name ?? ""}
                onChange={(event) => handleChange("name", event.target.value)}
                maxLength={ENTITY_TITLE_MAX_LENGTH}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="Подготовка к экзамену по биологии"
                required
              />
            </div>
            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Описание</label>
              <textarea
                value={form.description ?? ""}
                onChange={(event) =>
                  handleChange("description", event.target.value)
                }
                className="min-h-[96px] rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100"
                placeholder="План подготовки по неделям с контрольными точками"
              />
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Дедлайн</label>
                <input
                  type="datetime-local"
                  value={form.deadline ?? ""}
                  onFocus={(event) => event.currentTarget.showPicker?.()}
                  onChange={(event) => handleChange("deadline", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  style={{ colorScheme: "dark" }}
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Приоритет</label>
                <select
                  value={form.priority ?? "normal"}
                  onChange={(event) => handleChange("priority", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                >
                  <option value="high">Высокий</option>
                  <option value="normal">Обычный</option>
                  <option value="low">Низкий</option>
                </select>
              </div>
            </div>
            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Направление</label>
              <input
                value={form.metadataTrack ?? ""}
                onChange={(event) =>
                  handleChange("metadataTrack", event.target.value)
                }
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="биология"
              />
            </div>
            {errors.message ? (
              <p className="text-sm text-rose-200">{errors.message}</p>
            ) : null}
            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Отмена
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Создание..." : "Создать проект"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
