"use server";

import { auth, signIn, signOut } from "@/auth";
import { redirect } from "next/navigation";

export async function signInWithKeycloak() {
  await signIn("keycloak", { redirectTo: "/" });
}

export async function signOutWithKeycloak() {
  const issuer =
    process.env.KEYCLOAK_PUBLIC_ISSUER ?? process.env.KEYCLOAK_ISSUER;
  const appUrl = process.env.AUTH_URL ?? process.env.NEXTAUTH_URL;
  const clientId = process.env.KEYCLOAK_CLIENT_ID;

  if (!issuer || !appUrl) {
    await signOut({ redirectTo: "/" });
    return;
  }

  const session = await auth();
  const idToken = session?.idToken;

  const logoutUrl = new URL(
    "protocol/openid-connect/logout",
    issuer.endsWith("/") ? issuer : `${issuer}/`
  );

  if (idToken) {
    logoutUrl.searchParams.set("id_token_hint", idToken);
  }

  if (clientId) {
    logoutUrl.searchParams.set("client_id", clientId);
  }

  logoutUrl.searchParams.set("post_logout_redirect_uri", appUrl);

  await signOut({ redirect: false });
  redirect(logoutUrl.toString());
}

export async function openAccountConsole() {
  const issuer =
    process.env.KEYCLOAK_PUBLIC_ISSUER ?? process.env.KEYCLOAK_ISSUER;

  if (!issuer) {
    throw new Error("Missing KEYCLOAK_ISSUER");
  }

  redirect(`${issuer.replace(/\/$/, "")}/account`);
}
