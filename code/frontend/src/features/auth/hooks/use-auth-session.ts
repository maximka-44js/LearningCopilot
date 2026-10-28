"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";

import { useAuthStore } from "@/features/auth/store/auth-store";

type SessionUser = {
  name?: string | null;
  email?: string | null;
  image?: string | null;
};

type SessionPayload = {
  user?: SessionUser;
  expires?: string;
};

async function fetchSession(): Promise<SessionPayload | null> {
  const response = await fetch("/api/auth/session");
  if (!response.ok) {
    return null;
  }
  return response.json();
}

export function useAuthSession() {
  const setSession = useAuthStore((state) => state.setSession);

  const query = useQuery({
    queryKey: ["auth", "session"],
    queryFn: fetchSession,
    refetchOnWindowFocus: false,
  });

  React.useEffect(() => {
    if (query.isLoading) {
      setSession("loading", null);
      return;
    }

    if (query.data?.user) {
      setSession("authenticated", query.data.user);
      return;
    }

    setSession("unauthenticated", null);
  }, [query.data, query.isLoading, setSession]);

  return query;
}
