"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

type ConfirmDeleteTaskModalProps = {
  open: boolean;
  taskOutcome?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
  isPending?: boolean;
};

export default function ConfirmDeleteTaskModal({
  open,
  taskOutcome,
  onCancel,
  onConfirm,
  isPending = false,
}: ConfirmDeleteTaskModalProps) {
  const [isMounted, setIsMounted] = useState(false);
  const [value, setValue] = useState("");

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!open) return;
    setValue("");
  }, [open]);

  const normalizedOutcome = useMemo(() => (taskOutcome ?? "").trim(), [taskOutcome]);
  const canDelete = normalizedOutcome.length > 0 && value.trim() === normalizedOutcome;

  if (!open || !isMounted) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-[9999] grid place-items-center bg-slate-950/70 p-4 backdrop-blur-sm">
      <Card className="w-full max-w-lg">
        <CardHeader className="space-y-2">
          <CardTitle className="text-[1.2rem]">Delete task</CardTitle>
          <CardDescription className="text-sm text-slate-400">
            This action cannot be undone. Type the task outcome to confirm.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-100">
            This will permanently delete the task and its readiness history.
          </div>
          <div className="grid gap-2">
            <label className="text-sm text-slate-300">Task outcome</label>
            <input
              value={value}
              onChange={(event) => setValue(event.target.value)}
              className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
              placeholder={normalizedOutcome || "Task outcome"}
            />
          </div>
          <div className="flex flex-wrap items-center justify-end gap-3">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button
              type="button"
              onClick={onConfirm}
              disabled={!canDelete || isPending}
              className="bg-rose-500 text-white hover:bg-rose-400"
            >
              {isPending ? "Deleting..." : "Delete task"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
