import { useCallback, useEffect, useMemo, useState } from "react";

import type { RootTreeEntry, StorageClipboard } from "@/features/storage/lib/library-tree-view-types";
import { copyFileById, copyFolderById } from "@/features/storage/lib/library-tree-operations";
import type { UpdateFileMutationBody } from "@/generated/api/files/files";
import type { UpdateFolderMutationBody } from "@/generated/api/folders/folders";

type ToastFn = (options: {
  title: string;
  description?: string;
  variant?: "default" | "error";
}) => void;

type UpdateFileFn = (params: {
  id: string;
  data: UpdateFileMutationBody;
}) => Promise<unknown>;

type UpdateFolderFn = (params: {
  id: string;
  data: UpdateFolderMutationBody;
}) => Promise<unknown>;

type UseLibraryTreeClipboardParams = {
  currentFolderId: string | null;
  effectiveLibraryId: string;
  hasSelectedEntries: boolean;
  invalidateTreeData: () => Promise<void>;
  isActionLockedBase: boolean;
  resetSelection: () => void;
  selectedEntries: RootTreeEntry[];
  toast: ToastFn;
  updateFileAsync: UpdateFileFn;
  updateFolderAsync: UpdateFolderFn;
};

export function useLibraryTreeClipboard(params: UseLibraryTreeClipboardParams) {
  const {
    currentFolderId,
    effectiveLibraryId,
    hasSelectedEntries,
    invalidateTreeData,
    isActionLockedBase,
    resetSelection,
    selectedEntries,
    toast,
    updateFileAsync,
    updateFolderAsync,
  } = params;

  const [storageClipboard, setStorageClipboard] = useState<StorageClipboard | null>(null);
  const [isClipboardPasteRunning, setIsClipboardPasteRunning] = useState(false);
  const isActionLocked = isActionLockedBase || isClipboardPasteRunning;

  const canPaste =
    Boolean(storageClipboard?.entries.length) &&
    storageClipboard?.libraryId === effectiveLibraryId &&
    !isClipboardPasteRunning;

  const isCutClipboard = storageClipboard?.mode === "cut";

  const cutClipboardKeys = useMemo(() => {
    if (storageClipboard?.mode !== "cut") {
      return new Set<string>();
    }

    return new Set(
      storageClipboard.entries.map((entry) =>
        entry.kind === "folder"
          ? `folder:${entry.sourceFolderId ?? entry.id}`
          : `file:${entry.id}`
      )
    );
  }, [storageClipboard]);

  const handleCopySelected = useCallback(() => {
    if (selectedEntries.length === 0) {
      return;
    }

    setStorageClipboard({
      libraryId: effectiveLibraryId,
      mode: "copy",
      entries: selectedEntries.map((entry) => ({
        id: entry.id,
        name: entry.name,
        kind: entry.kind,
        sourceFolderId: entry.sourceFolderId,
        sourceParentId: entry.sourceParentId,
      })),
    });

    resetSelection();

    toast({
      title: "Copied",
      description: `${selectedEntries.length} item(s) copied to clipboard.`,
    });
  }, [effectiveLibraryId, resetSelection, selectedEntries, toast]);

  const handleCutSelected = useCallback(() => {
    if (selectedEntries.length === 0) {
      return;
    }

    setStorageClipboard({
      libraryId: effectiveLibraryId,
      mode: "cut",
      entries: selectedEntries.map((entry) => ({
        id: entry.id,
        name: entry.name,
        kind: entry.kind,
        sourceFolderId: entry.sourceFolderId,
        sourceParentId: entry.sourceParentId,
      })),
    });

    resetSelection();

    toast({
      title: "Cut",
      description: `${selectedEntries.length} item(s) ready to move.`,
    });
  }, [effectiveLibraryId, resetSelection, selectedEntries, toast]);

  const handlePasteClipboard = useCallback(async () => {
    if (!storageClipboard || storageClipboard.entries.length === 0) {
      return;
    }

    if (storageClipboard.libraryId !== effectiveLibraryId) {
      toast({
        title: "Cannot paste",
        description: "Paste is allowed only inside the same library.",
        variant: "error",
      });
      return;
    }

    const targetFolderId = currentFolderId ?? undefined;

    if (storageClipboard.mode === "cut") {
      const movableEntries = storageClipboard.entries.filter((entry) => {
        if (entry.kind === "folder") {
          return entry.sourceParentId !== targetFolderId;
        }

        return entry.sourceFolderId !== targetFolderId;
      });

      if (movableEntries.length === 0) {
        toast({
          title: "Nothing to move",
          description: "Selected items are already in this folder.",
        });
        return;
      }

      setIsClipboardPasteRunning(true);

      try {
        await Promise.all(
          movableEntries.map((entry) => {
            if (entry.kind === "folder") {
              const folderId = entry.sourceFolderId || entry.id;
              return updateFolderAsync({
                id: folderId,
                data: {
                  name: entry.name,
                  parent_id: targetFolderId,
                },
              });
            }

            return updateFileAsync({
              id: entry.id,
              data: {
                name: entry.name,
                folder_id: targetFolderId,
              },
            });
          })
        );

        await invalidateTreeData();
        resetSelection();
        setStorageClipboard(null);
        toast({
          title: "Move complete",
          description: `${movableEntries.length} item(s) moved.`,
        });
      } catch {
        toast({
          title: "Move failed",
          description: "Unable to move selected items right now.",
          variant: "error",
        });
      } finally {
        setIsClipboardPasteRunning(false);
      }

      return;
    }

    setIsClipboardPasteRunning(true);

    try {
      await Promise.all(
        storageClipboard.entries.map((entry) => {
          if (entry.kind === "folder") {
            return copyFolderById(entry.id, {
              target_parent_id: targetFolderId,
            });
          }

          return copyFileById(entry.id, {
            target_folder_id: targetFolderId,
          });
        })
      );

      await invalidateTreeData();
      toast({
        title: "Paste complete",
        description: `${storageClipboard.entries.length} item(s) inserted.`,
      });
    } catch {
      toast({
        title: "Paste failed",
        description: "Unable to paste selected items right now.",
        variant: "error",
      });
    } finally {
      setIsClipboardPasteRunning(false);
    }
  }, [
    currentFolderId,
    effectiveLibraryId,
    invalidateTreeData,
    resetSelection,
    storageClipboard,
    toast,
    updateFileAsync,
    updateFolderAsync,
  ]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      const tagName = target?.tagName.toLowerCase();
      const isEditable =
        Boolean(target?.isContentEditable) ||
        tagName === "input" ||
        tagName === "textarea" ||
        tagName === "select";

      if (isEditable) {
        return;
      }

      const isMeta = event.ctrlKey || event.metaKey;
      if (!isMeta) {
        return;
      }

      const key = event.key.toLowerCase();
      if (key === "c") {
        if (!hasSelectedEntries || isActionLocked) {
          return;
        }
        event.preventDefault();
        handleCopySelected();
        return;
      }

      if (key === "x") {
        if (!hasSelectedEntries || isActionLocked) {
          return;
        }
        event.preventDefault();
        handleCutSelected();
        return;
      }

      if (key === "v") {
        if (!canPaste || isActionLocked) {
          return;
        }
        event.preventDefault();
        void handlePasteClipboard();
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [
    canPaste,
    handleCopySelected,
    handleCutSelected,
    handlePasteClipboard,
    hasSelectedEntries,
    isActionLocked,
  ]);

  return {
    canPaste,
    cutClipboardKeys,
    handleCopySelected,
    handleCutSelected,
    handlePasteClipboard,
    isClipboardPasteRunning,
    isCutClipboard,
  };
}
