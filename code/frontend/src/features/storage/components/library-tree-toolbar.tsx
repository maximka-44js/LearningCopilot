import { Button } from "@/components/ui/button";
import { CardTitle } from "@/components/ui/card";
import type { FolderTrailItem } from "@/features/storage/lib/library-tree-view-types";

type LibraryTreeToolbarProps = {
  isBrowsingFolder: boolean;
  folderTrail: FolderTrailItem[];
  onGoToRoot: () => void;
  onGoToTrailIndex: (index: number) => void;
  selectedSingleEntry: boolean;
  hasSelectedEntries: boolean;
  canPaste: boolean;
  isCutClipboard: boolean;
  isClipboardPasteRunning: boolean;
  isDeleteRunning: boolean;
  canDownloadSelectedFile: boolean;
  isDownloadStarting: boolean;
  isActionLocked: boolean;
  isUploadRunning: boolean;
  isRenameRunning: boolean;
  onRename: () => void;
  onCopy: () => void;
  onCut: () => void;
  onPaste: () => void;
  onDelete: () => void;
  onDownload: () => void;
  onCreateFolder: () => void;
  onUploadFile: () => void;
};

export default function LibraryTreeToolbar({
  isBrowsingFolder,
  folderTrail,
  onGoToRoot,
  onGoToTrailIndex,
  selectedSingleEntry,
  hasSelectedEntries,
  canPaste,
  isCutClipboard,
  isClipboardPasteRunning,
  isDeleteRunning,
  canDownloadSelectedFile,
  isDownloadStarting,
  isActionLocked,
  isUploadRunning,
  isRenameRunning,
  onRename,
  onCopy,
  onCut,
  onPaste,
  onDelete,
  onDownload,
  onCreateFolder,
  onUploadFile,
}: LibraryTreeToolbarProps) {
  return (
    <>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="space-y-2">
          <CardTitle className="text-lg">
            {isBrowsingFolder ? "Folder contents" : "Root structure"}
          </CardTitle>
          <div className="flex flex-wrap items-center gap-2 text-sm text-slate-400">
            <button
              type="button"
              onClick={onGoToRoot}
              className="rounded border border-slate-700 px-2 py-0.5 transition hover:bg-slate-900/50"
            >
              Root
            </button>
            {folderTrail.map((item, index) => (
              <div key={item.id} className="flex items-center gap-2">
                <span className="text-slate-600">/</span>
                <button
                  type="button"
                  onClick={() => onGoToTrailIndex(index)}
                  className="max-w-48 truncate rounded border border-slate-700 px-2 py-0.5 transition hover:bg-slate-900/50"
                  title={item.name}
                >
                  {item.name}
                </button>
              </div>
            ))}
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onRename}
            disabled={!selectedSingleEntry || isActionLocked}
            title="Rename selected"
          >
            {isRenameRunning ? "Renaming..." : "Rename"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onCopy}
            disabled={!hasSelectedEntries || isActionLocked}
            title="Copy selected (Ctrl/Cmd+C)"
          >
            Copy
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onCut}
            disabled={!hasSelectedEntries || isActionLocked}
            title="Cut selected (Ctrl/Cmd+X)"
          >
            Cut
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onPaste}
            disabled={!canPaste || isActionLocked}
            title={isCutClipboard ? "Move here (Ctrl/Cmd+V)" : "Paste (Ctrl/Cmd+V)"}
          >
            {isClipboardPasteRunning
              ? isCutClipboard
                ? "Moving..."
                : "Pasting..."
              : isCutClipboard
                ? "Move here"
                : "Paste"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onDelete}
            disabled={!hasSelectedEntries || isActionLocked}
            title="Delete selected"
            className="hover:border-rose-500/50 hover:bg-rose-500/20 hover:text-rose-100"
          >
            {isDeleteRunning ? "Deleting..." : "Delete"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onDownload}
            disabled={!canDownloadSelectedFile || isActionLocked || isDownloadStarting}
            title="Download selected file"
          >
            {isDownloadStarting ? "Starting..." : "Download"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onCreateFolder}
            disabled={isActionLocked}
          >
            Create folder
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onUploadFile}
            disabled={isUploadRunning}
          >
            Upload file
          </Button>
        </div>
      </div>
    </>
  );
}
