import "./globals.css";

import { House } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";

import { auth } from "@/auth";
import { AppProviders } from "@/components/providers/app-providers";
import { cn } from "@/lib/utils";
import { Geist } from "next/font/google";

const geist = Geist({subsets:['latin'],variable:'--font-sans'});

export const metadata: Metadata = {
  title: "App",
  description: "Minimal Next.js skeleton",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const session = await auth();
  const isAuthenticated = Boolean(session?.user) && !session?.error;

  const guestLinks = [
    { href: "/auth/sign-in", label: "Log in" },
    { href: "/docs/", label: "Docs" },
  ];

  const authenticatedLinks = [
    { href: "/projects", label: "Projects" },
    { href: "/storage/libraries", label: "Storage" },
    { href: "/access/api-keys", label: "API keys" },
    { href: "/docs/", label: "Docs" },
  ];

  const links = isAuthenticated ? authenticatedLinks : guestLinks;

  return (
    <html lang="en" className={cn("dark", "font-sans", geist.variable)}>
      <body>
        <AppProviders>
          <header className="sticky top-0 z-40 border-b border-slate-800/70 bg-slate-950/85 backdrop-blur">
            <nav className="mx-auto flex w-full max-w-7xl items-center gap-3 px-4 py-3">
              <Link
                href="/"
                aria-label="Home"
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
            </nav>
          </header>

          {children}
        </AppProviders>
      </body>
    </html>
  );
}
