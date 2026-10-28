import { cn } from "@/lib/utils";

export const formatDate = (value?: string) => {
  if (!value) return "N/A";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "N/A";
  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "short",
    day: "2-digit",
  }).format(date);
};

export const formatPercent = (value?: number) => {
  if (value === null || value === undefined) return "N/A";
  const rounded = Math.round(value);
  return `${rounded}%`;
};

export const formatLabel = (value?: string) => {
  if (!value) return "N/A";

  const normalized = value
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();

  if (!normalized) return "N/A";

  return normalized
    .split(" ")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(" ");
};

export const formatReason = (value?: string) => {
  if (!value) return "N/A";
  const trimmed = value.trim();
  if (!trimmed) return "N/A";

  // Only transform enum-like reasons (snake_case / kebab-case),
  // keep normal sentences as-is.
  if (trimmed.includes("_") || trimmed.includes("-")) {
    return formatLabel(trimmed);
  }

  return trimmed;
};

export const badgeTone = (value?: string) => {
  switch (value) {
    case "critical":
      return "border-rose-500/40 text-rose-200 bg-rose-500/10";
    case "high":
      return "border-amber-400/40 text-amber-200 bg-amber-400/10";
    case "medium":
      return "border-sky-400/40 text-sky-200 bg-sky-400/10";
    case "low":
      return "border-emerald-400/40 text-emerald-200 bg-emerald-400/10";
    default:
      return "border-slate-700 text-slate-200 bg-slate-900/60";
  }
};

export const riskTone = (value?: string) => {
  switch (value) {
    case "critical":
      return "rose";
    case "high":
      return "amber";
    case "medium":
      return "sky";
    case "low":
      return "emerald";
    default:
      return "slate";
  }
};

export const statusDotClass = (value?: string) => {
  switch (value) {
    case "active":
      return "bg-emerald-400";
    case "paused":
      return "bg-amber-300";
    case "archived":
      return "bg-slate-500";
    default:
      return "bg-slate-400";
  }
};

export const taskStatusDotClass = (value?: string) => {
  switch (value) {
    case "learning":
      return "bg-sky-400";
    case "baseline_passed":
      return "bg-amber-300";
    case "at_risk":
      return "bg-rose-400";
    case "stable":
      return "bg-emerald-400";
    case "exam_ready":
      return "bg-emerald-300";
    case "archived":
      return "bg-slate-500";
    case "backlog":
    default:
      return "bg-slate-400";
  }
};

export const metricTone = (value?: number) => {
  if (value === null || value === undefined) return "slate";
  if (value >= 80) return "emerald";
  if (value >= 60) return "sky";
  if (value >= 40) return "amber";
  return "rose";
};

const toneClasses = (tone: string) => {
  switch (tone) {
    case "emerald":
      return "text-emerald-200 bg-emerald-500/10 border-emerald-500/30";
    case "sky":
      return "text-sky-200 bg-sky-500/10 border-sky-500/30";
    case "amber":
      return "text-amber-200 bg-amber-400/10 border-amber-400/30";
    case "rose":
      return "text-rose-200 bg-rose-500/10 border-rose-500/30";
    default:
      return "text-slate-200 bg-slate-900/60 border-slate-800/70";
  }
};

const highlightToneClass = (tone: string) => {
  switch (tone) {
    case "emerald":
      return "bg-emerald-500/20";
    case "sky":
      return "bg-sky-500/20";
    case "amber":
      return "bg-amber-400/20";
    case "rose":
      return "bg-rose-500/20";
    default:
      return "bg-slate-500/20";
  }
};

const ProgressBar = ({ value, tone }: { value?: number; tone: string }) => {
  const safeValue = value === null || value === undefined ? 0 : value;
  const clamped = Math.max(0, Math.min(100, safeValue));

  return (
    <div className="mt-2 h-2 w-full rounded-full bg-slate-900">
      <div
        className={cn("h-full rounded-full", {
          "bg-emerald-400": tone === "emerald",
          "bg-sky-400": tone === "sky",
          "bg-amber-300": tone === "amber",
          "bg-rose-400": tone === "rose",
          "bg-slate-600": tone === "slate",
        })}
        style={{ width: `${clamped}%` }}
      />
    </div>
  );
};

export const Metric = ({
  label,
  value,
  tone,
  progress,
  compactOnMobile = false,
  highlightOnMobile = false,
}: {
  label: string;
  value: string;
  tone: string;
  progress?: number;
  compactOnMobile?: boolean;
  highlightOnMobile?: boolean;
}) => {
  const clamped = progress === null || progress === undefined ? 0 : Math.max(0, Math.min(100, progress));

  return (
    <div className="rounded-lg border border-slate-800/70 bg-slate-950/40 px-3 py-2">
      <div
        className={cn("relative flex flex-col", {
          "flex-row items-baseline justify-between gap-2 sm:flex-col": compactOnMobile,
        })}
      >
        {highlightOnMobile && progress !== undefined ? (
          <span
            className={cn(
              "absolute inset-y-0 left-0 rounded-sm sm:hidden",
              highlightToneClass(tone)
            )}
            style={{ width: `${clamped}%` }}
          />
        ) : null}
        <p className="relative z-10 text-[0.78rem] uppercase tracking-wide text-slate-400">
          {label}
        </p>
        <p
          className={cn(
            "relative z-10 text-[1.05rem] font-semibold",
            toneClasses(tone),
            compactOnMobile ? "sm:mt-1" : "mt-1"
          )}
        >
          {value}
        </p>
      </div>
      {progress !== undefined ? (
        highlightOnMobile ? (
          <div className="hidden sm:block">
            <ProgressBar value={progress} tone={tone} />
          </div>
        ) : (
          <ProgressBar value={progress} tone={tone} />
        )
      ) : null}
    </div>
  );
};
