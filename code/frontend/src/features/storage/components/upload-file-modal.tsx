"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { z } from "zod";

import { useToast } from "@/components/providers/toast-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { getLibraryTreeQueryKey } from "@/features/storage/lib/library-tree";
import {
    useConfirmUpload,
    useInitUpload,
    usePresignUploadPart,
} from "@/generated/api/files/files";
import { getGetFolderContentsQueryKey } from "@/generated/api/folders/folders";
import { orvalAxios } from "@/lib/orval/mutator";

type UploadFileModalProps = {
  open: boolean;
  libraryId: string;
  folderId?: string;
  currentFolderId?: string;
  onClose: () => void;
  onUploadProgressChange?: (state: UploadProgressState) => void;
  onUploadCancelHandlerChange?: (handler: (() => Promise<void>) | null) => void;
};

type UploadUiStatus = "idle" | "uploading" | "error" | "success";

export type UploadProgressState = {
  visible: boolean;
  status: UploadUiStatus;
  progressPct: number | null;
  progressLabel: string;
  errorMessage: string | null;
  fileName: string | null;
  canCancel: boolean;
};

class UploadCanceledError extends Error {
  constructor() {
    super("Upload canceled by user.");
  }
}

const MAX_FILE_SIZE_BYTES = 500 * 1024 * 1024;
const CHUNK_SIZE_BYTES = 4 * 1024 * 1024;
const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

const ALLOWED_MIME_TYPES = new Set([
  "application/pdf",
  "text/plain",
  "text/markdown",
  "text/csv",
  "application/json",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  "application/msword",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "application/vnd.ms-excel",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
  "application/vnd.ms-powerpoint",
  "text/html",
]);

const ALLOWED_EXTENSIONS = new Set([
  "pdf",
  "txt",
  "md",
  "markdown",
  "csv",
  "json",
  "doc",
  "docx",
  "xls",
  "xlsx",
  "ppt",
  "pptx",
  "html",
]);

function getFileExtension(fileName: string) {
  const parts = fileName.split(".");
  if (parts.length < 2) {
    return "";
  }
  return parts.at(-1)?.toLowerCase() ?? "";
}

function formatUploadErrorByStatus(status?: number) {
  if (status === 400) {
    return "Invalid upload data. Check selected file and options.";
  }

  if (status === 401) {
    return "Session expired. Please sign in again.";
  }

  if (status === 403) {
    return "You do not have permission to upload into this location.";
  }

  if (typeof status === "number" && status >= 500) {
    return "Server error during upload. Please retry.";
  }

  return "Upload failed. Please retry.";
}

function normalizeInitResponse(
  value: unknown
): { fileId: string; uploadId?: string; presignUrl?: string } | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const payload =
    "data" in value && value.data && typeof value.data === "object"
      ? (value.data as Record<string, unknown>)
      : (value as Record<string, unknown>);

  const file = payload.file ?? payload.File;
  const fileIdFromNestedFile =
    file &&
    typeof file === "object" &&
    (("id" in file && typeof (file as { id?: unknown }).id === "string") ||
      ("ID" in file && typeof (file as { ID?: unknown }).ID === "string") ||
      ("file_id" in file && typeof (file as { file_id?: unknown }).file_id === "string") ||
      ("fileID" in file && typeof (file as { fileID?: unknown }).fileID === "string") ||
      ("FileID" in file && typeof (file as { FileID?: unknown }).FileID === "string"))
      ? String(
          (file as {
            id?: unknown;
            ID?: unknown;
            file_id?: unknown;
            fileID?: unknown;
            FileID?: unknown;
          }).id ??
            (file as { ID?: unknown }).ID ??
            (file as { file_id?: unknown }).file_id ??
            (file as { fileID?: unknown }).fileID ??
            (file as { FileID?: unknown }).FileID
        ).trim()
      : "";

  const fileIdFromTopLevel =
    typeof payload.file_id === "string"
      ? payload.file_id.trim()
      : typeof payload.fileID === "string"
        ? payload.fileID.trim()
        : typeof payload.FileID === "string"
          ? payload.FileID.trim()
        : typeof payload.id === "string"
          ? payload.id.trim()
          : typeof payload.ID === "string"
            ? payload.ID.trim()
          : "";

  const fileId = fileIdFromNestedFile || fileIdFromTopLevel;

  if (!fileId) {
    return null;
  }

  const uploadIdRaw =
    payload.uploadID ?? payload.upload_id ?? payload.UploadID ?? payload.uploadId;
  const presignUrlRaw =
    payload.presignURL ??
    payload.presign_url ??
    payload.PresignURL ??
    payload.presignUrl;

  return {
    fileId,
    uploadId: typeof uploadIdRaw === "string" ? uploadIdRaw.trim() || undefined : undefined,
    presignUrl:
      typeof presignUrlRaw === "string" ? presignUrlRaw.trim() || undefined : undefined,
  };
}

