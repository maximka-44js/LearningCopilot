"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getGetStageBoardQueryKey } from "@/generated/api/projects-read-model/projects-read-model";
import { useAddTask } from "@/generated/api/tasks/tasks";

const defaultForm = {
  outcome: "",
  difficulty: "3",
  deadline: "",
  priority: "normal",
  comprehensionParts: "1",
  memorizationParts: "1",
  practiceParts: "1",
};

type CreateTaskFormState = typeof defaultForm;

type FormErrors = {
  message?: string;
};

type CreateTaskFormProps = {
  projectId: string;
  stageId: string;
  open: boolean;
  onClose: () => void;
};

const createTaskSchema = z
  .object({
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
    comprehensionParts: z
      .string()
      .trim()
      .refine((value) => Number.isFinite(Number(value)), "Comprehension parts are required.")
      .transform((value) => Number(value))
      .refine((value) => Number.isInteger(value) && value >= 0, "Parts must be 0 or more."),
    memorizationParts: z
      .string()
      .trim()
      .refine((value) => Number.isFinite(Number(value)), "Memorization parts are required.")
      .transform((value) => Number(value))
      .refine((value) => Number.isInteger(value) && value >= 0, "Parts must be 0 or more."),
    practiceParts: z
      .string()
      .trim()
      .refine((value) => Number.isFinite(Number(value)), "Practice parts are required.")
      .transform((value) => Number(value))
      .refine((value) => Number.isInteger(value) && value >= 0, "Parts must be 0 or more."),
  })
  .refine(
    (values) =>
      values.comprehensionParts + values.memorizationParts + values.practiceParts > 0,
    {
      message: "At least one block must have a part.",
      path: ["comprehensionParts"],
    }
  );

export default function CreateTaskForm({
  projectId,
  stageId,
  open,
  onClose,
}: CreateTaskFormProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<CreateTaskFormState>(defaultForm);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isMounted, setIsMounted] = useState(false);
  const { toast } = useToast();

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

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!open) return;
    setForm(defaultForm);
    setErrors({});
  }, [open]);

  const { mutateAsync, isPending } = useAddTask({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetStageBoardQueryKey(projectId, stageId),
        });
      },
    },
  });

  const handleChange = (field: keyof CreateTaskFormState, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrors({});

    const parsed = createTaskSchema.safeParse(form);
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
        data: {
          outcome: parsed.data.outcome.trim(),
          difficulty: parsed.data.difficulty,
          deadline_at,
          priority: parsed.data.priority,
          blocks: [
            {
              type: "comprehension",
              weight: computedWeights.comp,
            },
            {
              type: "memorization",
              weight: computedWeights.memo,
            },
            {
              type: "practice",
              weight: computedWeights.prac,
            },
          ],
        },
      });

      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 201) {
          setErrors({ message: "Task could not be created." });
          toast({
            title: "Task not created",
            description: "Server rejected the request.",
            variant: "error",
          });
          return;
        }
      }

      toast({
        title: "Task created",
        description: "The task is now available in the list.",
      });
      onClose();
    } catch {
      setErrors({ message: "Task could not be created." });
      toast({
        title: "Task not created",
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
              <CardTitle className="text-[1.2rem]">Create task</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Define outcome, priority, and readiness weights.
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

            <div className="grid gap-4 md:grid-cols-3">
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Comprehension parts</label>
                <input
                  inputMode="numeric"
                  value={form.comprehensionParts}
                  onChange={(event) => handleChange("comprehensionParts", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="1"
                />
                <span className="text-xs text-slate-500">
                  {computedWeights.comp}%
                </span>
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Memorization parts</label>
                <input
                  inputMode="numeric"
                  value={form.memorizationParts}
                  onChange={(event) => handleChange("memorizationParts", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="1"
                />
                <span className="text-xs text-slate-500">
                  {computedWeights.memo}%
                </span>
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Practice parts</label>
                <input
                  inputMode="numeric"
                  value={form.practiceParts}
                  onChange={(event) => handleChange("practiceParts", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="1"
                />
                <span className="text-xs text-slate-500">
                  {computedWeights.prac}%
                </span>
              </div>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3 text-sm">
              <div className="flex flex-wrap items-center gap-3">
                <span className="text-slate-400">Total parts: {totalParts}</span>
                {totalParts > 0 ? (
                  <span className="text-emerald-300">Ready</span>
                ) : (
                  <span className="text-amber-200">Add parts to continue</span>
                )}
              </div>
            </div>

            {errors.message ? (
              <p className="text-sm text-rose-200">{errors.message}</p>
            ) : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Creating..." : "Create task"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
