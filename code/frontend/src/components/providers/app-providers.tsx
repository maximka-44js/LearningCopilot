"use client";

import * as React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useSession, SessionProvider } from "next-auth/react";
import { usePathname, useRouter } from "next/navigation";

import { ToastProvider } from "@/components/providers/toast-provider";

const PROTECTED_PATHS = ["/projects", "/storage", "/access"];

function SessionSynchronizer() {
  const { status } = useSession();
  const pathname = usePathname();
  const router = useRouter();
  const isProtectedPath = PROTECTED_PATHS.some(
    (path) => pathname === path || pathname.startsWith(`${path}/`)
  );

  React.useEffect(() => {
    if (status === "unauthenticated" && isProtectedPath) {
      router.refresh();
    }
  }, [isProtectedPath, router, status]);

  return null;
}

export function AppProviders({ children }: { children: React.ReactNode }) {
  const [queryClient] = React.useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60_000,
            retry: 1,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      <SessionProvider refetchInterval={5} refetchOnWindowFocus>
        <SessionSynchronizer />
        <ToastProvider>{children}</ToastProvider>
      </SessionProvider>
    </QueryClientProvider>
  );
}
