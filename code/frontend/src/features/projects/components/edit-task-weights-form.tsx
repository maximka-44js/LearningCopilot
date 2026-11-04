"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getGetTaskReadinessViewQueryKey } from "@/generated/api/projects-read-model/projects-read-model";
import { getGetTaskByIdQueryKey, useConfigureTaskWeights } from "@/generated/api/tasks/tasks";
import type { GithubComMaximka44jsLcInternalProjectsEntityTask as Task } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityTask";

const defaultForm = {
  comprehensionParts: "1",
  memorizationParts: "1",
  practiceParts: "1",
};

type EditTaskWeightsState = typeof defaultForm;

type FormErrors = {
  message?: string;
};

type EditTaskWeightsFormProps = {
  projectId: string;
  stageId: string;
  taskId: string;
  task?: Task | null;
  open: boolean;
  onClose: () => void;
};

const weightsSchema = z
  .object({
    comprehensionParts: z
      .string()
      .trim()
      .refine((value) => Number.isFinite(Number(value)), "Введите части понимания.")
      .transform((value) => Number(value))
      .refine((value) => Number.isInteger(value) && value >= 0, "Количество частей не может быть отрицательным."),
    memorizationParts: z
      .string()
      .trim()
      .refine((value) => Number.isFinite(Number(value)), "Введите части запоминания.")
      .transform((value) => Number(value))
      .refine((value) => Number.isInteger(value) && value >= 0, "Количество частей не может быть отрицательным."),
    practiceParts: z
      .string()
      .trim()
      .refine((value) => Number.isFinite(Number(value)), "Введите части практики.")
      .transform((value) => Number(value))
      .refine((value) => Number.isInteger(value) && value >= 0, "Количество частей не может быть отрицательным."),
  })
  .refine(
    (values) =>
      values.comprehensionParts + values.memorizationParts + values.practiceParts > 0,
    {
      message: "Добавьте хотя бы одну часть в любой контекст.",
      path: ["comprehensionParts"],
    }
  );

const getWeightFromTask = (task: Task | null | undefined, type: string) => {
  if (!task || !Array.isArray(task.blocks)) return undefined;
  const block = task.blocks.find((item) => item?.type === type);
  return typeof block?.weight === "number" ? block.weight : undefined;
};

