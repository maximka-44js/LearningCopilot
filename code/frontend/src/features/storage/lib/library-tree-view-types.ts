export type RootTreeEntry = {
  id: string;
  name: string;
  kind: "folder" | "file";
  sourceFolderId?: string;
  sourceParentId?: string;
  updatedAt?: string;
  size?: number;
  mimeType?: string;
};

export type FolderTrailItem = {
  id: string;
  name: string;
};

export type ClipboardEntry = {
  id: string;
  name: string;
  kind: "file" | "folder";
  sourceFolderId?: string;
  sourceParentId?: string;
};

export type StorageClipboard = {
  libraryId: string;
  mode: "copy" | "cut";
  entries: ClipboardEntry[];
};
