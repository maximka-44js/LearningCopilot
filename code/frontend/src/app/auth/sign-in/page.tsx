import { redirect } from "next/navigation";

import {
  sanitizeCallbackUrl,
} from "@/features/auth/lib/route-protection";

type SignInPageProps = {
  searchParams?: Promise<{ callbackUrl?: string }>;
};

export default async function SignInPage({ searchParams }: SignInPageProps) {
  const params = await searchParams;
  const safeCallbackUrl = sanitizeCallbackUrl(params?.callbackUrl);
  const signInParams = new URLSearchParams({ callbackUrl: safeCallbackUrl });

  redirect(`/auth/sign-in/start?${signInParams.toString()}`);
}
