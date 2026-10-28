"use client";

import { useQueryClient } from "@tanstack/react-query";
import { CalendarIcon } from "lucide-react";
import { useMemo, useRef, useState } from "react";
import { z } from "zod";

import { useToast } from "@/components/providers/toast-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import {
    getListApiKeysQueryKey,
    useCreateApiKey,
    useListApiKeys,
    useRevokeApiKey,
    useRotateApiKey,
} from "@/generated/api/access/access";
import type {
    InternalAccessControllerRestapiV1ApiKeyResponse,
    InternalAccessControllerRestapiV1CreateAPIKeyRequest,
} from "@/generated/model";

import {
    API_KEY_SCOPE_GROUPS,
    API_KEY_SCOPE_PRESETS,
    API_KEY_SCOPE_RULES,
    API_KEY_SCOPE_VALUES,
    type ApiKeyScopeValue,
} from "@/features/api-keys/lib/scopes";

type IssuedSecretState = {
  action: "created" | "rotated";
  keyName: string;
  secret: string;
};

type MaybeWrappedResponse<TData> = {
  status?: number;
  data?: TData;
};

const VALID_SCOPE_SET = new Set<string>(API_KEY_SCOPE_VALUES);

const createApiKeySchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Set a key name before creating an API key.")
    .max(120, "Name must be 120 characters or less."),
  expires_at: z
    .string()
    .trim()
    .optional()
    .refine(
      (value) =>
        !value ||
        (() => {
          const parsed = new Date(value);
          return !Number.isNaN(parsed.getTime());
        })(),
      "Expiration date is invalid."
    ),
  scopes: z
    .array(z.string())
    .min(1, "Scopes must include at least one valid domain:access value.")
    .refine(
      (scopes) => scopes.every((scope) => VALID_SCOPE_SET.has(scope)),
      "Only storage/projects with read/write access are allowed."
    ),
});

