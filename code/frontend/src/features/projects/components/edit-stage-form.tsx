"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getGetStageBoardQueryKey,
} from "@/generated/api/projects-read-model/projects-read-model";
import {
  getListStagesByProjectQueryKey,
  useUpdateStage,
} from "@/generated/api/stages/stages";
import type { GithubComMaximka44jsLcInternalProjectsEntityStage as Stage } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityStage";

const defaultForm = {
  name: "",
  goal: "",
  deadline: "",
  order: "1",
};

type EditStageFormState = typeof defaultForm;

type FormErrors = {
  message?: string;
};

type EditStageFormProps = {
  projectId: string;
  stageId: string;
  stage?: Stage | null;
  open: boolean;
  onClose: () => void;
};

const editStageSchema = z.object({
  name: z.string().trim().min(1, "Name is required.").max(160),
  goal: z.string().trim().max(2000).optional().or(z.literal("")),
  deadline: z
    .string()
    .optional()
    .refine(
      (value) => !value || !Number.isNaN(Date.parse(value)),
      "Deadline must be a valid date."
    )
    .or(z.literal("")),
  order: z
    .string()
    .trim()
    .refine((value) => value.length > 0, "Order is required.")
    .refine((value) => Number.isFinite(Number(value)), "Order must be a number.")
    .transform((value) => Number(value))
    .refine((value) => Number.isInteger(value) && value >= 1, "Order must be >= 1."),
});

const toInputDateTime = (value?: string) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString().slice(0, 16);
};

export default function EditStageForm({
  projectId,
  stageId,
  stage,
  open,
  onClose,
}: EditStageFormProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<EditStageFormState>(defaultForm);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isMounted, setIsMounted] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const initialForm = useMemo(() => {
    if (!stage) return defaultForm;
    return {
      name: stage.name ?? "",
      goal: stage.goal ?? "",
      deadline: toInputDateTime(stage.deadline_at),
      order: typeof stage.order === "number" ? String(stage.order) : "1",
    };
  }, [stage]);

  useEffect(() => {
    if (!open) return;
    setForm(initialForm);
    setErrors({});
  }, [open, initialForm]);

  const isDirty = useMemo(() => {
    return (
      form.name !== initialForm.name ||
      form.goal !== initialForm.goal ||
      form.deadline !== initialForm.deadline ||
      form.order !== initialForm.order
    );
  }, [form, initialForm]);

  const { mutateAsync, isPending } = useUpdateStage({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetStageBoardQueryKey(projectId, stageId),
        });
        queryClient.invalidateQueries({
          queryKey: getListStagesByProjectQueryKey(projectId),
        });
      },
    },
  });

  const handleChange = (field: keyof EditStageFormState, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrors({});

    const parsed = editStageSchema.safeParse(form);
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
          name: parsed.data.name.trim(),
          goal: parsed.data.goal?.trim() || undefined,
          deadline_at,
          order: parsed.data.order,
        },
      });

      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 200) {
          setErrors({ message: "Stage could not be updated." });
          toast({
            title: "Stage not updated",
            description: "Server rejected the request.",
            variant: "error",
          });
          return;
        }
      }

      toast({
        title: "Stage updated",
        description: "The stage details were saved.",
      });
      onClose();
    } catch {
      setErrors({ message: "Stage could not be updated." });
      toast({
        title: "Stage not updated",
        description: "Something went wrong while saving the stage.",
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
              <CardTitle className="text-[1.2rem]">Edit stage</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Update name, order, and optional goal or deadline.
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
              <label className="text-sm text-slate-300">Name</label>
              <input
                value={form.name}
                onChange={(event) => handleChange("name", event.target.value)}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="Cell biology"
                required
              />
            </div>

            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Goal</label>
              <textarea
                value={form.goal}
                onChange={(event) => handleChange("goal", event.target.value)}
                className="min-h-[96px] rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100"
                placeholder="Understand structure and function of cells."
              />
            </div>

            <div className="grid gap-4 md:grid-cols-2">
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
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Order</label>
                <input
                  inputMode="numeric"
                  value={form.order}
                  onChange={(event) => handleChange("order", event.target.value)}
                  className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                  placeholder="1"
                  required
                />
              </div>
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
