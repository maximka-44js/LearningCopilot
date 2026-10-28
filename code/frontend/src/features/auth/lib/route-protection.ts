export const PROTECTED_ROUTE_PREFIXES = ["/projects", "/storage", "/access"] as const;

export function isProtectedPath(pathname: string) {
  return PROTECTED_ROUTE_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`)
  );
}

export function sanitizeCallbackUrl(callbackUrl?: string) {
  if (!callbackUrl?.startsWith("/")) {
    return "/";
  }

  if (callbackUrl.startsWith("//") || callbackUrl.startsWith("/\\")) {
    return "/";
  }

  return callbackUrl;
}

export function buildSignInUrl(callbackUrl?: string) {
  const safeCallbackUrl = sanitizeCallbackUrl(callbackUrl);
  const searchParams = new URLSearchParams({ callbackUrl: safeCallbackUrl });
  return `/auth/sign-in?${searchParams.toString()}`;
}