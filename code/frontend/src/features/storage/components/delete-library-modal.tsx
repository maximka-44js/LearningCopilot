"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { z } from "zod";

import { useToast } from "@/components/providers/toast-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getGetLibraryByIdQueryKey,
  getListLibrariesQueryKey,
  useDeleteLibrary,
} from "@/generated/api/libraries/libraries";
import type { LibrariesListItem } from "@/features/storage/components/libraries-list";

type DeleteLibraryModalProps = {
  open: boolean;
  library: LibrariesListItem | null;
  onClose: () => void;
};

function normalizeLibraryId(value: string | undefined) {
  return value?.trim() ?? "";
}

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

function buildDeleteLibrarySchema(expectedTitle: string) {
  return z.object({
    libraryId: z
      .string()
      .trim()
      .regex(UUID_PATTERN, "Library ID is invalid. Reload and try again."),
    confirmText: z
      .string()
      .trim()
      .refine((value) => value === expectedTitle, {
        message: "Confirmation text does not match library title.",
      }),
  });
}

export default function DeleteLibraryModal({
  open,
  library,
  onClose,
}: DeleteLibraryModalProps) {
  const [isMounted, setIsMounted] = useState(false);
  const [confirmText, setConfirmText] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const { toast } = useToast();
  const queryClient = useQueryClient();
  const router = useRouter();

  const libraryId = normalizeLibraryId(library?.id);
  const expectedText = library?.title.trim() ?? "";
  const isConfirmTextValid = confirmText.trim() === expectedText;

  const { mutateAsync, isPending } = useDeleteLibrary({
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
    if (!open) {
      return;
    }
    setConfirmText("");
    setErrorMessage(null);
  }, [open, library?.id]);

  const handleDelete = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage(null);

    const parsed = buildDeleteLibrarySchema(expectedText).safeParse({
      libraryId,
      confirmText,
    });

    if (!parsed.success) {
      const message = parsed.error.issues[0]?.message ?? "Invalid delete form values.";
      setErrorMessage(message);
      toast({
        title: "Library not deleted",
        description: message,
        variant: "error",
      });
      return;
    }

    if (!library) {
      const message = "Library data is missing. Reload and try again.";
      setErrorMessage(message);
      toast({
        title: "Library not deleted",
        description: message,
        variant: "error",
      });
      return;
    }

    try {
      const response = await mutateAsync({ id: parsed.data.libraryId });
      const status =
        typeof (response as { status?: unknown }).status === "number"
          ? (response as { status: number }).status
          : undefined;

      if (typeof status === "number" && status >= 400) {
        const message = "Server rejected library deletion.";
        setErrorMessage(message);
        toast({
          title: "Library not deleted",
          description: message,
          variant: "error",
        });
        return;
      }

      toast({
        title: "Library deleted",
        description: "Library has been removed.",
      });
      onClose();
      router.push("/storage/libraries");
    } catch {
      const message = "Something went wrong while deleting library.";
      setErrorMessage(message);
      toast({
        title: "Library not deleted",
        description: message,
        variant: "error",
      });
    }
  };

  if (!open || !isMounted || !library) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-9999 grid place-items-center bg-slate-950/75 p-4 backdrop-blur-sm">
      <Card className="w-full max-w-xl border-rose-500/30">
        <CardHeader className="space-y-2">
          <CardTitle className="text-[1.2rem] text-rose-200">Delete library</CardTitle>
          <CardDescription className="text-sm text-slate-300">
            This action cannot be undone. This will permanently delete the library and
            its storage data.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <form onSubmit={handleDelete} className="grid gap-4">
            <div className="rounded-md border border-slate-800 bg-slate-900/50 p-3 text-sm text-slate-300">
              <p>
                To confirm, type library title:
              </p>
              <p className="mt-2 rounded border border-slate-700 bg-slate-950/70 px-2 py-1 font-mono text-slate-100">
                {expectedText}
              </p>
            </div>

            <div className="grid gap-2">
              <label className="text-sm text-slate-300">Library title</label>
              <input
                value={confirmText}
                onChange={(event) => setConfirmText(event.target.value)}
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="Type exact library title"
                required
              />
            </div>

            {errorMessage ? <p className="text-sm text-rose-200">{errorMessage}</p> : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={isPending || !isConfirmTextValid}
                className="bg-rose-500 text-white hover:bg-rose-400"
              >
                {isPending ? "Deleting..." : "I understand, delete this library"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
