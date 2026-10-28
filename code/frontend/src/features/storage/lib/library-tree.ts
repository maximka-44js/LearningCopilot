import { useQuery, type UseQueryResult } from "@tanstack/react-query";

import { orvalAxios } from "@/lib/orval/mutator";
import type { GithubComMaximka44jsLcInternalStorageEntityFile as StorageFile } from "@/generated/model/githubComMaximka44jsLcInternalStorageEntityFile";
import type { GithubComMaximka44jsLcInternalStorageEntityFolder as StorageFolder } from "@/generated/model/githubComMaximka44jsLcInternalStorageEntityFolder";

export type LibraryTreeContents = {
  files?: StorageFile[];
  folders?: StorageFolder[];
  library_id?: string;
  parent_id?: string;
};

export type GetLibraryTreeResponse = {
  data: LibraryTreeContents;
  status: 200;
  headers: Headers;
};

type LibraryTreeRawResponse = LibraryTreeContents | GetLibraryTreeResponse;

const getLibraryTreeUrl = (id: string) => `/storage/libraries/${id}/tree`;

export function getLibraryTreeQueryKey(id: string) {
  return ["/storage/libraries", id, "tree"] as const;
}

export async function getLibraryTree(
  id: string,
  options?: RequestInit
): Promise<LibraryTreeRawResponse> {
  return orvalAxios<LibraryTreeRawResponse>(getLibraryTreeUrl(id), {
    ...options,
    method: "GET",
  });
}

export function useLibraryTree(
  id: string
): UseQueryResult<LibraryTreeRawResponse, Error> {
  return useQuery({
    queryKey: getLibraryTreeQueryKey(id),
    queryFn: ({ signal }) => getLibraryTree(id, { signal }),
    enabled: id.trim().length > 0,
  });
}
