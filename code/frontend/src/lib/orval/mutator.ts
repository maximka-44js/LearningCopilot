import { AxiosHeaders, type AxiosRequestConfig } from "axios";

import { axiosInstance } from "@/lib/axios";

type OrvalRequestOptions = Omit<RequestInit, "headers" | "method"> & {
  headers?: HeadersInit;
  method?: string;
  params?: AxiosRequestConfig["params"];
  data?: AxiosRequestConfig["data"];
  timeout?: AxiosRequestConfig["timeout"];
  responseType?: AxiosRequestConfig["responseType"];
  withCredentials?: AxiosRequestConfig["withCredentials"];
  baseURL?: AxiosRequestConfig["baseURL"];
  skipAuthRedirect?: boolean;
};

function normalizeHeaders(headers?: HeadersInit): AxiosRequestConfig["headers"] {
  if (!headers) {
    return undefined;
  }

  if (headers instanceof Headers) {
    return AxiosHeaders.from(Object.fromEntries(headers.entries()));
  }

  if (Array.isArray(headers)) {
    return AxiosHeaders.from(Object.fromEntries(headers));
  }

  return AxiosHeaders.from(headers);
}

export function orvalAxios<T>(
  url: string,
  options: OrvalRequestOptions = {}
): Promise<T> {
  const normalizedHeaders = normalizeHeaders(options.headers);
  const body = options.body as AxiosRequestConfig["data"] | undefined;
  const resolvedData = options.data ?? body;
  const isFormDataPayload =
    typeof FormData !== "undefined" && resolvedData instanceof FormData;
  const requestHeaders = isFormDataPayload ? undefined : normalizedHeaders;

  const config: AxiosRequestConfig = {
    url,
    ...options,
    method: options.method as AxiosRequestConfig["method"],
    headers: requestHeaders,
    signal: options.signal ?? undefined,
    data: resolvedData,
  };

  return axiosInstance<T>(config).then(({ data }) => data);
}
