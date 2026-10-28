"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { useToast } from "@/components/providers/toast-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getGetFolderContentsQueryKey,
  useCreateFolder,
} from "@/generated/api/folders/folders";
import { getLibraryTreeQueryKey } from "@/features/storage/lib/library-tree";

type CreateFolderModalProps = {
  open: boolean;
  libraryId: string;
  defaultParentId?: string;
  currentFolderId?: string;
  onClose: () => void;
};

type FormState = {
  name: string;
};

const STORAGE_NAME_PATTERN = /^[\p{L}\p{N} _-]+$/u;
const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

function isValidStorageName(value: string) {
  return value !== "." && value !== ".." && STORAGE_NAME_PATTERN.test(value);
}

function isUuid(value: string) {
  return UUID_PATTERN.test(value);
}

const formSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Folder name is required.")
    .max(100, "Folder name must be at most 100 characters.")
    .refine(isValidStorageName, {
      message:
        "Folder name may contain letters, numbers, spaces, '-' and '_' only. '.' and '..' are not allowed.",
    }),
});

const defaultFormState: FormState = {
  name: "",
};

export default function CreateFolderModal({
  open,
  libraryId,
  defaultParentId,
  currentFolderId,
  onClose,
}: CreateFolderModalProps) {
  const [isMounted, setIsMounted] = useState(false);
  const [form, setForm] = useState<FormState>(defaultFormState);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const { toast } = useToast();
  const queryClient = useQueryClient();

  const { mutateAsync, isPending } = useCreateFolder({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getLibraryTreeQueryKey(libraryId),
        });

        const normalizedCurrentFolderId = currentFolderId?.trim();
        if (normalizedCurrentFolderId) {
          queryClient.invalidateQueries({
            queryKey: getGetFolderContentsQueryKey(normalizedCurrentFolderId),
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

    setForm({
      name: "",
    });
    setErrorMessage(null);
  }, [defaultParentId, open]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage(null);

    const normalizedLibraryId = libraryId.trim();
    const normalizedParentId = defaultParentId?.trim();

    if (!isUuid(normalizedLibraryId)) {
      const message = "Library ID is invalid. Reload the page and try again.";
      setErrorMessage(message);
      toast({
        title: "Folder not created",
        description: message,
        variant: "error",
      });
      return;
    }

    if (normalizedParentId && !isUuid(normalizedParentId)) {
      const message = "Parent folder ID is invalid. Reload the page and try again.";
      setErrorMessage(message);
      toast({
        title: "Folder not created",
        description: message,
        variant: "error",
      });
      return;
    }

    const parsed = formSchema.safeParse(form);
    if (!parsed.success) {
      const message = parsed.error.issues[0]?.message ?? "Invalid form values.";
      setErrorMessage(message);
      toast({
        title: "Check form values",
        description: message,
        variant: "error",
      });
      return;
    }

    try {
      const response = await mutateAsync({
        data: {
          lib_id: normalizedLibraryId,
          name: parsed.data.name.trim(),
          parent_id: normalizedParentId || undefined,
        },
      });

      const status =
        typeof (response as { status?: unknown }).status === "number"
          ? (response as { status: number }).status
          : undefined;

      if (typeof status === "number" && status >= 400) {
        setErrorMessage("Folder could not be created.");
        toast({
          title: "Folder not created",
          description: "Server rejected the request.",
          variant: "error",
        });
        return;
      }

      toast({
        title: "Folder created",
        description: "Library tree has been updated.",
      });
      onClose();
    } catch {
      setErrorMessage("Folder could not be created.");
      toast({
        title: "Folder not created",
        description: "Something went wrong while creating folder.",
        variant: "error",
      });
    }
  };

  if (!open || !isMounted) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-9999 grid place-items-center bg-slate-950/70 p-4 backdrop-blur-sm">
      <Card className="w-full max-w-xl">
        <CardHeader className="space-y-2">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="space-y-1">
              <CardTitle className="text-[1.2rem]">Create folder</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Add a new folder to this library.
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
              <label className="text-sm text-slate-300">Folder name</label>
              <input
                value={form.name}
                onChange={(event) =>
                  setForm((current) => ({ ...current, name: event.target.value }))
                }
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                placeholder="Project docs"
                maxLength={100}
                required
              />
            </div>

            <p className="text-xs text-slate-500">
              Parent folder is selected automatically from current location.
            </p>

            {errorMessage ? <p className="text-sm text-rose-200">{errorMessage}</p> : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Creating..." : "Create folder"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