function debugDescribeKeys(value: unknown) {
  if (!value || typeof value !== "object") {
    return "<non-object>";
  }

  const rootKeys = Object.keys(value as Record<string, unknown>);
  const data = (value as { data?: unknown }).data;
  const dataKeys =
    data && typeof data === "object"
      ? Object.keys(data as Record<string, unknown>)
      : [];

  return `rootKeys=[${rootKeys.join(", ")}], dataKeys=[${dataKeys.join(", ")}]`;
}

function normalizePresignPartUrl(value: unknown): string | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const payload =
    "data" in value && value.data && typeof value.data === "object"
      ? (value.data as Record<string, unknown>)
      : (value as Record<string, unknown>);

  const directKeys = ["url", "presignURL", "presign_url"];
  for (const key of directKeys) {
    const candidate = payload[key];
    if (typeof candidate === "string" && candidate.trim().length > 0) {
      return candidate.trim();
    }
  }

  const firstStringValue = Object.values(payload).find(
    (entry) => typeof entry === "string" && entry.trim().length > 0
  ) as string | undefined;

  return firstStringValue?.trim() ?? null;
}

function bytesToBase64(bytes: Uint8Array) {
  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }

  return btoa(binary);
}

async function sha256Base64(blob: Blob): Promise<string | undefined> {
  if (!crypto.subtle) {
    return undefined;
  }

  const buffer = await blob.arrayBuffer();
  const digest = await crypto.subtle.digest("SHA-256", buffer);
  return bytesToBase64(new Uint8Array(digest));
}

function resolvePresignPutUrl(rawUrl: string) {
  if (/^https?:\/\//i.test(rawUrl)) {
    return rawUrl;
  }

  if (typeof window === "undefined") {
    return rawUrl;
  }

  const configuredBase = process.env.NEXT_PUBLIC_API_BASE_URL;
  if (configuredBase) {
    try {
      const parsed = new URL(configuredBase, window.location.origin);
      return new URL(rawUrl, parsed.origin).toString();
    } catch {
      return new URL(rawUrl, window.location.origin).toString();
    }
  }

  return new URL(rawUrl, window.location.origin).toString();
}

async function uploadBinaryPart(
  url: string,
  blob: Blob,
  checksumSha256: string | undefined,
  signal?: AbortSignal
) {
  const headers: HeadersInit = {
    "Content-Type": blob.type || "application/octet-stream",
  };

  if (checksumSha256) {
    headers["x-amz-checksum-algorithm"] = "SHA256";
    headers["x-amz-checksum-sha256"] = checksumSha256;
  }

  const response = await fetch(resolvePresignPutUrl(url), {
    method: "PUT",
    body: blob,
    signal,
    headers,
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new Error(
        "Part upload failed with 401 Unauthorized. Local object endpoint requires auth for presigned PUT; check access token/session validity."
      );
    }
    throw new Error(`Part upload failed with status ${response.status}`);
  }

  const etag =
    response.headers.get("etag")?.trim() ||
    response.headers.get("ETag")?.trim() ||
    response.headers.get("x-amz-etag")?.trim() ||
    response.headers.get("X-Amz-Etag")?.trim();

  if (!etag) {
    throw new Error(
      "Upload part response does not include ETag. Configure S3 CORS to expose ETag."
    );
  }

  return etag;
}

