import Image from "next/image";

import { truncateEntryName } from "@/features/storage/lib/library-tree-utils";
import type { RootTreeEntry } from "@/features/storage/lib/library-tree-view-types";

type LibraryTreeEntriesTableProps = {
  entries: RootTreeEntry[];
  selectedEntryIds: Record<string, boolean>;
  useSelectionColumn: boolean;
  headerGridClass: string;
  rowGridClass: string;
  isActionLocked: boolean;
  cutClipboardKeys: Set<string>;
  onRowClick: (entry: RootTreeEntry, index: number, event: React.MouseEvent<HTMLElement>) => void;
  onRowDoubleClick: (entry: RootTreeEntry) => void;
  onRowTouchStart: (entry: RootTreeEntry, index: number) => void;
  onClearLongPressTimer: () => void;
  onToggleEntrySelection: (entryId: string, checked: boolean) => void;
  onOpenEntryInfo: (entry: RootTreeEntry) => void;
};

export default function LibraryTreeEntriesTable({
  entries,
  selectedEntryIds,
  useSelectionColumn,
  headerGridClass,
  rowGridClass,
  isActionLocked,
  cutClipboardKeys,
  onRowClick,
  onRowDoubleClick,
  onRowTouchStart,
  onClearLongPressTimer,
  onToggleEntrySelection,
  onOpenEntryInfo,
}: LibraryTreeEntriesTableProps) {
  return (
    <div className="overflow-x-auto rounded-lg border border-slate-800/80 bg-slate-950/30">
      <div className="w-full min-w-130 sm:min-w-0">
        <div
          className={`${headerGridClass} border-b border-slate-800/80 bg-slate-900/40 px-4 py-2 text-xs uppercase tracking-[0.15em] text-slate-500`}
        >
          {useSelectionColumn ? <p>Select</p> : null}
          <p>Name</p>
          <p className="text-right">Info</p>
        </div>

        <ul>
          {entries.map((entry, index) => {
            const isSelected = Boolean(selectedEntryIds[entry.id]);
            const entryClipboardKey =
              entry.kind === "folder"
                ? `folder:${entry.sourceFolderId ?? entry.id}`
                : `file:${entry.id}`;
            const isCutPending = cutClipboardKeys.has(entryClipboardKey);
            const displayName = truncateEntryName(entry.name, entry.kind === "file" ? 56 : 64);

            return (
              <li
                key={entry.id}
                className={`${rowGridClass} items-center gap-3 border-b border-slate-800/60 px-4 py-3 text-sm last:border-b-0 cursor-pointer ${
                  isSelected ? "bg-slate-800/80" : "hover:bg-slate-900/30"
                } ${isCutPending ? "opacity-55" : ""}`}
                onClick={(event) => onRowClick(entry, index, event)}
                onDoubleClick={() => onRowDoubleClick(entry)}
                onTouchStart={() => onRowTouchStart(entry, index)}
                onTouchEnd={onClearLongPressTimer}
                onTouchCancel={onClearLongPressTimer}
                onTouchMove={onClearLongPressTimer}
              >
                {useSelectionColumn ? (
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={(event) => onToggleEntrySelection(entry.id, event.target.checked)}
                      onClick={(event) => event.stopPropagation()}
                      className="h-4 w-4 rounded border-slate-700 bg-slate-900 opacity-60"
                      aria-label={`Select ${entry.name}`}
                      disabled={isActionLocked}
                    />
                  </div>
                ) : null}

                <div className="flex min-w-0 items-center gap-3">
                  {entry.kind === "folder" ? (
                    <Image src="/icons/folder-svgrepo-com.svg" alt="Folder" width={18} height={18} />
                  ) : (
                    <Image
                      src="/icons/file-svgrepo-com.svg"
                      alt="File"
                      width={18}
                      height={18}
                    />
                  )}
                  <p className="truncate font-medium text-slate-100" title={entry.name}>
                    {displayName}
                  </p>
                  {isCutPending ? (
                    <span className="rounded border border-amber-400/35 bg-amber-400/10 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.08em] text-amber-200">
                      Cut
                    </span>
                  ) : null}
                </div>

                <div
                  className="flex w-full items-center justify-end text-slate-400"
                  onClick={(event) => event.stopPropagation()}
                >
                  <button
                    type="button"
                    onClick={() => onOpenEntryInfo(entry)}
                    className="inline-flex h-7 w-7 items-center justify-center rounded-md border border-slate-700/70 bg-slate-900/40 transition hover:bg-slate-800/60"
                    aria-label={`Open info for ${entry.name}`}
                  >
                    <Image
                      src="/icons/info-circle-svgrepo-com.svg"
                      alt="Info"
                      width={15}
                      height={15}
                    />
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