function formatTimestamp(value?: string, fallback = "Never") {
  if (!value) {
    return fallback;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return fallback;
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function normalizeListResponse(data: unknown): InternalAccessControllerRestapiV1ApiKeyResponse[] {
  if (Array.isArray(data)) {
    return data;
  }

  if (!data || typeof data !== "object") {
    return [];
  }

  if ("status" in data && "data" in data) {
    const wrapped = data as { status?: number; data?: unknown };
    if (wrapped.status === 200 && Array.isArray(wrapped.data)) {
      return wrapped.data as InternalAccessControllerRestapiV1ApiKeyResponse[];
    }
  }

  return [];
}

function normalizeMutationResponse<TData>(response: unknown): {
  status: number | null;
  data: TData | null;
} {
  if (!response || typeof response !== "object") {
    return { status: null, data: null };
  }

  if ("status" in response && "data" in response) {
    const wrapped = response as MaybeWrappedResponse<TData>;
    return {
      status: typeof wrapped.status === "number" ? wrapped.status : null,
      data: (wrapped.data ?? null) as TData | null,
    };
  }

  return {
    status: null,
    data: response as TData,
  };
}

function keyState(item: InternalAccessControllerRestapiV1ApiKeyResponse): {
  label: "active" | "expired" | "revoked";
  tone: string;
} {
  if (item.revoked_at) {
    return {
      label: "revoked",
      tone: "border-rose-500/40 bg-rose-500/10 text-rose-200",
    };
  }

  if (item.expires_at) {
    const expiresAt = Date.parse(item.expires_at);
    if (!Number.isNaN(expiresAt) && expiresAt <= Date.now()) {
      return {
        label: "expired",
        tone: "border-amber-500/40 bg-amber-500/10 text-amber-200",
      };
    }
  }

  return {
    label: "active",
    tone: "border-cyan-500/30 bg-cyan-500/10 text-cyan-200",
  };
}

function toIsoOrUndefined(value: string): string | undefined {
  if (!value.trim()) {
    return undefined;
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return undefined;
  }

  return parsed.toISOString();
}

export default function ApiKeysManager() {
  const [name, setName] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const [selectedScopes, setSelectedScopes] = useState<ApiKeyScopeValue[]>([...API_KEY_SCOPE_VALUES]);
  const [issuedSecret, setIssuedSecret] = useState<IssuedSecretState | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const expiresDateInputRef = useRef<HTMLInputElement | null>(null);

  const { toast } = useToast();
  const queryClient = useQueryClient();

  const listQuery = useListApiKeys();
  const { mutateAsync: createApiKey } = useCreateApiKey();
  const { mutateAsync: revokeApiKey, isPending: isRevoking } = useRevokeApiKey();
  const { mutateAsync: rotateApiKey, isPending: isRotating } = useRotateApiKey();

  const keys = useMemo(() => {
    const items = normalizeListResponse(listQuery.data);
    return [...items].sort((left, right) => {
      const leftTime = left.created_at ? Date.parse(left.created_at) : 0;
      const rightTime = right.created_at ? Date.parse(right.created_at) : 0;
      return rightTime - leftTime;
    });
  }, [listQuery.data]);

  const setPreset = (presetId: (typeof API_KEY_SCOPE_PRESETS)[number]["id"]) => {
    const preset = API_KEY_SCOPE_PRESETS.find((item) => item.id === presetId);
    if (!preset) {
      return;
    }
    setSelectedScopes([...preset.scopes]);
  };

  const toggleScope = (scope: ApiKeyScopeValue) => {
    setSelectedScopes((current) => {
      if (current.includes(scope)) {
        return current.filter((item) => item !== scope);
      }

      return [...current, scope].sort();
    });
  };

  const copySecret = async () => {
    if (!issuedSecret?.secret) {
      return;
    }

    try {
      await navigator.clipboard.writeText(issuedSecret.secret);
      toast({
        title: "Secret copied",
        description: "The new key secret is copied to clipboard.",
      });
    } catch {
      toast({
        title: "Copy failed",
        description: "Please copy the secret manually.",
        variant: "error",
      });
    }
  };

  const openExpiresDatePicker = () => {
    const input = expiresDateInputRef.current;
    if (!input) {
      return;
    }

    if (typeof input.showPicker === "function") {
      input.showPicker();
      return;
    }

    input.focus();
    input.click();
  };

  const handleCreate = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const parsed = createApiKeySchema.safeParse({
      name,
      expires_at: expiresAt,
      scopes: selectedScopes,
    });

    if (!parsed.success) {
      const message = parsed.error.issues[0]?.message ?? "Check form values.";
      toast({
        title: "Check form values",
        description: message,
        variant: "error",
      });
      return;
    }

    const normalizedName = parsed.data.name.trim();
    const normalizedScopes = parsed.data.scopes.filter((scope) => VALID_SCOPE_SET.has(scope));

    const payload: InternalAccessControllerRestapiV1CreateAPIKeyRequest = {
      name: normalizedName,
      scopes: normalizedScopes,
      expires_at: toIsoOrUndefined(parsed.data.expires_at ?? ""),
    };

    setIsSubmitting(true);
    let wasCreated = false;
    try {
      const response = await createApiKey({ data: payload });
      const normalized = normalizeMutationResponse<{
        api_key?: InternalAccessControllerRestapiV1ApiKeyResponse;
        secret?: string;
      }>(response);
      const createData = normalized.data;
      const isCreated =
        normalized.status === 201 ||
        Boolean(createData?.api_key?.id) ||
        Boolean(createData?.secret?.trim());

      if (!isCreated) {
        toast({
          title: "Key was not created",
          description: "Server rejected the create request.",
          variant: "error",
        });
        return;
      }

      const secret = createData?.secret?.trim();
      if (secret) {
        const createdName = createData?.api_key?.name?.trim() || normalizedName;
        setIssuedSecret({
          action: "created",
          keyName: createdName,
          secret,
        });
      } else {
        setIssuedSecret(null);
      }

      setName("");
      setExpiresAt("");
      setSelectedScopes([...API_KEY_SCOPE_VALUES]);
      wasCreated = true;

      toast({
        title: "API key created",
        description: secret
          ? "Store the secret now: it may not be visible again."
          : "Key created, but no secret was returned by API.",
      });
    } catch {
      toast({
        title: "Key creation failed",
        description: "Something went wrong while creating API key.",
        variant: "error",
      });
    } finally {
      if (wasCreated) {
        void queryClient.invalidateQueries({ queryKey: getListApiKeysQueryKey() });
      }
      setIsSubmitting(false);
    }
  };

  const handleRevoke = async (item: InternalAccessControllerRestapiV1ApiKeyResponse) => {
    if (!item.id) {
      return;
    }

    const shouldProceed = window.confirm(
      `Revoke key \"${item.name ?? item.prefix ?? item.id}\" immediately?`
    );
    if (!shouldProceed) {
      return;
    }

    try {
      const response = await revokeApiKey({ id: item.id });
      const normalized = normalizeMutationResponse<InternalAccessControllerRestapiV1ApiKeyResponse>(
        response
      );
      const isRevoked = normalized.status === 200 || Boolean(normalized.data?.revoked_at);

      if (!isRevoked) {
        toast({
          title: "Revoke failed",
          description: "Server did not accept revoke request.",
          variant: "error",
        });
        return;
      }

      await queryClient.invalidateQueries({ queryKey: getListApiKeysQueryKey() });
      toast({
        title: "Key revoked",
        description: "The API key is now inactive.",
      });
    } catch {
      toast({
        title: "Revoke failed",
        description: "Something went wrong while revoking key.",
        variant: "error",
      });
    }
  };

  const handleRotate = async (item: InternalAccessControllerRestapiV1ApiKeyResponse) => {
    if (!item.id) {
      return;
    }

    const suggestedName = item.name ?? "";
    const promptValue = window.prompt("Optional new key name for rotated key:", suggestedName);
    if (promptValue === null) {
      return;
    }

    const nextName = promptValue.trim();
    try {
      const response = await rotateApiKey({
        id: item.id,
        data: nextName ? { name: nextName } : undefined,
      });
      const normalized = normalizeMutationResponse<{
        api_key?: InternalAccessControllerRestapiV1ApiKeyResponse;
        secret?: string;
      }>(response);
      const rotateData = normalized.data;
      const isRotated =
        normalized.status === 200 ||
        Boolean(rotateData?.api_key?.id) ||
        Boolean(rotateData?.secret?.trim());

      if (!isRotated) {
        toast({
          title: "Rotate failed",
          description: "Server did not accept rotate request.",
          variant: "error",
        });
        return;
      }

      const secret = rotateData?.secret?.trim();
      if (secret) {
        setIssuedSecret({
          action: "rotated",
          keyName: nextName || item.name || item.prefix || item.id,
          secret,
        });
      } else {
        setIssuedSecret(null);
      }

      await queryClient.invalidateQueries({ queryKey: getListApiKeysQueryKey() });

      toast({
        title: "Key rotated",
        description: secret
          ? "New secret is available now. Save it before leaving the page."
          : "Key rotated successfully.",
      });
    } catch {
      toast({
        title: "Rotate failed",
        description: "Something went wrong while rotating key.",
        variant: "error",
      });
    }
  };

  if (listQuery.isLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading API keys...
      </div>
    );
  }

  if (listQuery.isError) {
    return (
      <div className="rounded-xl border border-rose-500/40 bg-rose-500/10 px-6 py-12 text-center text-sm text-rose-100">
        Unable to load API keys right now.
      </div>
    );
  }

  return (
    <section className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Create API key</CardTitle>
          <CardDescription>
            Scopes must use the format <code>domain:access</code>. Valid domains: storage,
            projects.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreate} className="grid gap-5">
            <div className="grid gap-2">
              <label htmlFor="api-key-name" className="text-sm text-slate-300">
                Name
              </label>
              <input
                id="api-key-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                maxLength={120}
                required
                placeholder="Frontend generated key"
                className="h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
              />
            </div>

            <div className="grid gap-2">
              <label htmlFor="api-key-expires" className="text-sm text-slate-300">
                Expires at (optional)
              </label>
              <div className="flex items-center gap-2">
                <input
                  ref={expiresDateInputRef}
                  id="api-key-expires"
                  value={expiresAt}
                  onChange={(event) => setExpiresAt(event.target.value)}
                  type="date"
                  className="h-10 flex-1 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={openExpiresDatePicker}
                  aria-label="Open calendar"
                >
                  <CalendarIcon className="size-4" />
                </Button>
              </div>
              <p className="text-xs text-slate-500">Leave empty for a non-expiring key.</p>
            </div>

            <div className="grid gap-3">
              <div className="flex flex-wrap items-center gap-2">
                {API_KEY_SCOPE_PRESETS.map((preset) => (
                  <Button
                    key={preset.id}
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setPreset(preset.id)}
                  >
                    {preset.label}
                  </Button>
                ))}
              </div>

              <div className="rounded-lg border border-slate-800/70 bg-slate-900/40 p-3 text-xs text-slate-400">
                <p>
                  <span className="font-medium text-slate-200">read</span> is required for methods:
                  {" "}
                  {API_KEY_SCOPE_RULES.readMethods}
                </p>
                <p>
                  <span className="font-medium text-slate-200">write</span> is required for methods:
                  {" "}
                  {API_KEY_SCOPE_RULES.writeMethods}
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-3">
                {API_KEY_SCOPE_GROUPS.map((group) => (
                  <div key={group.domain} className="rounded-lg border border-slate-800/80 p-3">
                    <p className="text-xs uppercase tracking-[0.16em] text-slate-500">{group.domain}</p>
                    <div className="mt-3 grid gap-2">
                      {group.scopes.map((scope) => (
                        <label key={scope} className="flex items-center gap-2 text-sm text-slate-200">
                          <input
                            type="checkbox"
                            checked={selectedScopes.includes(scope)}
                            onChange={() => toggleScope(scope)}
                            className="h-4 w-4 rounded border-slate-700 bg-slate-900"
                          />
                          <span>{scope}</span>
                        </label>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="flex items-center justify-end">
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? "Creating..." : "Create key"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Existing keys</CardTitle>
          <CardDescription>Revoke compromised keys or rotate secrets with optional rename.</CardDescription>
        </CardHeader>
        <CardContent>
          {keys.length === 0 ? (
            <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
              No API keys created yet.
            </div>
          ) : (
            <div className="grid gap-3">
              {keys.map((item) => {
                const state = keyState(item);
                const isRevoked = state.label === "revoked";

                return (
                  <div
                    key={item.id ?? `${item.prefix ?? "key"}-${item.created_at ?? "unknown"}`}
                    className="rounded-lg border border-slate-800/80 bg-slate-900/40 p-4"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div className="space-y-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <p className="text-sm font-medium text-slate-100">
                            {item.name || item.prefix || "Unnamed key"}
                          </p>
                          <span
                            className={`rounded-full border px-2 py-0.5 text-[0.7rem] uppercase tracking-[0.16em] ${state.tone}`}
                          >
                            {state.label}
                          </span>
                        </div>
                        <p className="text-xs text-slate-400">Prefix: {item.prefix || "unknown"}</p>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => void handleRotate(item)}
                          disabled={isRotating || isRevoked}
                        >
                          Rotate
                        </Button>
                        <Button
                          type="button"
                          variant="destructive"
                          size="sm"
                          onClick={() => void handleRevoke(item)}
                          disabled={isRevoking || isRevoked}
                        >
                          Revoke
                        </Button>
                      </div>
                    </div>

                    <div className="mt-3 grid gap-3 text-xs text-slate-400 md:grid-cols-3">
                      <p>Created: {formatTimestamp(item.created_at, "Unknown")}</p>
                      <p>Last used: {formatTimestamp(item.last_used_at)}</p>
                      <p>Expires: {formatTimestamp(item.expires_at)}</p>
                    </div>

                    <div className="mt-3 flex flex-wrap gap-2">
                      {(item.scopes ?? []).map((scope) => (
                        <span
                          key={`${item.id ?? item.prefix}-${scope}`}
                          className="rounded-full border border-slate-700 bg-slate-950/60 px-2 py-0.5 text-[0.72rem] text-slate-200"
                        >
                          {scope}
                        </span>
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={Boolean(issuedSecret)} onOpenChange={(open) => (!open ? setIssuedSecret(null) : null)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Save secret now</DialogTitle>
            <DialogDescription>
              {issuedSecret
                ? `Key ${issuedSecret.action === "created" ? "created" : "rotated"} for ${issuedSecret.keyName}. Secret may be shown only once.`
                : ""}
            </DialogDescription>
          </DialogHeader>

          {issuedSecret ? (
            <pre className="max-h-64 overflow-auto rounded-md border border-amber-300/30 bg-slate-950/80 px-3 py-2 text-xs text-slate-100">
              {issuedSecret.secret}
            </pre>
          ) : null}

          <DialogFooter className="bg-transparent p-0 sm:justify-end">
            <Button type="button" variant="outline" onClick={copySecret}>
              Copy secret
            </Button>
            <Button type="button" onClick={() => setIssuedSecret(null)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  );
}