async function abortUpload(fileId: string, uploadId: string) {
  await orvalAxios<unknown>(`/storage/files/${fileId}/abort`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      upload_id: uploadId,
    }),
  });
}

function formatUploadErrorByClientFailure(error: unknown, fileSizeBytes?: number) {
  if (!error || typeof error !== "object") {
    return null;
  }

  const errorCode = "code" in error ? String((error as { code?: unknown }).code ?? "") : "";
  const errorMessage =
    "message" in error ? String((error as { message?: unknown }).message ?? "") : "";

  const isNetworkLikeFailure =
    errorCode === "ERR_NETWORK" || /network\s*error|cors|failed/i.test(errorMessage);

  if (!isNetworkLikeFailure) {
    return null;
  }

  if (typeof fileSizeBytes === "number" && fileSizeBytes >= 100 * 1024 * 1024) {
    return "Large file upload was blocked before app received a response. Most likely backend/proxy body-size limit or missing CORS headers on error response (often 413/502/504).";
  }

  return "Network/CORS failure during upload. Check backend CORS and proxy limits.";
}

function isUuid(value: string) {
  return UUID_PATTERN.test(value);
}

const uploadFormSchema = z
  .object({
    libraryId: z.string().trim().min(1, "Library ID is empty. Reload the page and try again."),
    folderId: z.string().trim().optional(),
    file: z.instanceof(File, { message: "File is required." }),
  })
  .superRefine((value, ctx) => {
    if (!isUuid(value.libraryId)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["libraryId"],
        message: "Library ID is invalid. Reload the page and try again.",
      });
    }

    if (value.folderId && !isUuid(value.folderId)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["folderId"],
        message: "Folder ID is invalid. Reload the page and try again.",
      });
    }

    if (value.file.size <= 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["file"],
        message: "Selected file is empty.",
      });
      return;
    }

    if (value.file.size > MAX_FILE_SIZE_BYTES) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["file"],
        message: "File is too large. Maximum size is 500 MB.",
      });
    }

    const extension = getFileExtension(value.file.name);
    const isMimeAllowed = value.file.type ? ALLOWED_MIME_TYPES.has(value.file.type) : false;
    const isExtensionAllowed = extension ? ALLOWED_EXTENSIONS.has(extension) : false;

    if (!isMimeAllowed && !isExtensionAllowed) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["file"],
        message: "Unsupported file type.",
      });
    }

  });

