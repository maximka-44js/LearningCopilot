"use client";

import { useMemo, useState } from "react";

import Link from "next/link";

import CreateProjectForm from "@/features/projects/components/create-project-form";
import ProjectCard, { type ProjectCardData } from "@/features/projects/components/project-card";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type SortKey = "readiness" | "coverage" | "deadline" | "name" | "risk";

export type ProjectsListItem = {
  project: ProjectCardData;
  href: string;
};

type ProjectsListProps = {
  items: ProjectsListItem[];
  className?: string;
};

const sortLabels: Record<SortKey, string> = {
  readiness: "Readiness",
  coverage: "Coverage",
  deadline: "Deadline",
  name: "Name",
  risk: "Risk",
};

const riskRank: Record<string, number> = {
  critical: 4,
  high: 3,
  medium: 2,
  low: 1,
};

export default function ProjectsList({ items, className }: ProjectsListProps) {
  const [statusFilter, setStatusFilter] = useState("all");
  const [priorityFilter, setPriorityFilter] = useState("all");
  const [riskFilter, setRiskFilter] = useState("all");
  const [trackFilter, setTrackFilter] = useState("all");
  const [sortKey, setSortKey] = useState<SortKey>("readiness");
  const [sortDir, setSortDir] = useState("desc");
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const trackOptions = useMemo(() => {
    const tracks = new Set<string>();
    items.forEach(({ project }) => {
      const track = project.metadata?.track;
      if (track) {
        tracks.add(track);
      }
    });

    return Array.from(tracks).sort((a, b) => a.localeCompare(b));
  }, [items]);

  const visibleProjects = useMemo(() => {
    const filtered = items.filter(({ project }) => {
      const statusOk = statusFilter === "all" || project.status === statusFilter;
      const priorityOk =
        priorityFilter === "all" || project.priority === priorityFilter;
      const riskOk = riskFilter === "all" || project.risk === riskFilter;
      const trackOk =
        trackFilter === "all" || project.metadata?.track === trackFilter;

      return statusOk && priorityOk && riskOk && trackOk;
    });

    const sorted = [...filtered].sort((a, b) => {
      const direction = sortDir === "asc" ? 1 : -1;
      const left = a.project;
      const right = b.project;

      switch (sortKey) {
        case "name":
          return direction * (left.name || "").localeCompare(right.name || "");
        case "deadline": {
          const aValue = left.deadline_at ? Date.parse(left.deadline_at) : 0;
          const bValue = right.deadline_at ? Date.parse(right.deadline_at) : 0;
          return direction * (aValue - bValue);
        }
        case "coverage": {
          const aValue = left.coverage ?? -1;
          const bValue = right.coverage ?? -1;
          return direction * (aValue - bValue);
        }
        case "risk": {
          const aValue = left.risk ? riskRank[left.risk] ?? 0 : 0;
          const bValue = right.risk ? riskRank[right.risk] ?? 0 : 0;
          return direction * (aValue - bValue);
        }
        case "readiness":
        default: {
          const aValue = left.readiness ?? -1;
          const bValue = right.readiness ?? -1;
          return direction * (aValue - bValue);
        }
      }
    });

    return sorted;
  }, [
    items,
    priorityFilter,
    riskFilter,
    sortDir,
    sortKey,
    statusFilter,
    trackFilter,
  ]);

  return (
    <section className={cn("space-y-6", className)}>
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-2">
          <label className="text-[0.8rem] uppercase tracking-[0.2em] text-slate-400">
            Status
          </label>
          <select
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value)}
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] text-slate-100"
          >
            <option value="all">All</option>
            <option value="active">Active</option>
            <option value="paused">Paused</option>
            <option value="archived">Archived</option>
          </select>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-[0.8rem] uppercase tracking-[0.2em] text-slate-400">
            Priority
          </label>
          <select
            value={priorityFilter}
            onChange={(event) => setPriorityFilter(event.target.value)}
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] text-slate-100"
          >
            <option value="all">All</option>
            <option value="high">High</option>
            <option value="normal">Normal</option>
            <option value="low">Low</option>
          </select>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-[0.8rem] uppercase tracking-[0.2em] text-slate-400">
            Risk
          </label>
          <select
            value={riskFilter}
            onChange={(event) => setRiskFilter(event.target.value)}
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] text-slate-100"
          >
            <option value="all">All</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-[0.8rem] uppercase tracking-[0.2em] text-slate-400">
            Track
          </label>
          <select
            value={trackFilter}
            onChange={(event) => setTrackFilter(event.target.value)}
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] text-slate-100"
          >
            <option value="all">All</option>
            {trackOptions.map((track) => (
              <option key={track} value={track}>
                {track}
              </option>
            ))}
          </select>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-[0.8rem] uppercase tracking-[0.2em] text-slate-400">
            Sort
          </label>
          <select
            value={sortKey}
            onChange={(event) => setSortKey(event.target.value as SortKey)}
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] text-slate-100"
          >
            {Object.entries(sortLabels).map(([key, label]) => (
              <option key={key} value={key}>
                {label}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={() => setSortDir(sortDir === "asc" ? "desc" : "asc")}
            className="rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-[0.95rem] text-slate-100"
          >
            {sortDir === "asc" ? "Asc" : "Desc"}
          </button>
        </div>
        <div className="flex items-center gap-2">
          <Button type="button" size="sm" onClick={() => setIsCreateOpen(true)}>
            New project
          </Button>
        </div>
        <div className="flex items-center gap-2">
          <Link
            href="/projects/policies"
            className={cn(buttonVariants({ size: "sm", variant: "outline" }))}
          >
            Policy settings
          </Link>
        </div>
      </div>

      <div className="grid gap-6">
        {visibleProjects.length === 0 ? (
          <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
            No projects match the selected filters.
          </div>
        ) : (
          visibleProjects.map((item, index) => (
            <ProjectCard
              key={`${item.project.name}-${index}`}
              href={item.href}
              project={item.project}
            />
          ))
        )}
      </div>
      <CreateProjectForm
        open={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
      />
    </section>
  );
}
