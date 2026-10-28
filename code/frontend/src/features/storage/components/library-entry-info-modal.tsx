import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatBytes, formatDate } from "@/features/storage/lib/library-tree-utils";
import type { RootTreeEntry } from "@/features/storage/lib/library-tree-view-types";

type LibraryEntryInfoModalProps = {
  entry: RootTreeEntry | null;
  onClose: () => void;
};

export default function LibraryEntryInfoModal({ entry, onClose }: LibraryEntryInfoModalProps) {
  if (!entry) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 z-9999 grid place-items-center bg-slate-950/75 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <Card className="w-full max-w-md" onClick={(event) => event.stopPropagation()}>
        <CardHeader className="space-y-2">
          <div className="flex items-center justify-between gap-3">
            <CardTitle className="truncate text-lg">{entry.name}</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={onClose}>
              Close
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 text-sm text-slate-300">
            <p>
              <span className="text-slate-500">Type:</span> {entry.kind === "folder" ? "Folder" : "File"}
            </p>
            <p>
              <span className="text-slate-500">Updated:</span> {formatDate(entry.updatedAt)}
            </p>
            {entry.kind === "file" ? (
              <>
                <p>
                  <span className="text-slate-500">Size:</span> {formatBytes(entry.size)}
                </p>
                <p>
                  <span className="text-slate-500">MIME:</span> {entry.mimeType || "Unknown"}
                </p>
              </>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
