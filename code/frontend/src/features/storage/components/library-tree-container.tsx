"use client";

import type { QueryClient } from "@tanstack/react-query";
import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState } from "react";

import { useToast } from "@/components/providers/toast-provider";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import CreateFolderModal from "@/features/storage/components/create-folder-modal";
import DeleteLibraryModal from "@/features/storage/components/delete-library-modal";
import EditLibraryModal from "@/features/storage/components/edit-library-modal";
import LibraryDetailsCards from "@/features/storage/components/library-details-cards";
import LibraryEntryInfoModal from "@/features/storage/components/library-entry-info-modal";
import LibraryTreeEntriesTable from "@/features/storage/components/library-tree-entries-table";
import LibraryTreePageHeader from "@/features/storage/components/library-tree-page-header";
import LibraryTreeToolbar from "@/features/storage/components/library-tree-toolbar";
import UploadFileModal, {
    type UploadProgressState,
} from "@/features/storage/components/upload-file-modal";
import UploadProgressBanner from "@/features/storage/components/upload-progress-banner";
import {
    getLibraryTreeQueryKey,
    useLibraryTree,
} from "@/features/storage/lib/library-tree";
import {
    deleteFolderById,
} from "@/features/storage/lib/library-tree-operations";
import {
    isValidStorageName,
} from "@/features/storage/lib/library-tree-utils";
import {
    buildLibraryListItem,
    buildRootEntries,
    normalizeFolderData,
    normalizeLibraryData,
    normalizeTreeData,
} from "@/features/storage/lib/library-tree-view-model";
import type {
    FolderTrailItem,
    RootTreeEntry,
} from "@/features/storage/lib/library-tree-view-types";
import { useLibraryTreeClipboard } from "@/features/storage/lib/use-library-tree-clipboard";
import { useLibraryTreeSelection } from "@/features/storage/lib/use-library-tree-selection";
import {
    useDeleteFile,
    useUpdateFile,
} from "@/generated/api/files/files";
import {
    getGetFolderContentsQueryKey,
    useGetFolderById,
    useGetFolderContents,
    useUpdateFolder,
} from "@/generated/api/folders/folders";
import { useGetLibraryById } from "@/generated/api/libraries/libraries";
import { axiosInstance } from "@/lib/axios";

type LibraryTreeContainerProps = {
  libraryId: string;
};

const defaultUploadProgressState: UploadProgressState = {
  visible: false,
  status: "idle",
  progressPct: null,
  progressLabel: "",
  errorMessage: null,
  fileName: null,
  canCancel: false,
};

async function invalidateTreeQueries(params: {
  currentFolderId: string | null;
  effectiveLibraryId: string;
  queryClient: QueryClient;
}) {
  const { currentFolderId, effectiveLibraryId, queryClient } = params;

  if (effectiveLibraryId) {
    await queryClient.invalidateQueries({
      queryKey: getLibraryTreeQueryKey(effectiveLibraryId),
    });
  }

  const normalizedCurrentFolderId = currentFolderId?.trim();
  if (normalizedCurrentFolderId) {
    await queryClient.invalidateQueries({
      queryKey: getGetFolderContentsQueryKey(normalizedCurrentFolderId),
    });
  }
}

