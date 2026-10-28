import "server-only";

import { redirect } from "next/navigation";
import type { Session } from "next-auth";

import { auth } from "@/auth";
import { buildSignInUrl } from "@/features/auth/lib/route-protection";

export async function requireUserSession(callbackUrl: string): Promise<Session> {
  const session = await auth();

  if (!session?.user || session.error) {
    redirect(buildSignInUrl(callbackUrl));
  }

  return session;
}