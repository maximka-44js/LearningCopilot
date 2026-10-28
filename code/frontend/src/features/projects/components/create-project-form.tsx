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
  name: z.string().trim().min(1, "Name is required.").max(200),
  description: z.string().trim().max(4000).optional().or(z.literal("")),
  deadline: z
    .string()
    .optional()
    .refine(
      (value) => !value || !Number.isNaN(Date.parse(value)),
      "Deadline must be a valid date."
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
          setErrors({ message: "Project could not be created." });
          toast({
            title: "Project not created",
            description: "Server rejected the request.",
            variant: "error",
          });
          return;
        }
      }

      setErrors({});
      setForm(defaultForm);
      toast({
        title: "Project created",
        description: "Your project is ready in the list.",
      });
      onClose();
    } catch (error) {
      setErrors({ message: "Project could not be created." });
      toast({
        title: "Project not created",
        description: "Something went wrong while saving the project.",
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
              <CardTitle className="text-[1.2rem]">Create project</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Priority, deadline, and a short description.
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
                value={form.name ?? ""}
                onChange={(event) => handleChange("name", event.target.value)}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="Prepare for microbiology exam"
                required
              />
            </div>
            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Description</label>
              <textarea
                value={form.description ?? ""}
                onChange={(event) =>
                  handleChange("description", event.target.value)
                }
                className="min-h-[96px] rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100"
                placeholder="Weekly study plan with milestones"
              />
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="grid gap-2">
                <label className="text-sm text-slate-300">Deadline</label>
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
                <label className="text-sm text-slate-300">Priority</label>
                <select
                  value={form.priority ?? "normal"}
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
              <label className="text-sm text-slate-300">Metadata track</label>
              <input
                value={form.metadataTrack ?? ""}
                onChange={(event) =>
                  handleChange("metadataTrack", event.target.value)
                }
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="biology"
              />
            </div>
            {errors.message ? (
              <p className="text-sm text-rose-200">{errors.message}</p>
            ) : null}
            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Creating..." : "Create project"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
