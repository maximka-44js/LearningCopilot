import { orvalAxios } from "@/lib/orval/mutator";

export async function deleteFolderById(id: string) {
  await orvalAxios<unknown>(`/storage/folders/${id}`, {
    method: "DELETE",
  });
}

export async function copyFileById(
  id: string,
  payload: { name?: string; target_folder_id?: string }
) {
  await orvalAxios<unknown>(`/storage/files/${id}/copy`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export async function copyFolderById(
  id: string,
  payload: { name?: string; target_parent_id?: string }
) {
  await orvalAxios<unknown>(`/storage/folders/${id}/copy`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}
