import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { MouseEvent as ReactMouseEvent, RefObject } from "react";

import type { RootTreeEntry } from "@/features/storage/lib/library-tree-view-types";

type UseLibraryTreeSelectionParams = {
  rootEntries: RootTreeEntry[];
  rootEntriesById: Map<string, RootTreeEntry>;
  fileSystemCardRef: RefObject<HTMLDivElement | null>;
  onOpenFolder: (entry: RootTreeEntry) => void;
};

export function useLibraryTreeSelection(params: UseLibraryTreeSelectionParams) {
  const { fileSystemCardRef, onOpenFolder, rootEntries, rootEntriesById } = params;
  const [selectedEntryIds, setSelectedEntryIds] = useState<Record<string, boolean>>({});
  const [lastSelectedIndex, setLastSelectedIndex] = useState<number | null>(null);
  const [isTouchDevice, setIsTouchDevice] = useState(false);
  const [isMobileSelectionMode, setIsMobileSelectionMode] = useState(false);
  const longPressTimerRef = useRef<number | null>(null);
  const longPressTriggeredRef = useRef(false);

  const selectedEntries = useMemo(
    () =>
      Object.entries(selectedEntryIds)
        .filter(([, checked]) => checked)
        .map(([entryId]) => rootEntriesById.get(entryId))
        .filter((entry): entry is RootTreeEntry => Boolean(entry)),
    [rootEntriesById, selectedEntryIds]
  );

  const hasSelectedEntries = selectedEntries.length > 0;
  const selectedSingleEntry = selectedEntries.length === 1 ? selectedEntries[0] : null;

  const resetSelection = useCallback(() => {
    setSelectedEntryIds({});
    setLastSelectedIndex(null);
    setIsMobileSelectionMode(false);
  }, []);

  const clearLongPressTimer = useCallback(() => {
    if (longPressTimerRef.current !== null) {
      window.clearTimeout(longPressTimerRef.current);
      longPressTimerRef.current = null;
    }
  }, []);

  const applySingleSelection = useCallback((entryId: string, index: number) => {
    setSelectedEntryIds({ [entryId]: true });
    setLastSelectedIndex(index);
  }, []);

  const applyRangeSelection = useCallback(
    (index: number) => {
      const fallbackAnchorFromSelection = rootEntries.findIndex(
        (entry) => Boolean(selectedEntryIds[entry.id])
      );
      const anchorIndex =
        lastSelectedIndex !== null
          ? lastSelectedIndex
          : fallbackAnchorFromSelection >= 0
            ? fallbackAnchorFromSelection
            : null;

      if (anchorIndex === null) {
        const fallbackEntry = rootEntries[index];
        if (fallbackEntry) {
          applySingleSelection(fallbackEntry.id, index);
        }
        return;
      }

      const start = Math.min(anchorIndex, index);
      const end = Math.max(anchorIndex, index);
      const rangeIds = rootEntries.slice(start, end + 1).map((entry) => entry.id);

      setSelectedEntryIds(() => {
        const nextSelection: Record<string, boolean> = {};
        for (const id of rangeIds) {
          nextSelection[id] = true;
        }
        return nextSelection;
      });
    },
    [applySingleSelection, lastSelectedIndex, rootEntries, selectedEntryIds]
  );

  const toggleEntrySelection = useCallback((entryId: string, checked: boolean) => {
    setSelectedEntryIds((current) => ({
      ...current,
      [entryId]: checked,
    }));
  }, []);

  const handleRowClick = useCallback(
    (entry: RootTreeEntry, index: number, event: ReactMouseEvent<HTMLElement>) => {
      if (isTouchDevice) {
        if (longPressTriggeredRef.current) {
          longPressTriggeredRef.current = false;
          return;
        }

        if (!isMobileSelectionMode) {
          if (entry.kind === "folder") {
            onOpenFolder(entry);
            return;
          }

          applySingleSelection(entry.id, index);
        }
        return;
      }

      if (event.shiftKey) {
        applyRangeSelection(index);
        return;
      }

      if (event.metaKey || event.ctrlKey) {
        setSelectedEntryIds((current) => {
          const next = { ...current };
          if (next[entry.id]) {
            delete next[entry.id];
          } else {
            next[entry.id] = true;
          }
          return next;
        });
        setLastSelectedIndex(index);
        return;
      }

      applySingleSelection(entry.id, index);
    },
    [
      applyRangeSelection,
      applySingleSelection,
      isMobileSelectionMode,
      isTouchDevice,
      onOpenFolder,
    ]
  );

  const handleRowDoubleClick = useCallback(
    (entry: RootTreeEntry) => {
      if (isTouchDevice) {
        return;
      }

      if (entry.kind === "folder") {
        onOpenFolder(entry);
      }
    },
    [isTouchDevice, onOpenFolder]
  );

  const startLongPressSelection = useCallback(
    (entry: RootTreeEntry, index: number) => {
      if (!isTouchDevice || isMobileSelectionMode) {
        return;
      }

      clearLongPressTimer();
      longPressTriggeredRef.current = false;
      longPressTimerRef.current = window.setTimeout(() => {
        longPressTriggeredRef.current = true;
        setIsMobileSelectionMode(true);
        applySingleSelection(entry.id, index);
      }, 420);
    },
    [
      applySingleSelection,
      clearLongPressTimer,
      isMobileSelectionMode,
      isTouchDevice,
    ]
  );

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const pointerMedia = window.matchMedia("(pointer: coarse)");
    const updateTouchMode = () => {
      setIsTouchDevice(pointerMedia.matches);
    };

    updateTouchMode();
    pointerMedia.addEventListener("change", updateTouchMode);

    return () => {
      pointerMedia.removeEventListener("change", updateTouchMode);
    };
  }, []);

  useEffect(() => {
    return () => {
      clearLongPressTimer();
    };
  }, [clearLongPressTimer]);

  useEffect(() => {
    if (isTouchDevice || !hasSelectedEntries) {
      return;
    }

    const onDocumentMouseDown = (event: MouseEvent) => {
      const target = event.target as Node | null;
      if (!target) {
        return;
      }

      if (fileSystemCardRef.current?.contains(target)) {
        return;
      }

      resetSelection();
    };

    document.addEventListener("mousedown", onDocumentMouseDown);
    return () => {
      document.removeEventListener("mousedown", onDocumentMouseDown);
    };
  }, [fileSystemCardRef, hasSelectedEntries, isTouchDevice, resetSelection]);

  return {
    clearLongPressTimer,
    handleRowClick,
    handleRowDoubleClick,
    hasSelectedEntries,
    isMobileSelectionMode,
    isTouchDevice,
    resetSelection,
    selectedEntries,
    selectedEntryIds,
    selectedSingleEntry,
    startLongPressSelection,
    toggleEntrySelection,
  };
}
