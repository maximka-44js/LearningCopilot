import { cookies } from "next/headers";

export const dynamic = "force-dynamic";

export async function GET() {
  const cookieStore = await cookies();
  const sessionCookiePrefixes = [
    "authjs.session-token",
    "__Secure-authjs.session-token",
  ];

  for (const cookie of cookieStore.getAll()) {
    if (
      sessionCookiePrefixes.some((prefix) => cookie.name.startsWith(prefix))
    ) {
      cookieStore.delete(cookie.name);
    }
  }

  return new Response(null, {
    status: 204,
    headers: {
      "Cache-Control": "no-store",
    },
  });
}
