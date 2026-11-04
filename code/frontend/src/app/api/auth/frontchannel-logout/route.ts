import { cookies } from "next/headers";

export const dynamic = "force-dynamic";

export async function GET() {
  const cookieStore = await cookies();
  const sessionCookiePrefixes = [
    "authjs.session-token",
    "__Secure-authjs.session-token",
    "next-auth.session-token",
    "__Secure-next-auth.session-token",
  ];

  for (const cookie of cookieStore.getAll()) {
    if (
      sessionCookiePrefixes.some((prefix) => cookie.name.startsWith(prefix))
    ) {
      cookieStore.set(cookie.name, "", {
        expires: new Date(0),
        httpOnly: true,
        path: "/",
        sameSite: "lax",
        secure: cookie.name.startsWith("__Secure-"),
      });
    }
  }

  return new Response(null, {
    status: 204,
    headers: {
      "Cache-Control": "no-store",
    },
  });
}
