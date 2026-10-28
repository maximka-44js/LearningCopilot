"use client";

import * as React from "react";
import { createPortal } from "react-dom";

import { cn } from "@/lib/utils";

type ToastVariant = "default" | "error";

type ToastOptions = {
  title: string;
  description?: string;
  variant?: ToastVariant;
  duration?: number;
};

type ToastItem = ToastOptions & {
  id: string;
};

type ToastContextValue = {
  toast: (options: ToastOptions) => void;
};

const ToastContext = React.createContext<ToastContextValue | null>(null);

const toastTone = (variant: ToastVariant) => {
  if (variant === "error") {
    return "border-rose-500/40 bg-rose-500/10 text-rose-100";
  }

  return "border-slate-700 bg-slate-950/90 text-slate-100";
};

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<ToastItem[]>([]);
  const [isMounted, setIsMounted] = React.useState(false);

  React.useEffect(() => {
    setIsMounted(true);
  }, []);

  const removeToast = React.useCallback((id: string) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const toast = React.useCallback(
    ({ duration = 4000, variant = "default", ...options }: ToastOptions) => {
      const id = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}`;
      const nextToast: ToastItem = { id, variant, duration, ...options };

      setToasts((current) => [nextToast, ...current]);

      if (duration > 0) {
        window.setTimeout(() => removeToast(id), duration);
      }
    },
    [removeToast]
  );

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      {isMounted
        ? createPortal(
            <div className="fixed right-6 top-6 z-[9999] flex w-full max-w-sm flex-col gap-3">
              {toasts.map((item) => (
                <div
                  key={item.id}
                  className={cn(
                    "rounded-lg border px-4 py-3 shadow-lg backdrop-blur",
                    toastTone(item.variant ?? "default")
                  )}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="space-y-1">
                      <p className="text-sm font-semibold">{item.title}</p>
                      {item.description ? (
                        <p className="text-sm text-slate-300">
                          {item.description}
                        </p>
                      ) : null}
                    </div>
                    <button
                      type="button"
                      onClick={() => removeToast(item.id)}
                      className="text-xs text-slate-400 hover:text-slate-100"
                    >
                      Close
                    </button>
                  </div>
                </div>
              ))}
            </div>,
            document.body
          )
        : null}
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = React.useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within ToastProvider");
  }

  return context;
}
