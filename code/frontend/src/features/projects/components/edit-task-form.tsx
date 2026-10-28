"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getGetTaskReadinessViewQueryKey,
} from "@/generated/api/projects-read-model/projects-read-model";
import { getGetTaskByIdQueryKey, useUpdateTask } from "@/generated/api/tasks/tasks";
import type { GithubComMaximka44jsLcInternalProjectsEntityTask as Task } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityTask";

const defaultForm = {
  outcome: "",
  difficulty: "3",
  deadline: "",
  priority: "normal",
};

type EditTaskFormState = typeof defaultForm;

type FormErrors = {
  message?: string;
};

type EditTaskFormProps = {
  projectId: string;
  stageId: string;
  taskId: string;
  task?: Task | null;
  open: boolean;
  onClose: () => void;
};

const editTaskSchema = z.object({
  outcome: z.string().trim().min(1, "Outcome is required.").max(2000),
  difficulty: z
    .string()
    .trim()
    .refine((value) => Number.isFinite(Number(value)), "Difficulty is required.")
    .transform((value) => Number(value))
    .refine((value) => Number.isInteger(value) && value >= 1 && value <= 10, "Difficulty must be 1-10."),
  deadline: z
    .string()
    .optional()
    .refine(
      (value) => !value || !Number.isNaN(Date.parse(value)),
      "Deadline must be a valid date."
    )
    .or(z.literal("")),
  priority: z.enum(["low", "normal", "high"]),
});

const toInputDateTime = (value?: string) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString().slice(0, 16);
};

export default function EditTaskForm({
  projectId,
  stageId,
  taskId,
  task,
  open,
  onClose,
}: EditTaskFormProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<EditTaskFormState>(defaultForm);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isMounted, setIsMounted] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const initialForm = useMemo(() => {
    if (!task) return defaultForm;
    return {
      outcome: task.outcome ?? "",
      difficulty: typeof task.difficulty === "number" ? String(task.difficulty) : "3",
      deadline: toInputDateTime(task.deadline_at),
      priority: task.priority ?? "normal",
    };
  }, [task]);

  useEffect(() => {
    if (!open) return;
    setForm(initialForm);
    setErrors({});
  }, [open, initialForm]);

  const isDirty = useMemo(() => {
    return (
      form.outcome !== initialForm.outcome ||
      form.difficulty !== initialForm.difficulty ||
      form.deadline !== initialForm.deadline ||
      form.priority !== initialForm.priority
    );
  }, [form, initialForm]);

  const { mutateAsync, isPending } = useUpdateTask({
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

  const handleChange = (field: keyof EditTaskFormState, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrors({});

    const parsed = editTaskSchema.safeParse(form);
    if (!parsed.success) {
      const firstIssue = parsed.error.issues[0];
      const message = firstIssue?.message ?? "Invalid form values.";
      setErrors({ message });
      toast({
        title: "Check the form",
        description: message,
        variant: "error",
      });
      return;
    }

    try {
      const deadline_at = parsed.data.deadline
        ? new Date(parsed.data.deadline).toISOString()
        : undefined;

      const response = await mutateAsync({
        projectId,
        stageId,
        taskId,
        data: {
          outcome: parsed.data.outcome.trim(),
          difficulty: parsed.data.difficulty,
          deadline_at,
          priority: parsed.data.priority,
        },
      });

      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 200) {
          setErrors({ message: "Task could not be updated." });
          toast({
            title: "Task not updated",
            description: "Server rejected the request.",
            variant: "error",
          });
          return;
        }
      }

      toast({
        title: "Task updated",
        description: "The task details were saved.",
      });
      onClose();
    } catch {
      setErrors({ message: "Task could not be updated." });
      toast({
        title: "Task not updated",
        description: "Something went wrong while saving the task.",
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
              <CardTitle className="text-[1.2rem]">Edit task</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Update the outcome, priority, difficulty, or deadline.
              </CardDescription>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={onClose}>
              Close
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Outcome</label>
              <textarea
                value={form.outcome}
                onChange={(event) => handleChange("outcome", event.target.value)}
                className="min-h-[96px] rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100"
                placeholder="Explain cell membrane transport principles"
                required
              />
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Difficulty (1-10)</label>
                <input
                  inputMode="numeric"
                  value={form.difficulty}
                  onChange={(event) => handleChange("difficulty", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="3"
                  required
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Priority</label>
                <select
                  value={form.priority}
                  onChange={(event) => handleChange("priority", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                >
                  <option value="high">High</option>
                  <option value="normal">Normal</option>
                  <option value="low">Low</option>
                </select>
              </div>
            </div>

            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Deadline</label>
              <input
                type="datetime-local"
                value={form.deadline}
                onFocus={(event) => event.currentTarget.showPicker?.()}
                onChange={(event) => handleChange("deadline", event.target.value)}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                style={{ colorScheme: "dark" }}
              />
            </div>

            {errors.message ? (
              <p className="text-sm text-rose-200">{errors.message}</p>
            ) : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={isPending || !isDirty}>
                {isPending ? "Saving..." : "Save changes"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