export default function EditTaskWeightsForm({
  projectId,
  stageId,
  taskId,
  task,
  open,
  onClose,
}: EditTaskWeightsFormProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<EditTaskWeightsState>(defaultForm);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isMounted, setIsMounted] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const initialForm = useMemo(() => {
    const comp = getWeightFromTask(task, "comprehension");
    const memo = getWeightFromTask(task, "memorization");
    const prac = getWeightFromTask(task, "practice");

    return {
      comprehensionParts: comp !== undefined ? String(Math.max(0, Math.round(comp))) : "1",
      memorizationParts: memo !== undefined ? String(Math.max(0, Math.round(memo))) : "1",
      practiceParts: prac !== undefined ? String(Math.max(0, Math.round(prac))) : "1",
    };
  }, [task]);

  useEffect(() => {
    if (!open) return;
    setForm(initialForm);
    setErrors({});
  }, [open, initialForm]);

  const isDirty = useMemo(() => {
    return (
      form.comprehensionParts !== initialForm.comprehensionParts ||
      form.memorizationParts !== initialForm.memorizationParts ||
      form.practiceParts !== initialForm.practiceParts
    );
  }, [form, initialForm]);

  const parts = useMemo(() => {
    const comp = Number(form.comprehensionParts || 0);
    const memo = Number(form.memorizationParts || 0);
    const prac = Number(form.practiceParts || 0);
    return {
      comp: Number.isFinite(comp) ? comp : 0,
      memo: Number.isFinite(memo) ? memo : 0,
      prac: Number.isFinite(prac) ? prac : 0,
    };
  }, [form.comprehensionParts, form.memorizationParts, form.practiceParts]);

  const totalParts = useMemo(() => parts.comp + parts.memo + parts.prac, [parts]);

  const computedWeights = useMemo(() => {
    if (totalParts <= 0) {
      return { comp: 0, memo: 0, prac: 0 };
    }

    const raw = [
      { key: "comp", value: (parts.comp / totalParts) * 100 },
      { key: "memo", value: (parts.memo / totalParts) * 100 },
      { key: "prac", value: (parts.prac / totalParts) * 100 },
    ];

    const floors = raw.map((item) => ({
      key: item.key,
      floor: Math.floor(item.value),
      frac: item.value - Math.floor(item.value),
    }));

    let remainder = 100 - floors.reduce((sum, item) => sum + item.floor, 0);
    const sorted = [...floors].sort((a, b) => b.frac - a.frac);
    for (let i = 0; i < sorted.length && remainder > 0; i += 1) {
      sorted[i].floor += 1;
      remainder -= 1;
    }

    const result = { comp: 0, memo: 0, prac: 0 };
    sorted.forEach((item) => {
      if (item.key === "comp") result.comp = item.floor;
      if (item.key === "memo") result.memo = item.floor;
      if (item.key === "prac") result.prac = item.floor;
    });

    return result;
  }, [parts, totalParts]);

  const { mutateAsync, isPending } = useConfigureTaskWeights({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetTaskReadinessViewQueryKey(projectId, stageId, taskId),
        });
        queryClient.invalidateQueries({
          queryKey: getGetTaskByIdQueryKey(projectId, stageId, taskId),
        });
      },
    },
  });

  const handleChange = (field: keyof EditTaskWeightsState, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrors({});

    const parsed = weightsSchema.safeParse(form);
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
      const response = await mutateAsync({
        projectId,
        stageId,
        taskId,
        data: {
          weights: {
            comprehension: computedWeights.comp,
            memorization: computedWeights.memo,
            practice: computedWeights.prac,
          },
        },
      });

      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 200) {
          setErrors({ message: "Не удалось обновить веса." });
          toast({
            title: "Веса не обновлены",
            description: "Сервер отклонил запрос.",
            variant: "error",
          });
          return;
        }
      }

      toast({
        title: "Веса обновлены",
        description: "Веса задачи сохранены.",
      });
      onClose();
    } catch {
      setErrors({ message: "Не удалось обновить веса." });
      toast({
        title: "Веса не обновлены",
        description: "При сохранении весов произошла ошибка.",
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
              <CardTitle className="text-[1.2rem]">Изменить веса контекстов</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Настройте долю каждого учебного контекста.
              </CardDescription>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={onClose}>
              Закрыть
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Части понимания</label>
                <input
                  inputMode="numeric"
                  value={form.comprehensionParts}
                  onChange={(event) => handleChange("comprehensionParts", event.target.value)}
                  className="h-10 w-full rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="40"
                />
                <span className="text-xs text-slate-500">
                  {computedWeights.comp}%
                </span>
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Части запоминания</label>
                <input
                  inputMode="numeric"
                  value={form.memorizationParts}
                  onChange={(event) => handleChange("memorizationParts", event.target.value)}
                  className="h-10 w-full rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="30"
                />
                <span className="text-xs text-slate-500">
                  {computedWeights.memo}%
                </span>
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Части практики</label>
                <input
                  inputMode="numeric"
                  value={form.practiceParts}
                  onChange={(event) => handleChange("practiceParts", event.target.value)}
                  className="h-10 w-full rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="30"
                />
                <span className="text-xs text-slate-500">
                  {computedWeights.prac}%
                </span>
              </div>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3 text-sm">
              <div className="flex flex-wrap items-center gap-3">
                <span className="text-slate-400">Всего частей: {totalParts}</span>
                {totalParts > 0 ? (
                  <span className="text-emerald-300">Готово</span>
                ) : (
                  <span className="text-amber-200">Добавьте части, чтобы продолжить</span>
                )}
              </div>
              <span className="text-slate-500">Общий вес: 100%</span>
            </div>

            {errors.message ? (
              <p className="text-sm text-rose-200">{errors.message}</p>
            ) : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Отмена
              </Button>
              <Button type="submit" disabled={isPending || !isDirty}>
                {isPending ? "Сохранение..." : "Сохранить веса"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
