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
            {isBrowsingFolder ? "Содержимое папки" : "Структура"}
          </CardTitle>
          <div className="flex flex-wrap items-center gap-2 text-sm text-slate-400">
            <button
              type="button"
              onClick={onGoToRoot}
              className="rounded border border-slate-700 px-2 py-0.5 transition hover:bg-slate-900/50"
            >
              root
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
            title="Переименовать выбранное"
          >
            {isRenameRunning ? "Переименование..." : "Переименовать"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onCopy}
            disabled={!hasSelectedEntries || isActionLocked}
            title="Копировать выбранное (Ctrl/Cmd+C)"
          >
            Копировать
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onCut}
            disabled={!hasSelectedEntries || isActionLocked}
            title="Вырезать выбранное (Ctrl/Cmd+X)"
          >
            Вырезать
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onPaste}
            disabled={!canPaste || isActionLocked}
            title={isCutClipboard ? "Переместить сюда (Ctrl/Cmd+V)" : "Вставить (Ctrl/Cmd+V)"}
          >
            {isClipboardPasteRunning
              ? isCutClipboard
                ? "Перемещение..."
                : "Вставка..."
              : isCutClipboard
                ? "Переместить сюда"
                : "Вставить"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onDelete}
            disabled={!hasSelectedEntries || isActionLocked}
            title="Удалить выбранное"
            className="hover:border-rose-500/50 hover:bg-rose-500/20 hover:text-rose-100"
          >
            {isDeleteRunning ? "Удаление..." : "Удалить"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onDownload}
            disabled={!canDownloadSelectedFile || isActionLocked || isDownloadStarting}
            title="Скачать выбранный файл"
          >
            {isDownloadStarting ? "Подготовка..." : "Скачать"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onCreateFolder}
            disabled={isActionLocked}
          >
            Создать папку
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onUploadFile}
            disabled={isUploadRunning}
          >
            Загрузить файл
          </Button>
        </div>
      </div>
    </>
  );
}