export default function UploadFileModal({
  open,
  libraryId,
  folderId,
  currentFolderId,
  onClose,
  onUploadProgressChange,
  onUploadCancelHandlerChange,
}: UploadFileModalProps) {
  const [isMounted, setIsMounted] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [uploadProgressPct, setUploadProgressPct] = useState<number | null>(null);
  const [uploadProgressLabel, setUploadProgressLabel] = useState<string>("");
  const [isUploadingFlow, setIsUploadingFlow] = useState(false);
  const [uploadUiStatus, setUploadUiStatus] = useState<UploadUiStatus>("idle");
  const activeUploadRef = useRef<{
    fileId: string;
    uploadId: string;
    cancelRequested: boolean;
    abortRequested: boolean;
    finalized: boolean;
  } | null>(null);
  const activePartAbortControllerRef = useRef<AbortController | null>(null);

  const { toast } = useToast();
  const queryClient = useQueryClient();
  const effectiveLibraryId = libraryId.trim();

  const { mutateAsync: initUploadAsync, isPending: isInitPending } = useInitUpload();
  const { mutateAsync: presignPartAsync, isPending: isPresignPending } =
    usePresignUploadPart({
      request: {
        skipAuthRedirect: true,
      },
    });
  const { mutateAsync: confirmUploadAsync, isPending: isConfirmPending } = useConfirmUpload({
    mutation: {
      onSuccess: () => {
        if (!effectiveLibraryId) {
          return;
        }
        queryClient.invalidateQueries({
          queryKey: getLibraryTreeQueryKey(effectiveLibraryId),
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

  const isPending = isInitPending || isPresignPending || isConfirmPending;
  const isBusy = isPending || isUploadingFlow;

  const resetActiveUpload = useCallback(() => {
    activePartAbortControllerRef.current = null;
    activeUploadRef.current = null;
  }, []);

  const requestAbortOnce = useCallback(async () => {
    const activeUpload = activeUploadRef.current;
    if (!activeUpload || !activeUpload.uploadId || activeUpload.abortRequested || activeUpload.finalized) {
      return;
    }

    activeUpload.abortRequested = true;

    try {
      await abortUpload(activeUpload.fileId, activeUpload.uploadId);
    } catch (abortError) {
      console.error("Failed to abort upload session:", abortError);
    }
  }, []);

  const cancelUploadFromOutside = useCallback(async () => {
    const activeUpload = activeUploadRef.current;
    if (!activeUpload || activeUpload.finalized) {
      return;
    }

    activeUpload.cancelRequested = true;
    activePartAbortControllerRef.current?.abort();
    setUploadUiStatus("uploading");
    setUploadProgressLabel("Cancelling upload...");
    await requestAbortOnce();
    setUploadUiStatus("error");
    setUploadProgressLabel("Upload canceled.");
    setErrorMessage("Upload canceled.");
  }, [requestAbortOnce]);

  const throwIfCancelRequested = useCallback(() => {
    if (activeUploadRef.current?.cancelRequested) {
      throw new UploadCanceledError();
    }
  }, []);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!open) {
      return;
    }

    setFile(null);
    setErrorMessage(null);
    setUploadProgressPct(null);
    setUploadProgressLabel("");
    setIsUploadingFlow(false);
    setUploadUiStatus("idle");
  }, [open]);

  useEffect(() => {
    if (!onUploadProgressChange) {
      return;
    }

    const shouldShowProgress =
      uploadUiStatus === "uploading" ||
      uploadUiStatus === "error" ||
      uploadUiStatus === "success";

    onUploadProgressChange({
      visible: shouldShowProgress,
      status: uploadUiStatus,
      progressPct: uploadProgressPct,
      progressLabel: uploadProgressLabel,
      errorMessage,
      fileName: file?.name ?? null,
      canCancel: uploadUiStatus === "uploading" && isBusy,
    });
  }, [
    errorMessage,
    file,
    isBusy,
    onUploadProgressChange,
    uploadProgressLabel,
    uploadProgressPct,
    uploadUiStatus,
  ]);

  useEffect(() => {
    if (!onUploadCancelHandlerChange) {
      return;
    }

    if (uploadUiStatus === "uploading" && isBusy) {
      onUploadCancelHandlerChange(cancelUploadFromOutside);
      return;
    }

    onUploadCancelHandlerChange(null);
  }, [cancelUploadFromOutside, isBusy, onUploadCancelHandlerChange, uploadUiStatus]);

  const fileMetaText = useMemo(() => {
    if (!file) {
      return "No file selected";
    }

    const mb = (file.size / (1024 * 1024)).toFixed(2);
    return `${file.name} (${mb} MB)`;
  }, [file]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage(null);
    const parsed = uploadFormSchema.safeParse({
      libraryId: effectiveLibraryId,
      folderId: folderId?.trim() || undefined,
      file,
    });

    if (!parsed.success) {
      const message = parsed.error.issues[0]?.message ?? "Invalid upload form values.";
      setErrorMessage(message);
      toast({
        title: "Check file",
        description: message,
        variant: "error",
      });
      return;
    }

    const normalizedFolderId = parsed.data.folderId;

    setIsUploadingFlow(true);
    setUploadUiStatus("uploading");
    setUploadProgressPct(0);
    setUploadProgressLabel("Initializing upload...");
    setErrorMessage(null);

    // Hide the form while keeping upload flow running in component state.
    onClose();

    try {
      const initResponse = await initUploadAsync({
        data: {
          library_id: parsed.data.libraryId,
          folder_id: normalizedFolderId,
          name: parsed.data.file.name,
          mime_type: parsed.data.file.type || "application/octet-stream",
          size: parsed.data.file.size,
        },
      });

      const normalizedInit = normalizeInitResponse(initResponse);
      if (!normalizedInit) {
        console.error("Init upload response parse failure:", initResponse);
        throw new Error(
          `Init upload response does not include required upload metadata (${debugDescribeKeys(initResponse)}).`
        );
      }

      activeUploadRef.current = {
        fileId: normalizedInit.fileId,
        uploadId: normalizedInit.uploadId ?? "",
        cancelRequested: false,
        abortRequested: false,
        finalized: false,
      };

      throwIfCancelRequested();

      const completedParts: Array<{ part_number: number; etag: string }> = [];

      if (normalizedInit.uploadId) {
        const totalParts = Math.ceil(parsed.data.file.size / CHUNK_SIZE_BYTES);
        let uploadedBytes = 0;
        setUploadProgressPct(5);
        setUploadProgressLabel(`Uploading parts (0/${totalParts})...`);

        for (let partIndex = 0; partIndex < totalParts; partIndex += 1) {
          throwIfCancelRequested();

          const partNumber = partIndex + 1;
          const start = partIndex * CHUNK_SIZE_BYTES;
          const end = Math.min(parsed.data.file.size, start + CHUNK_SIZE_BYTES);
          const partBlob = parsed.data.file.slice(start, end);
          const checksumSha256 = await sha256Base64(partBlob);

          const presignResponse = await presignPartAsync({
            id: normalizedInit.fileId,
            data: {
              upload_id: normalizedInit.uploadId,
              part_number: partNumber,
              checksum_sha256: checksumSha256 ?? "",
              size: partBlob.size,
            },
          });

          const presignUrl = normalizePresignPartUrl(presignResponse);
          if (!presignUrl) {
            throw new Error(`Presign URL is missing for part ${partNumber}.`);
          }

          const partAbortController = new AbortController();
          activePartAbortControllerRef.current = partAbortController;

          const etag = await uploadBinaryPart(
            presignUrl,
            partBlob,
            checksumSha256,
            partAbortController.signal
          );
          completedParts.push({
            part_number: partNumber,
            etag,
          });

          activePartAbortControllerRef.current = null;
          throwIfCancelRequested();

          uploadedBytes += partBlob.size;
          const progress = Math.min(
            95,
            Math.max(5, Math.round((uploadedBytes / parsed.data.file.size) * 95))
          );
          setUploadProgressPct(progress);
          setUploadProgressLabel(`Uploading parts (${partNumber}/${totalParts})...`);
        }
      } else if (normalizedInit.presignUrl) {
        throwIfCancelRequested();
        setUploadProgressPct(15);
        setUploadProgressLabel("Uploading file...");
        const singlePartChecksum = await sha256Base64(parsed.data.file);
        const partAbortController = new AbortController();
        activePartAbortControllerRef.current = partAbortController;
        const etag = await uploadBinaryPart(
          normalizedInit.presignUrl,
          parsed.data.file,
          singlePartChecksum,
          partAbortController.signal
        );
        activePartAbortControllerRef.current = null;
        throwIfCancelRequested();
        setUploadProgressPct(90);
        setUploadProgressLabel("Upload complete. Finalizing...");
        completedParts.push({
          part_number: 1,
          etag,
        });
      } else {
        throw new Error("Init upload did not provide upload_id or presign_url.");
      }

      setUploadProgressPct(98);
      setUploadProgressLabel("Finalizing upload...");
      throwIfCancelRequested();
      const confirmResponse = await confirmUploadAsync({
        id: normalizedInit.fileId,
        data: {
          upload_id: normalizedInit.uploadId,
          completed_parts: completedParts,
        },
      });

      const status =
        typeof (confirmResponse as { status?: unknown }).status === "number"
          ? (confirmResponse as { status: number }).status
          : undefined;

      if (typeof status === "number" && status >= 400) {
        await requestAbortOnce();
        const message = formatUploadErrorByStatus(status);
        setErrorMessage(message);
        setUploadUiStatus("error");
        toast({
          title: "Upload failed",
          description: message,
          variant: "error",
        });
        return;
      }

      if (activeUploadRef.current) {
        activeUploadRef.current.finalized = true;
      }

      setUploadProgressPct(100);
      setUploadProgressLabel("Upload finished.");
      setUploadUiStatus("success");
      toast({
        title: "File uploaded",
        description: "The file is now available in library root structure.",
      });

      window.setTimeout(() => {
        setUploadUiStatus("idle");
        setUploadProgressPct(null);
        setUploadProgressLabel("");
        setErrorMessage(null);
      }, 1200);
    } catch (error) {
      const isCanceledError =
        error instanceof UploadCanceledError ||
        (error instanceof DOMException && error.name === "AbortError");

      if (isCanceledError) {
        await requestAbortOnce();
        setUploadUiStatus("error");
        setUploadProgressLabel("Upload canceled.");
        setErrorMessage("Upload canceled.");
        toast({
          title: "Upload canceled",
          description: "Upload session was canceled.",
          variant: "error",
        });
        return;
      }

      await requestAbortOnce();

      const diagnosticMessage = formatUploadErrorByClientFailure(error, parsed.data.file.size);
      if (diagnosticMessage) {
        setErrorMessage(diagnosticMessage);
        setUploadUiStatus("error");
        toast({
          title: "Upload failed",
          description: diagnosticMessage,
          variant: "error",
        });
        return;
      }

      const status =
        typeof error === "object" && error && "status" in error
          ? Number((error as { status?: unknown }).status)
          : undefined;

      const message = formatUploadErrorByStatus(Number.isFinite(status) ? status : undefined);
      setErrorMessage(message);
      setUploadUiStatus("error");
      toast({
        title: "Upload failed",
        description: message,
        variant: "error",
      });
    } finally {
      setIsUploadingFlow(false);
      resetActiveUpload();
    }
  };

  if (!isMounted || !open) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-9999 grid place-items-center bg-slate-950/70 p-4 backdrop-blur-sm">
      <Card className="w-full max-w-2xl">
        <CardHeader className="space-y-2">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="space-y-1">
              <CardTitle className="text-[1.2rem]">Upload file</CardTitle>
              <CardDescription className="text-sm text-slate-400">
                Upload file via multipart/form-data into current location.
              </CardDescription>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={onClose} disabled={isBusy}>
              Close
            </Button>
          </div>
        </CardHeader>

        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-2">
              <label className="text-sm text-slate-300">File</label>
              <input
                type="file"
                onChange={(event) => {
                  const nextFile = event.target.files?.[0] ?? null;
                  setFile(nextFile);
                }}
                className="block h-10 w-full cursor-pointer rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100 file:mr-3 file:rounded file:border-0 file:bg-slate-100 file:px-3 file:py-1 file:text-slate-900"
                required
              />
              <p className="text-xs text-slate-500">{fileMetaText}</p>
            </div>

            <p className="text-xs text-slate-500">
              library_id is set automatically. folder_id is inferred from current location.
            </p>

            {errorMessage ? <p className="text-sm text-rose-200">{errorMessage}</p> : null}

            <div className="flex flex-wrap items-center justify-end gap-3">
              <Button type="button" variant="outline" onClick={onClose} disabled={isBusy}>
                Cancel
              </Button>
              <Button type="submit" disabled={isBusy || !effectiveLibraryId}>
                {isBusy ? "Uploading..." : "Upload file"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
