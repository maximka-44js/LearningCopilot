import axios, {
    AxiosError,
    AxiosHeaders,
    type InternalAxiosRequestConfig,
} from "axios";

const DEFAULT_TIMEOUT_MS = 10_000;

type AuthAwareRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
  skipAuthRedirect?: boolean;
};

function resolveBaseURL() {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api";
}

function navigateToSignIn(url: string) {
  const targetWindow = window.top ?? window;
  targetWindow.location.assign(url);
}

export const axiosInstance = axios.create({
  baseURL: resolveBaseURL(),
  timeout: DEFAULT_TIMEOUT_MS,
  // We use Bearer auth; sending cross-origin cookies on localhost can overflow headers (431).
  withCredentials: false,
  headers: {
    "Content-Type": "application/json",
    Accept: "application/json",
  },
});

async function withAccessToken(config: InternalAxiosRequestConfig) {
  const isFormDataPayload =
    typeof FormData !== "undefined" && config.data instanceof FormData;

  if (isFormDataPayload) {
    const headers = AxiosHeaders.from(config.headers);
    headers.delete("Content-Type");
    headers.delete("content-type");
    config.headers = headers;
  }

  if (typeof window === "undefined") {
    return config;
  }

  const { getSession } = await import("next-auth/react");
  const session = await getSession();

  if (
    session?.error === "RefreshAccessTokenError" ||
    session?.error === "RefreshTokenExpired"
  ) {
    const signInUrl = new URL("/auth/sign-in", window.location.origin);
    signInUrl.searchParams.set(
      "callbackUrl",
      `${window.location.pathname}${window.location.search}` || "/"
    );
    navigateToSignIn(signInUrl.toString());
    return config;
  }

  const accessToken = session?.accessToken;

  if (!accessToken) {
    return config;
  }

  const headers = AxiosHeaders.from(config.headers);

  headers.set("Authorization", `Bearer ${accessToken}`);
  config.headers = headers;

  return config;
}

function handleUnauthorized(error: AxiosError) {
  if (typeof window === "undefined") {
    return Promise.reject(error);
  }

  const originalRequest = error.config as
    | AuthAwareRequestConfig
    | undefined;

  if (error.response?.status !== 401) {
    return Promise.reject(error);
  }

  if (originalRequest?.skipAuthRedirect) {
    return Promise.reject(error);
  }

  if (originalRequest?._retry) {
    return redirectToSignIn(error);
  }

  if (originalRequest) {
    originalRequest._retry = true;
  }

  return tryRefreshAndRetry(originalRequest, error);
}

async function tryRefreshAndRetry(
  originalRequest:
    | AuthAwareRequestConfig
    | undefined,
  error: AxiosError
) {
  const { getSession } = await import("next-auth/react");
  const session = await getSession();

  if (
    session?.error === "RefreshAccessTokenError" ||
    session?.error === "RefreshTokenExpired" ||
    !session?.accessToken
  ) {
    return redirectToSignIn(error);
  }

  if (!originalRequest) {
    return Promise.reject(error);
  }

  const headers = AxiosHeaders.from(originalRequest.headers);
  headers.set("Authorization", `Bearer ${session.accessToken}`);
  originalRequest.headers = headers;

  return axiosInstance.request(originalRequest);
}

function redirectToSignIn(error: AxiosError) {
  const currentPath = window.location.pathname;
  if (currentPath !== "/auth/sign-in") {
    const callbackUrl = `${currentPath}${window.location.search}`;
    const signInUrl = new URL("/auth/sign-in", window.location.origin);
    signInUrl.searchParams.set("callbackUrl", callbackUrl || "/");
    navigateToSignIn(signInUrl.toString());
  }

  return Promise.reject(error);
}

axiosInstance.interceptors.request.use(withAccessToken);
axiosInstance.interceptors.response.use((response) => response, handleUnauthorized);
