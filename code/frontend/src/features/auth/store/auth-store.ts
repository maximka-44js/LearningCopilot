import { create } from "zustand";

export type AuthUser = {
  name?: string | null;
  email?: string | null;
  image?: string | null;
};

type AuthState = {
  status: "loading" | "authenticated" | "unauthenticated";
  user: AuthUser | null;
  setSession: (status: AuthState["status"], user: AuthUser | null) => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  status: "loading",
  user: null,
  setSession: (status, user) => set({ status, user }),
}));
