"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { getApiErrorMessage } from "@/lib/api-error-message";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getGetTaskReadinessViewQueryKey } from "@/generated/api/projects-read-model/projects-read-model";
import { getGetTaskByIdQueryKey, useRecordAssessmentSnapshot } from "@/generated/api/tasks/tasks";
import type { GithubComMaximka44jsLcInternalProjectsEntityTask as Task } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityTask";

const defaultForm = {
  block_type: "comprehension",
  normalized_score: "",
};

type RecordAssessmentFormState = typeof defaultForm;

type FormErrors = {
  message?: string;
};

type RecordAssessmentFormProps = {
  projectId: string;
  stageId: string;
  taskId: string;
  task?: Task | null;
  open: boolean;
  onClose: () => void;
};

const toOptionalNumber = (value: string) => {
  if (!value.trim()) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const recordAssessmentSchema = z.object({
  block_type: z.enum(["comprehension", "memorization", "practice"]),
  normalized_score: z
    .string()
    .trim()
    .refine((value) => Number.isFinite(Number(value)), "Оценка должна быть числом.")
    .transform((value) => toOptionalNumber(value))
    .refine((value) => value !== undefined && value >= 0 && value <= 100, "Оценка должна быть от 0 до 100."),
});

export default function RecordAssessmentForm({
  projectId,
  stageId,
  taskId,
  task,
  open,
  onClose,
}: RecordAssessmentFormProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<RecordAssessmentFormState>(defaultForm);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isMounted, setIsMounted] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const initialForm = useMemo(() => {
    const firstBlock = task?.blocks?.[0]?.type ?? "comprehension";
    return {
      ...defaultForm,
      block_type: firstBlock,
    };
  }, [task]);

  useEffect(() => {
    if (!open) return;
    setForm(initialForm);
    setErrors({});
  }, [open, initialForm]);

  const { mutateAsync, isPending } = useRecordAssessmentSnapshot({
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

  const handleChange = (field: keyof RecordAssessmentFormState, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrors({});

    const parsed = recordAssessmentSchema.safeParse(form);
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
      const recorded_at = new Date().toISOString();

      const response = await mutateAsync({
        projectId,
        stageId,
        taskId,
        data: {
          block_type: parsed.data.block_type,
          normalized_score: parsed.data.normalized_score,
          recorded_at,
          source: "manual_confirmation",
        },
      });

      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 200) {
          setErrors({ message: "Не удалось сохранить оценку." });
          toast({
            title: "Оценка не сохранена",
            description: "Сервер отклонил запрос.",
            variant: "error",
          });
          return;
        }
      }

      toast({
        title: "Оценка сохранена",
        description: "Снимок оценки сохранён.",
      });
      onClose();
    } catch (error) {
      const message = getApiErrorMessage(error);
      setErrors({ message });
      toast({
        title: "Оценка не сохранена",
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
              <CardTitle className="text-[1.2rem]">Оценить задачу</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Сохраните новую оценку выбранного контекста задачи.
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
              <label className="text-sm text-slate-300">Ограниченный контекст</label>
              <select
                value={form.block_type}
                onChange={(event) => handleChange("block_type", event.target.value)}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
              >
                <option value="comprehension">Понимание</option>
                <option value="memorization">Запоминание</option>
                <option value="practice">Практика</option>
              </select>
            </div>

            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Самооценка</label>
              <input
                inputMode="decimal"
                value={form.normalized_score}
                onChange={(event) => handleChange("normalized_score", event.target.value)}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="71"
              />
              <p className="text-xs text-slate-500">Оценка от 0 до 100 для расчёта освоения.</p>
            </div>

            {errors.message ? (
              <p className="text-sm text-rose-200">{errors.message}</p>
            ) : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Отмена
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Сохранение..." : "Сохранить оценку"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
