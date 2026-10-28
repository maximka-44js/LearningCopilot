export type OrvalWrappedResponse<TData = unknown> = {
  status: number;
  data: TData;
};

export function unwrapOrvalData<TData = unknown>(response: unknown): TData | null {
  if (!response) return null;

  if (Array.isArray(response)) {
    return response as TData;
  }

  if (typeof response !== "object") {
    return null;
  }

  if ("status" in response && "data" in response) {
    const wrapped = response as OrvalWrappedResponse<TData>;
    if (wrapped.status >= 200 && wrapped.status < 300) {
      return wrapped.data;
    }
    return null;
  }

  return response as TData;
}

export function getOrvalStatus(response: unknown): number | null {
  if (!response || typeof response !== "object" || Array.isArray(response)) return null;
  if ("status" in response) {
    const status = (response as { status?: unknown }).status;
    return typeof status === "number" ? status : null;
  }
  return null;
}
