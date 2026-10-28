import type { LibrariesListItem } from "@/features/storage/components/libraries-list";
import type { LibraryTreeContents } from "@/features/storage/lib/library-tree";
import type { RootTreeEntry } from "@/features/storage/lib/library-tree-view-types";
import type { GithubComMaximka44jsLcInternalStorageEntityLibrary as Library } from "@/generated/model/githubComMaximka44jsLcInternalStorageEntityLibrary";

export function normalizeTreeData(value: unknown): LibraryTreeContents | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  if (
    "data" in value &&
    value.data &&
    typeof value.data === "object" &&
    "library_id" in value.data
  ) {
    return value.data as LibraryTreeContents;
  }

  if ("library_id" in value) {
    return value as LibraryTreeContents;
  }

  return null;
}
export function normalizeLibraryData(value: unknown): Library | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  if (
    "data" in value &&
    value.data &&
    typeof value.data === "object" &&
    "id" in value.data
  ) {
    return value.data as Library;
  }

  if ("id" in value) {
    return value as Library;
  }

  return null;
}

export function normalizeFolderData(value: unknown): { id?: string; name?: string } | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  if (
    "data" in value &&
    value.data &&
    typeof value.data === "object" &&
    "id" in value.data
  ) {
    return value.data as { id?: string; name?: string };
  }

  if ("id" in value) {
    return value as { id?: string; name?: string };
  }

  return null;
}

export function buildRootEntries(tree: LibraryTreeContents | null): RootTreeEntry[] {
  const folders = Array.isArray(tree?.folders) ? tree.folders : [];
  const files = Array.isArray(tree?.files) ? tree.files : [];

  return [
    ...folders.map((folder) => ({
      id: folder.id ?? `folder-${folder.name}-${folder.parent_id}`,
      sourceFolderId: folder.id ?? undefined,
      sourceParentId: folder.parent_id ?? undefined,
      name: folder.name?.trim() || "Untitled folder",
      kind: "folder" as const,
      updatedAt: folder.updated_at,
    })),
    ...files.map((file) => ({
      id: file.id ?? `file-${file.name}-${file.folder_id}`,
      sourceFolderId: file.folder_id ?? undefined,
      name: file.name?.trim() || "Unnamed file",
      kind: "file" as const,
      updatedAt: file.updated_at,
      size: file.size,
      mimeType: file.mime_type,
    })),
  ].sort((left, right) => {
    if (left.kind !== right.kind) {
      return left.kind === "folder" ? -1 : 1;
    }

    return left.name.localeCompare(right.name, "en-US", {
      sensitivity: "base",
    });
  });
}

export function buildLibraryListItem(params: {
  effectiveLibraryId: string;
  library: Library | null;
}): LibrariesListItem {
  const { effectiveLibraryId, library } = params;

  return {
    id: effectiveLibraryId,
    title: library?.title?.trim() || "Untitled library",
    description: library?.description?.trim() || "No description",
    isPublic: Boolean(library?.public),
    ownerId: library?.user_id ?? "Unknown owner",
    createdAt: library?.created_at ?? null,
    updatedAt: library?.updated_at ?? null,
    accesses: Array.isArray(library?.accesses) ? library.accesses : [],
  };
}
