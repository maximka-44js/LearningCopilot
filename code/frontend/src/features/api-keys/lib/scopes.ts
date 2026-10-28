export const API_KEY_SCOPE_VALUES = [
  "storage:read",
  "storage:write",
  "projects:read",
  "projects:write",
] as const;

export type ApiKeyScopeValue = (typeof API_KEY_SCOPE_VALUES)[number];

export const API_KEY_SCOPE_GROUPS: ReadonlyArray<{
  domain: "storage" | "projects";
  scopes: ReadonlyArray<ApiKeyScopeValue>;
}> = [
  {
    domain: "storage",
    scopes: ["storage:read", "storage:write"],
  },
  {
    domain: "projects",
    scopes: ["projects:read", "projects:write"],
  },
];

export const API_KEY_SCOPE_PRESETS: ReadonlyArray<{
  id: "full" | "projects-readonly";
  label: string;
  scopes: ReadonlyArray<ApiKeyScopeValue>;
}> = [
  {
    id: "full",
    label: "All scopes",
    scopes: [...API_KEY_SCOPE_VALUES],
  },
  {
    id: "projects-readonly",
    label: "Projects readonly",
    scopes: ["projects:read"],
  },
];

export const API_KEY_SCOPE_RULES = {
  readMethods: "GET, HEAD, OPTIONS",
  writeMethods: "POST, PUT, PATCH, DELETE and other non-read methods",
} as const;