export default function LibraryTreeContainer({ libraryId }: LibraryTreeContainerProps) {
  const [isCreateFolderOpen, setIsCreateFolderOpen] = useState(false);
  const [isUploadFileOpen, setIsUploadFileOpen] = useState(false);
  const [isEditLibraryOpen, setIsEditLibraryOpen] = useState(false);
  const [isDeleteLibraryOpen, setIsDeleteLibraryOpen] = useState(false);
  const [activeInfoEntry, setActiveInfoEntry] = useState<RootTreeEntry | null>(null);
  const [currentFolderId, setCurrentFolderId] = useState<string | null>(null);
  const [folderTrail, setFolderTrail] = useState<FolderTrailItem[]>([]);
  const [uploadProgress, setUploadProgress] = useState<UploadProgressState>(
    defaultUploadProgressState
  );
  const [cancelUploadHandler, setCancelUploadHandler] = useState<
    (() => Promise<void>) | null
  >(null);
  const [isDeleteRunning, setIsDeleteRunning] = useState(false);
  const [isDownloadStarting, setIsDownloadStarting] = useState(false);
  const [isRenameRunning, setIsRenameRunning] = useState(false);

  const fileSystemCardRef = useRef<HTMLDivElement | null>(null);
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const normalizedLibraryId = libraryId.trim();
  const {
    data: treeData,
    isLoading: isTreeLoading,
    isError: isTreeError,
  } = useLibraryTree(normalizedLibraryId);
  const {
    data: libraryData,
    isLoading: isLibraryLoading,
    isError: isLibraryError,
  } = useGetLibraryById(normalizedLibraryId, {
    query: {
      enabled: normalizedLibraryId.length > 0,
    },
  });
  const {
    data: folderData,
    isLoading: isFolderLoading,
    isError: isFolderError,
  } = useGetFolderById(currentFolderId || "", {
    query: {
      enabled: Boolean(currentFolderId),
    },
  });
  const {
    data: folderTreeData,
    isLoading: isFolderTreeLoading,
    isError: isFolderTreeError,
  } = useGetFolderContents(currentFolderId || "", {
    query: {
      enabled: Boolean(currentFolderId),
    },
  });

  const { mutateAsync: deleteFileAsync } = useDeleteFile();
  const { mutateAsync: updateFileAsync } = useUpdateFile();
  const { mutateAsync: updateFolderAsync } = useUpdateFolder();

  const isBrowsingFolder = Boolean(currentFolderId);
  const tree = normalizeTreeData(isBrowsingFolder ? folderTreeData : treeData);
  const library = normalizeLibraryData(libraryData);
  const currentFolder = normalizeFolderData(folderData);

  const effectiveLibraryId =
    normalizedLibraryId || tree?.library_id?.trim() || library?.id?.trim() || "";
  const currentLibraryItem = buildLibraryListItem({
    effectiveLibraryId,
    library,
  });
  const accessesCount = Array.isArray(library?.accesses) ? library.accesses.length : 0;

  const rootEntries = buildRootEntries(tree);
  const isRootEmpty = rootEntries.length === 0;
  const rootEntriesById = new Map(rootEntries.map((entry) => [entry.id, entry]));

  const openFolder = useCallback((entry: RootTreeEntry) => {
    const folderId = entry.sourceFolderId;
    if (entry.kind !== "folder" || !folderId) {
      return;
    }

    setCurrentFolderId(folderId);
    setFolderTrail((current) => {
      if (current[current.length - 1]?.id === folderId) {
        return current;
      }
      return [...current, { id: folderId, name: entry.name }];
    });
  }, []);

  const {
    clearLongPressTimer,
    handleRowClick: handleSelectionRowClick,
    handleRowDoubleClick: handleSelectionRowDoubleClick,
    hasSelectedEntries,
    isMobileSelectionMode,
    isTouchDevice,
    resetSelection,
    selectedEntries,
    selectedEntryIds,
    selectedSingleEntry,
    startLongPressSelection: startSelectionLongPress,
    toggleEntrySelection,
  } = useLibraryTreeSelection({
    fileSystemCardRef,
    onOpenFolder: openFolder,
    rootEntries,
    rootEntriesById,
  });

  const selectedFileForDownload =
    selectedEntries.length === 1 && selectedEntries[0]?.kind === "file"
      ? selectedEntries[0]
      : null;
  const isUploadRunning = uploadProgress.status === "uploading";
  const isActionLockedBase =
    isUploadRunning || isDeleteRunning || isRenameRunning;

  const {
    canPaste,
    cutClipboardKeys,
    handleCopySelected,
    handleCutSelected,
    handlePasteClipboard,
    isClipboardPasteRunning,
    isCutClipboard,
  } = useLibraryTreeClipboard({
    currentFolderId,
    effectiveLibraryId,
    hasSelectedEntries,
    invalidateTreeData: () =>
      invalidateTreeQueries({
        currentFolderId,
        effectiveLibraryId,
        queryClient,
      }),
    isActionLockedBase,
    resetSelection,
    selectedEntries,
    toast,
    updateFileAsync,
    updateFolderAsync,
  });

  const isActionLocked = isActionLockedBase || isClipboardPasteRunning;
  const canDownloadSelectedFile = Boolean(selectedFileForDownload);

  const handleFutureAction = (action: "folder" | "file") => {
    if (action === "folder") {
      setIsCreateFolderOpen(true);
      return;
    }

    setIsUploadFileOpen(true);
  };

  const goToRoot = () => {
    setCurrentFolderId(null);
    setFolderTrail([]);
    resetSelection();
  };

  const goToTrailIndex = (index: number) => {
    if (index < 0) {
      goToRoot();
      return;
    }

    const nextTrail = folderTrail.slice(0, index + 1);
    setFolderTrail(nextTrail);
    setCurrentFolderId(nextTrail[nextTrail.length - 1]?.id ?? null);
    resetSelection();
  };

  const openEntryInfoModal = (entry: RootTreeEntry) => {
    setActiveInfoEntry(entry);
  };

  const handleDeleteSelected = useCallback(async () => {
    if (selectedEntries.length === 0) {
      return;
    }

    const confirmed = window.confirm(
      `Delete ${selectedEntries.length} selected item(s)? This action cannot be undone.`
    );
    if (!confirmed) {
      return;
    }

    setIsDeleteRunning(true);

    try {
      const results = await Promise.allSettled(
        selectedEntries.map((entry) => {
          const targetId = entry.kind === "folder" ? entry.sourceFolderId : entry.id;
          if (!targetId) {
            throw new Error("Missing entry id");
          }

          if (entry.kind === "folder") {
            return deleteFolderById(targetId);
          }

          return deleteFileAsync({ id: targetId });
        })
      );

      const successCount = results.filter((result) => result.status === "fulfilled").length;
      const failCount = results.length - successCount;

      if (successCount > 0) {
        await invalidateTreeQueries({
          currentFolderId,
          effectiveLibraryId,
          queryClient,
        });
        resetSelection();
      }

      if (successCount > 0 && failCount === 0) {
        toast({
          title: "Delete complete",
          description: `${successCount} item(s) deleted.`,
        });
        return;
      }

      if (successCount > 0 && failCount > 0) {
        toast({
          title: "Delete partially complete",
          description: `${successCount} deleted, ${failCount} failed.`,
          variant: "error",
        });
        return;
      }

      toast({
        title: "Delete failed",
        description: "Unable to delete selected items right now.",
        variant: "error",
      });
    } finally {
      setIsDeleteRunning(false);
    }
  }, [
    currentFolderId,
    deleteFileAsync,
    effectiveLibraryId,
    queryClient,
    resetSelection,
    selectedEntries,
    toast,
  ]);

  const handleRenameSelected = async () => {
    if (!selectedSingleEntry || isRenameRunning) {
      return;
    }

    const currentName = selectedSingleEntry.name;
    const nextNameRaw = window.prompt("Rename item", currentName);
    if (nextNameRaw === null) {
      return;
    }

    const nextName = nextNameRaw.trim();
    if (!nextName) {
      toast({
        title: "Rename failed",
        description: "Name is required.",
        variant: "error",
      });
      return;
    }

    if (nextName.length > 100 || !isValidStorageName(nextName)) {
      toast({
        title: "Rename failed",
        description:
          "Invalid name. Use up to 100 characters and avoid \\ / : * ? \" < > |.",
        variant: "error",
      });
      return;
    }

    if (nextName === currentName) {
      return;
    }

    setIsRenameRunning(true);

    try {
      if (selectedSingleEntry.kind === "file") {
        await updateFileAsync({
          id: selectedSingleEntry.id,
          data: {
            name: nextName,
            folder_id: selectedSingleEntry.sourceFolderId,
          },
        });
      } else {
        const folderId = selectedSingleEntry.sourceFolderId;
        if (!folderId) {
          throw new Error("Missing folder id");
        }

        await updateFolderAsync({
          id: folderId,
          data: {
            name: nextName,
            parent_id: selectedSingleEntry.sourceParentId,
          },
        });
      }

      await invalidateTreeQueries({
        currentFolderId,
        effectiveLibraryId,
        queryClient,
      });
      toast({
        title: "Renamed",
        description: `\"${currentName}\" renamed to \"${nextName}\".`,
      });
    } catch {
      toast({
        title: "Rename failed",
        description: "Unable to rename selected item right now.",
        variant: "error",
      });
    } finally {
      setIsRenameRunning(false);
    }
  };

  const handleDownloadSelected = useCallback(async () => {
    if (!selectedFileForDownload || isDownloadStarting) {
      return;
    }

    const fileId = selectedFileForDownload.id;
    setIsDownloadStarting(true);
    console.info("download_clicked", { fileId });
    toast({
      title: "Download started",
      description: "Preparing file download...",
    });

    try {
      const response = await axiosInstance.get<{
        URL?: string;
        url?: string;
      }>(`/storage/files/${fileId}/download`);
      const downloadUrl = response.data.URL || response.data.url;

      if (!downloadUrl) {
        throw new Error("Download URL is missing.");
      }

      const anchor = document.createElement("a");
      anchor.href = downloadUrl;
      anchor.download = selectedFileForDownload.name || "download.bin";
      anchor.rel = "noopener";
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
    } catch (error) {
      const status =
        typeof error === "object" &&
        error &&
        "response" in error &&
        typeof (error as { response?: { status?: unknown } }).response?.status === "number"
          ? Number((error as { response?: { status?: unknown } }).response?.status)
          : undefined;

      console.warn("download_failed", { fileId, status, error });

      if (status === 401) {
        toast({
          title: "Download failed",
          description: "Session expired. Please sign in again.",
          variant: "error",
        });
      } else if (status === 403) {
        toast({
          title: "Download failed",
          description: "No access to this file.",
          variant: "error",
        });
      } else if (status === 404) {
        toast({
          title: "Download failed",
          description: "File was not found.",
          variant: "error",
        });
      } else if (status === 400) {
        toast({
          title: "Download failed",
          description: "This file is not available because its upload was not completed.",
          variant: "error",
        });
      } else {
        toast({
          title: "Download failed",
          description: "Temporary error. Please try again later.",
          variant: "error",
        });
      }
    } finally {
      window.setTimeout(() => {
        setIsDownloadStarting(false);
      }, 1200);
    }
  }, [isDownloadStarting, selectedFileForDownload, toast]);

  const handleRowClick = useCallback(
    (entry: RootTreeEntry, index: number, event: React.MouseEvent<HTMLElement>) => {
      if (isActionLocked) {
        return;
      }

      handleSelectionRowClick(entry, index, event);
    },
    [handleSelectionRowClick, isActionLocked]
  );

  const handleRowDoubleClick = useCallback(
    (entry: RootTreeEntry) => {
      if (isActionLocked) {
        return;
      }

      handleSelectionRowDoubleClick(entry);
    },
    [handleSelectionRowDoubleClick, isActionLocked]
  );

  const handleRowTouchStart = useCallback(
    (entry: RootTreeEntry, index: number) => {
      if (isActionLocked) {
        return;
      }

      startSelectionLongPress(entry, index);
    },
    [isActionLocked, startSelectionLongPress]
  );

  const clearUploadProgress = () => {
    setUploadProgress(defaultUploadProgressState);
    setCancelUploadHandler(null);
  };

  const showSelectionCheckboxes = isTouchDevice && isMobileSelectionMode;
  const useSelectionColumn = showSelectionCheckboxes;
  const headerGridClass = useSelectionColumn
    ? "grid grid-cols-[56px_minmax(0,2fr)_minmax(90px,0.8fr)_minmax(120px,1fr)] sm:grid-cols-[72px_minmax(0,2.2fr)_minmax(120px,0.9fr)_minmax(180px,1fr)]"
    : "grid grid-cols-[minmax(0,2fr)_minmax(90px,0.8fr)_minmax(120px,1fr)] sm:grid-cols-[minmax(0,2.2fr)_minmax(120px,0.9fr)_minmax(180px,1fr)]";
  const rowGridClass = useSelectionColumn
    ? "grid grid-cols-[56px_minmax(0,2fr)_minmax(90px,0.8fr)_minmax(120px,1fr)] sm:grid-cols-[72px_minmax(0,2.2fr)_minmax(120px,0.9fr)_minmax(180px,1fr)]"
    : "grid grid-cols-[minmax(0,2fr)_minmax(90px,0.8fr)_minmax(120px,1fr)] sm:grid-cols-[minmax(0,2.2fr)_minmax(120px,0.9fr)_minmax(180px,1fr)]";

  useEffect(() => {
    resetSelection();
  }, [currentFolderId, resetSelection]);

  if (!normalizedLibraryId) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Library ID is empty. Reload the page and try again.
      </div>
    );
  }

  if (
    isLibraryLoading ||
    (!isBrowsingFolder && isTreeLoading) ||
    (isBrowsingFolder && (isFolderLoading || isFolderTreeLoading))
  ) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading...
      </div>
    );
  }

  if (
    isLibraryError ||
    (!isBrowsingFolder && isTreeError) ||
    (isBrowsingFolder && (isFolderError || isFolderTreeError)) ||
    !tree ||
    !library
  ) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Unable to load folder information right now.
      </div>
    );
  }

  return (
    <section className="space-y-6">
      <LibraryTreePageHeader
        title={currentLibraryItem.title}
        libraryId={normalizedLibraryId}
        folderLabel={
          isBrowsingFolder
            ? `Folder: ${currentFolder?.name?.trim() || folderTrail[folderTrail.length - 1]?.name || "Unknown"}`
            : "Root: Top-level"
        }
      />

      <LibraryDetailsCards
        library={library}
        accessesCount={accessesCount}
        onEdit={() => setIsEditLibraryOpen(true)}
        onDelete={() => setIsDeleteLibraryOpen(true)}
      />

      <Card ref={fileSystemCardRef}>
        <CardHeader>
          <LibraryTreeToolbar
            isBrowsingFolder={isBrowsingFolder}
            folderTrail={folderTrail}
            onGoToRoot={goToRoot}
            onGoToTrailIndex={goToTrailIndex}
            selectedSingleEntry={Boolean(selectedSingleEntry)}
            hasSelectedEntries={hasSelectedEntries}
            canPaste={canPaste}
            isCutClipboard={Boolean(isCutClipboard)}
            isClipboardPasteRunning={isClipboardPasteRunning}
            isDeleteRunning={isDeleteRunning}
            canDownloadSelectedFile={canDownloadSelectedFile}
            isDownloadStarting={isDownloadStarting}
            isActionLocked={isActionLocked}
            isUploadRunning={isUploadRunning}
            isRenameRunning={isRenameRunning}
            onRename={() => {
              void handleRenameSelected();
            }}
            onCopy={handleCopySelected}
            onCut={handleCutSelected}
            onPaste={() => {
              void handlePasteClipboard();
            }}
            onDelete={() => {
              void handleDeleteSelected();
            }}
            onDownload={() => {
              void handleDownloadSelected();
            }}
            onCreateFolder={() => handleFutureAction("folder")}
            onUploadFile={() => handleFutureAction("file")}
          />
        </CardHeader>
        <CardContent className="pt-0">
          <UploadProgressBanner
            uploadProgress={uploadProgress}
            onCancelUpload={() => {
              void cancelUploadHandler?.();
            }}
            onClose={clearUploadProgress}
          />

          {isRootEmpty ? (
            <div className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center">
              <p className="text-sm text-slate-300">This library root is empty.</p>
              <p className="mt-2 text-sm text-slate-500">
                Create your first folder or upload files to start.
              </p>
            </div>
          ) : (
            <LibraryTreeEntriesTable
              entries={rootEntries}
              selectedEntryIds={selectedEntryIds}
              useSelectionColumn={useSelectionColumn}
              headerGridClass={headerGridClass}
              rowGridClass={rowGridClass}
              isActionLocked={isActionLocked}
              cutClipboardKeys={cutClipboardKeys}
              onRowClick={handleRowClick}
              onRowDoubleClick={handleRowDoubleClick}
              onRowTouchStart={handleRowTouchStart}
              onClearLongPressTimer={clearLongPressTimer}
              onToggleEntrySelection={toggleEntrySelection}
              onOpenEntryInfo={openEntryInfoModal}
            />
          )}
        </CardContent>
      </Card>

      <LibraryEntryInfoModal
        entry={activeInfoEntry}
        onClose={() => setActiveInfoEntry(null)}
      />

      <CreateFolderModal
        open={isCreateFolderOpen}
        libraryId={effectiveLibraryId}
        defaultParentId={currentFolderId ?? undefined}
        currentFolderId={currentFolderId ?? undefined}
        onClose={() => setIsCreateFolderOpen(false)}
      />

      <UploadFileModal
        open={isUploadFileOpen}
        libraryId={effectiveLibraryId}
        folderId={currentFolderId ?? undefined}
        currentFolderId={currentFolderId ?? undefined}
        onClose={() => setIsUploadFileOpen(false)}
        onUploadProgressChange={setUploadProgress}
        onUploadCancelHandlerChange={(handler) => {
          if (!handler) {
            setCancelUploadHandler(null);
            return;
          }

          setCancelUploadHandler(() => handler);
        }}
      />

      <EditLibraryModal
        open={isEditLibraryOpen}
        library={currentLibraryItem}
        onClose={() => setIsEditLibraryOpen(false)}
      />

      <DeleteLibraryModal
        open={isDeleteLibraryOpen}
        library={currentLibraryItem}
        onClose={() => setIsDeleteLibraryOpen(false)}
      />
    </section>
  );
}
