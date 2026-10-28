import { signIn } from "@/auth";
import { sanitizeCallbackUrl } from "@/features/auth/lib/route-protection";

export async function GET(request: Request) {
  const url = new URL(request.url);
  const callbackUrl = sanitizeCallbackUrl(
    url.searchParams.get("callbackUrl") ?? undefined
  );

  return await signIn("keycloak", { redirectTo: callbackUrl });
}
