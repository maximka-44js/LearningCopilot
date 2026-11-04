"use server";

import { auth, signIn, signOut } from "@/auth";
import { redirect } from "next/navigation";

export async function signInWithKeycloak() {
  await signIn("keycloak", {
    redirectTo: "/",
    authorizationParams: { ui_locales: "ru" },
  });
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

  let idToken: string | undefined;
  try {
    const session = await auth();
    idToken = session?.idToken;
  } catch (error) {
    console.error("Failed to read the current auth session during logout:", error);
  }

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
  const appUrl = process.env.AUTH_URL ?? process.env.NEXTAUTH_URL;
  const clientId = process.env.KEYCLOAK_CLIENT_ID ?? "lc-frontend";

  if (!issuer || !appUrl) {
    throw new Error("Missing Keycloak account configuration");
  }

  const accountUrl = new URL(
    "account",
    issuer.endsWith("/") ? issuer : `${issuer}/`
  );
  accountUrl.searchParams.set("referrer", clientId);
  accountUrl.searchParams.set("referrer_uri", appUrl);
  accountUrl.searchParams.set("locale", "ru");

  redirect(accountUrl.toString());
}
