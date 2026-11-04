import "./globals.css";

import { House } from "lucide-react";
import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";

import { auth } from "@/auth";
import { AppProviders } from "@/components/providers/app-providers";
import { cn } from "@/lib/utils";
import { Geist } from "next/font/google";

const geist = Geist({subsets:['latin'],variable:'--font-sans'});

export const metadata: Metadata = {
  title: "Learning Copilot",
  description: "Учебный помощник для работы с материалами и целями",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const session = await auth();
  const isAuthenticated = Boolean(session?.user) && !session?.error;

  const guestLinks = [
    { href: "/auth/sign-in", label: "Войти" },
    { href: "/docs/", label: "Документация" },
  ];

  const authenticatedLinks = [
    { href: "/projects", label: "Проекты" },
    { href: "/storage/libraries", label: "Материалы" },
    { href: "/access/api-keys", label: "API-ключи" },
    { href: "/docs/", label: "Документация" },
  ];

  const links = isAuthenticated ? authenticatedLinks : guestLinks;
  const keycloakIssuer =
    process.env.KEYCLOAK_PUBLIC_ISSUER ?? process.env.KEYCLOAK_ISSUER;
  const appUrl = process.env.AUTH_URL ?? process.env.NEXTAUTH_URL;
  const keycloakAccountUrl =
    keycloakIssuer && appUrl
      ? (() => {
          const accountUrl = new URL(
            "account",
            keycloakIssuer.endsWith("/") ? keycloakIssuer : `${keycloakIssuer}/`
          );
          accountUrl.searchParams.set(
            "referrer",
            process.env.KEYCLOAK_CLIENT_ID ?? "lc-frontend"
          );
          accountUrl.searchParams.set("referrer_uri", appUrl);
          accountUrl.searchParams.set("locale", "ru");
          return accountUrl.toString();
        })()
      : undefined;

  return (
    <html lang="ru" className={cn("dark", "font-sans", geist.variable)}>
      <body>
        <AppProviders>
          <header className="sticky top-0 z-40 border-b border-slate-800/70 bg-slate-950/85 backdrop-blur">
            <nav className="mx-auto flex w-full max-w-7xl items-center gap-3 px-4 py-3">
              <Link
                href="/"
                aria-label="На главную"
                className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-slate-700/80 text-slate-100 transition hover:bg-slate-900"
              >
                <House className="h-4 w-4" />
              </Link>

              <div className="flex flex-wrap items-center gap-2 text-sm">
                {links.map((link) => (
                  <Link
                    key={link.href}
                    href={link.href}
                    className="rounded-md px-3 py-2 text-slate-200 transition hover:bg-slate-900 hover:text-white"
                  >
                    {link.label}
                  </Link>
                ))}
              </div>

              {isAuthenticated && keycloakAccountUrl ? (
                <Link
                  href={keycloakAccountUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label="Открыть профиль Keycloak"
                  title="Профиль Keycloak"
                  className="ml-auto inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md transition hover:bg-slate-900"
                >
                  <Image
                    src="/icons/account-svgrepo-com.svg"
                    alt=""
                    width={30}
                    height={30}
                  />
                </Link>
              ) : null}
            </nav>
          </header>

          {children}
        </AppProviders>
      </body>
    </html>
  );
}
