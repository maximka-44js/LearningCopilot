import Image from "next/image";

import { Button } from "@/components/ui/button";
import type { UploadProgressState } from "@/features/storage/components/upload-file-modal";

type UploadProgressBannerProps = {
  uploadProgress: UploadProgressState;
  onCancelUpload: () => void;
  onClose: () => void;
};

export default function UploadProgressBanner({
  uploadProgress,
  onCancelUpload,
  onClose,
}: UploadProgressBannerProps) {
  if (!uploadProgress.visible) {
    return null;
  }

  return (
    <div className="mb-4 rounded-lg border border-slate-800/80 bg-slate-900/45 px-3 py-3">
      <div className="flex flex-wrap items-center gap-3">
        {uploadProgress.status === "uploading" ? (
          <Image src="/icons/file-uploading-svgrepo-com.svg" alt="Uploading" width={18} height={18} />
        ) : uploadProgress.status === "error" ? (
          <Image src="/icons/fail-svgrepo-com.svg" alt="Upload failed" width={18} height={18} />
        ) : null}
        <div className="h-2 min-w-48 flex-1 overflow-hidden rounded-full bg-slate-800/80">
          <div
            className="h-full rounded-full bg-cyan-400 transition-all duration-300"
            style={{ width: `${uploadProgress.progressPct ?? 0}%` }}
          />
        </div>
        {uploadProgress.canCancel ? (
          <Button type="button" size="sm" variant="outline" onClick={onCancelUpload}>
            Cancel upload
          </Button>
        ) : null}
        {uploadProgress.status !== "uploading" ? (
          <Button type="button" size="sm" variant="outline" onClick={onClose}>
            Close
          </Button>
        ) : null}
      </div>
      <p className="mt-2 text-xs text-slate-400">
        {uploadProgress.progressLabel || "Preparing upload..."}
        {typeof uploadProgress.progressPct === "number" ? ` ${uploadProgress.progressPct}%` : ""}
        {uploadProgress.fileName ? ` - ${uploadProgress.fileName}` : ""}
      </p>
    </div>
  );
}
