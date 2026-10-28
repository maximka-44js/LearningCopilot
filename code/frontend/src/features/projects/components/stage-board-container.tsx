"use client";

import { useMemo, useState } from "react";
import { useMutation, useQueries, useQueryClient } from "@tanstack/react-query";
import { Swiper, SwiperSlide } from "swiper/react";
import { useRouter } from "next/navigation";
import Link from "next/link";

import "swiper/css";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Metric,
  badgeTone,
  formatDate,
  formatLabel,
  formatPercent,
  metricTone,
  riskTone,
  statusDotClass,
} from "@/features/projects/components/card-metrics";
import CreateTaskForm from "@/features/projects/components/create-task-form";
import ConfirmDeleteStageModal from "@/features/projects/components/confirm-delete-stage-modal";
import EditStageForm from "@/features/projects/components/edit-stage-form";
import TaskCard from "@/features/projects/components/task-card";
import { getOrvalStatus, unwrapOrvalData } from "@/features/projects/lib/orval";
import { useGetStageBoard, getGetStageBoardQueryKey } from "@/generated/api/projects-read-model/projects-read-model";
import { getListStagesByProjectQueryKey } from "@/generated/api/stages/stages";
import { getGetTaskByIdQueryOptions } from "@/generated/api/tasks/tasks";
import type { InternalProjectsControllerRestapiV1GetStageBoardResponse } from "@/generated/model/internalProjectsControllerRestapiV1GetStageBoardResponse";
import type { GithubComMaximka44jsLcInternalProjectsEntityTask as Task } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityTask";
import { orvalAxios } from "@/lib/orval/mutator";
import { cn } from "@/lib/utils";

const ValueRow = ({ label, value }: { label: string; value: string }) => {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:items-baseline sm:justify-between">
      <dt className="text-xs uppercase tracking-wide text-slate-500">
        {label}
      </dt>
      <dd className="text-sm text-slate-200">{value}</dd>
    </div>
  );
};

const countValues = (values: Record<string, string | undefined> | null) => {
  const result: Record<string, number> = {};
  if (!values) return result;
  for (const value of Object.values(values)) {
    if (!value) continue;
    result[value] = (result[value] ?? 0) + 1;
  }
  return result;
};

const countDeficits = (values: Record<string, string[] | undefined> | null) => {
  const result: Record<string, number> = {};
  if (!values) return result;
  for (const list of Object.values(values)) {
    if (!Array.isArray(list)) continue;
    for (const item of list) {
      if (!item) continue;
      result[item] = (result[item] ?? 0) + 1;
    }
  }
  return result;
};

const blockTypes = ["comprehension", "memorization", "practice"] as const;

type TaskStatusFilter =
  | "all"
  | "backlog"
  | "learning"
  | "baseline_passed"
  | "at_risk"
  | "stable"
  | "exam_ready"
  | "archived";
type TaskRiskFilter = "all" | "low" | "medium" | "high" | "critical";
type TaskPriorityFilter = "all" | "low" | "normal" | "high";
type StageAction = "activate" | "pause" | "archive";

const normalizeText = (value: string) => value.trim().toLowerCase();

export default function StageBoardContainer({
  projectId,
  stageId,
}: {
  projectId: string;
  stageId: string;
}) {
  const { data, isLoading, isError } = useGetStageBoard(projectId, stageId);
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const router = useRouter();
  const [taskSearch, setTaskSearch] = useState("");
  const [taskStatusFilter, setTaskStatusFilter] = useState<TaskStatusFilter>("all");
  const [taskRiskFilter, setTaskRiskFilter] = useState<TaskRiskFilter>("all");
  const [taskPriorityFilter, setTaskPriorityFilter] = useState<TaskPriorityFilter>("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [detailsCollapsed, setDetailsCollapsed] = useState(true);
  const [signalsCollapsed, setSignalsCollapsed] = useState(true);
  const [stageAction, setStageAction] = useState<StageAction | null>(null);
  const [deletePending, setDeletePending] = useState(false);

  const { mutateAsync: mutateStage, isPending: stageActionPending } = useMutation({
    mutationFn: async ({ action }: { action: StageAction }) => {
      return orvalAxios(`/projects/${projectId}/stages/${stageId}/${action}`, {
        method: "PATCH",
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getGetStageBoardQueryKey(projectId, stageId),
      });
      queryClient.invalidateQueries({
        queryKey: getListStagesByProjectQueryKey(projectId),
      });
    },
    onError: () => {
      toast({
        title: "Stage update failed",
        description: "Unable to update the stage status right now.",
        variant: "error",
      });
    },
    onSettled: () => {
      setStageAction(null);
    },
  });

  const handleStageAction = async (action: StageAction) => {
    try {
      setStageAction(action);
      await mutateStage({ action });
    } catch {
      return;
    }
  };

  const handleDeleteStage = async () => {
    if (!stageId || !projectId) return;
    try {
      setDeletePending(true);
      await orvalAxios(`/projects/${projectId}/stages/${stageId}`, {
        method: "DELETE",
      });
      toast({
        title: "Stage deleted",
        description: "The stage was removed.",
      });
      setDeleteOpen(false);
      queryClient.invalidateQueries({
        queryKey: getListStagesByProjectQueryKey(projectId),
      });
      router.push(`/projects/${projectId}/stages`);
    } catch {
      toast({
        title: "Stage not deleted",
        description: "Unable to delete the stage right now.",
        variant: "error",
      });
    } finally {
      setDeletePending(false);
    }
  };

  const board = unwrapOrvalData<InternalProjectsControllerRestapiV1GetStageBoardResponse>(data);
  const stage = board?.stage;
  const stageStatus = stage?.status;
  const tasks = Array.isArray(board?.tasks) ? board?.tasks : [];
  const blockDeficits = board?.blockDeficits ?? {};
  const nextActionPriority = board?.nextActionPriority ?? {};
  const deficitCounts = countDeficits(blockDeficits as Record<string, string[] | undefined>);
  const priorityCounts = countValues(nextActionPriority as Record<string, string | undefined>);

  const taskIds = useMemo(() => {
    return tasks.map((task) => task.id).filter((id): id is string => Boolean(id));
  }, [tasks]);

  const taskQueries = useQueries({
    queries: taskIds.map((taskId) => {
      const options = getGetTaskByIdQueryOptions(projectId, stageId, taskId);
      return {
        ...options,
        enabled: Boolean(projectId) && Boolean(stageId) && Boolean(taskId),
      };
    }),
  });

  const taskDetailsById = useMemo(() => {
    const map = new Map<string, Task>();
    taskQueries.forEach((query, index) => {
      const taskId = taskIds[index];
      const task = unwrapOrvalData<Task>(query.data);
      if (taskId && task) {
        map.set(taskId, task);
      }
    });
    return map;
  }, [taskIds, taskQueries]);

  const anyTaskLoading = taskQueries.some((query) => query.isLoading);
  const anyTaskError = taskQueries.some((query) => query.isError);

  const filteredTasks = useMemo(() => {
    const query = normalizeText(taskSearch);

    return tasks.filter((task) => {
      if (!task) return false;
      if (taskStatusFilter === "all" && task.status === "archived") return false;
      if (taskStatusFilter !== "all" && task.status !== taskStatusFilter) return false;
      if (taskRiskFilter !== "all" && task.risk !== taskRiskFilter) return false;
      if (taskPriorityFilter !== "all" && task.priority !== taskPriorityFilter) return false;

      if (!query) return true;
      const haystack = normalizeText(`${task.outcome ?? ""}`);
      return haystack.includes(query);
    });
  }, [tasks, taskSearch, taskStatusFilter, taskRiskFilter, taskPriorityFilter]);

  if (isLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading stage board...
      </div>
    );
  }

  if (isError || !board) {
    const status = getOrvalStatus(data);
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Unable to load the stage board right now.
        {status ? (
          <span className="mt-2 block text-xs text-slate-400">
            API returned status {status}.
          </span>
        ) : null}
      </div>
    );
  }

  return (
    <div className="space-y-8 overflow-hidden">
      <div className="space-y-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
              Stage
            </p>
            <h1 className="text-3xl font-semibold">
              {stage?.name || "Stage board"}
            </h1>
            <p className="max-w-2xl text-sm text-slate-400">
              {stage?.goal?.trim() ||
                "Task readiness, deficits, and next actions for this stage."}
            </p>
          </div>

          <div className="flex flex-wrap gap-2 text-sm">
            {stageStatus === "active" ? (
              <>
                <Button
                  variant="outline"
                  onClick={() => handleStageAction("pause")}
                  disabled={stageActionPending}
                >
                  <img
                    src="/icons/pause-button-svgrepo-com.svg"
                    alt=""
                    aria-hidden="true"
                    className="mr-2 h-4.5 w-4.5 opacity-70"
                  />
                  {stageAction === "pause" ? "Pausing..." : "Pause"}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => handleStageAction("archive")}
                  disabled={stageActionPending}
                >
                  <img
                    src="/icons/archive-down-svgrepo-com.svg"
                    alt=""
                    aria-hidden="true"
                    className="mr-2 h-4.5 w-4.5 opacity-70"
                  />
                  {stageAction === "archive" ? "Archiving..." : "Archive"}
                </Button>
              </>
            ) : (
              <Button
                variant="outline"
                onClick={() => handleStageAction("activate")}
                disabled={stageActionPending}
              >
                <img
                  src={
                    stageStatus === "archived"
                      ? "/icons/archive-up-svgrepo-com.svg"
                      : "/icons/play-button-svgrepo-com.svg"
                  }
                  alt=""
                  aria-hidden="true"
                  className="mr-2 h-4.5 w-4.5 opacity-70"
                />
                {stageAction === "activate"
                  ? "Activating..."
                  : stageStatus === "archived"
                    ? "Unarchive"
                    : "Activate"}
              </Button>
            )}
          </div>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <Metric
            label="Readiness"
            value={formatPercent(stage?.readiness)}
            tone={metricTone(stage?.readiness)}
            progress={stage?.readiness}
          />
          <Metric
            label="Coverage"
            value={formatPercent(stage?.coverage)}
            tone={metricTone(stage?.coverage)}
            progress={stage?.coverage}
          />
          <Metric
            label="Stage risk"
            value={formatLabel(stage?.risk)}
            tone={riskTone(stage?.risk)}
          />
          <Metric
            label="Deadline"
            value={formatDate(stage?.deadline_at)}
            tone="slate"
          />
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="lg:flex lg:flex-col overflow-hidden">
          <CardHeader className="space-y-2">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <button
                type="button"
                onClick={() => setDetailsCollapsed((value) => !value)}
                className="flex flex-1 items-start justify-between gap-4 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
              >
                <div className="space-y-1">
                  <CardTitle>Stage details</CardTitle>
                  <CardDescription>
                    Status, risk, and key dates returned by the stage board.
                  </CardDescription>
                </div>
                <span className="text-lg font-semibold text-slate-300 lg:hidden">
                  {detailsCollapsed ? "+" : "-"}
                </span>
              </button>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setEditOpen(true)}
                  disabled={!stage}
                  aria-label="Edit stage"
                >
                  <img
                    src="/icons/edit-svgrepo-com.svg"
                    alt=""
                    aria-hidden="true"
                    className="h-4 w-4 opacity-70"
                  />
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setDeleteOpen(true)}
                  size="sm"
                  disabled={!stage || deletePending}
                  aria-label="Delete stage"
                >
                  <img
                    src="/icons/delete-2-svgrepo-com.svg"
                    alt=""
                    aria-hidden="true"
                    className="h-4.5 w-4.5 opacity-70"
                  />
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent
            className={cn(
              "space-y-4",
              detailsCollapsed ? "hidden lg:block" : "block"
            )}
          >
            <dl className="space-y-3">
              <ValueRow label="Status" value={formatLabel(stage?.status)} />
              <ValueRow label="Risk" value={formatLabel(stage?.risk)} />
              <ValueRow
                label="Order"
                value={
                  typeof stage?.order === "number" ? `${stage.order}` : "N/A"
                }
              />
              <ValueRow label="Deadline" value={formatDate(stage?.deadline_at)} />
              <ValueRow label="Created" value={formatDate(stage?.created_at)} />
              <ValueRow label="Updated" value={formatDate(stage?.updated_at)} />
              <ValueRow label="Archived" value={formatDate(stage?.archived_at)} />
            </dl>

            {stage ? (
              <div className="flex flex-wrap gap-2 text-sm">
                <span
                  className={cn(
                    "inline-flex items-center gap-2 rounded-md border px-2 py-1 font-medium",
                    badgeTone(stage.status)
                  )}
                >
                  <span
                    className={cn(
                      "h-2 w-2 rounded-full",
                      statusDotClass(stage.status)
                    )}
                  />
                  Status: {formatLabel(stage.status)}
                </span>
                <span
                  className={cn(
                    "inline-flex items-center rounded-md border px-2 py-1 font-medium",
                    badgeTone(stage.risk)
                  )}
                >
                  Risk: {formatLabel(stage.risk)}
                </span>
              </div>
            ) : null}
          </CardContent>
        </Card>

        <Card className="lg:flex lg:flex-col overflow-hidden">
          <CardHeader className="space-y-2">
            <button
              type="button"
              onClick={() => setSignalsCollapsed((value) => !value)}
              className="flex w-full items-center justify-between gap-4 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
            >
              <div className="space-y-1">
                <CardTitle>Signals</CardTitle>
                <CardDescription>
                  Aggregated deficits and next-action priorities for stage tasks.
                </CardDescription>
              </div>
              <span className="text-lg font-semibold text-slate-300 lg:hidden">
                {signalsCollapsed ? "+" : "-"}
              </span>
            </button>
          </CardHeader>
          <CardContent
            className={cn(
              "space-y-4",
              signalsCollapsed ? "hidden lg:block" : "block"
            )}
          >
            <div className="space-y-2">
              <p className="text-xs uppercase tracking-wide text-slate-500">
                Block deficits
              </p>
              {anyTaskLoading ? (
                <p className="text-sm text-slate-400">Loading task signals...</p>
              ) : anyTaskError ? (
                <p className="text-sm text-rose-200">
                  Unable to load task signals right now.
                </p>
              ) : Object.keys(deficitCounts).length === 0 ? (
                <p className="text-sm text-slate-400">No deficits reported yet.</p>
              ) : (
                <div className="overflow-hidden">
                  <Swiper
                    spaceBetween={12}
                    slidesPerView={1}
                    className="touch-pan-x"
                  >
                  {taskIds.map((taskId, index) => {
                    const task = taskDetailsById.get(taskId) ?? tasks[index];
                    const deficits = Array.isArray(blockDeficits[taskId])
                      ? blockDeficits[taskId]
                      : [];
                    const blocks = Array.isArray(task?.blocks) ? task.blocks : [];

                    const deficitLabels = deficits.map((deficit) => {
                      const block = blocks.find((item) => item?.type === deficit);
                      const mastery = formatPercent(block?.mastery);
                      return `${formatLabel(deficit)} ${mastery}`;
                    });

                    return (
                      <SwiperSlide key={taskId} className="w-full!">
                        <div className="h-full rounded-lg border border-slate-800/70 bg-slate-950/40 p-3">
                          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
                            Task
                          </p>
                          <Link
                            href={`/projects/${projectId}/stages/${stageId}/tasks/${taskId}`}
                            className="mt-2 inline-flex items-center gap-1 text-sm font-semibold text-slate-100 transition hover:text-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
                          >
                            <span className="truncate">
                              {task?.outcome?.trim() || "Task"}
                            </span>
                            <img
                              src="/icons/external-link.svg"
                              alt=""
                              aria-hidden="true"
                              className="h-3.5 w-3.5 shrink-0 opacity-70"
                            />
                          </Link>
                          <div className="mt-3 flex flex-wrap gap-2">
                            {deficits.length === 0 ? (
                              <span className="rounded-md border border-slate-800/70 bg-slate-950/40 px-2 py-1 text-xs text-slate-200">
                                No deficits
                              </span>
                            ) : (
                              deficitLabels.map((label, labelIndex) => (
                                <span
                                  key={`${taskId}-${labelIndex}`}
                                  className="rounded-md border border-slate-800/70 bg-slate-950/40 px-2 py-1 text-xs text-slate-200"
                                >
                                  {label}
                                </span>
                              ))
                            )}
                          </div>
                          <div className="mt-4 space-y-2 text-xs text-slate-300">
                            {blockTypes.map((type) => {
                              const block = blocks.find((item) => item?.type === type);
                              return (
                                <div
                                  key={`${taskId}-${type}`}
                                  className="flex items-center justify-between"
                                >
                                  <span>{formatLabel(type)}</span>
                                  <span className="text-slate-400">
                                    {formatPercent(block?.mastery)}
                                  </span>
                                </div>
                              );
                            })}
                          </div>
                        </div>
                      </SwiperSlide>
                    );
                  })}
                  </Swiper>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <p className="text-xs uppercase tracking-wide text-slate-500">
                Next action priorities
              </p>
              {anyTaskLoading ? (
                <p className="text-sm text-slate-400">Loading task signals...</p>
              ) : anyTaskError ? (
                <p className="text-sm text-rose-200">
                  Unable to load task signals right now.
                </p>
              ) : Object.keys(priorityCounts).length === 0 ? (
                <p className="text-sm text-slate-400">
                  No next-action priorities yet.
                </p>
              ) : (
                <div className="overflow-hidden">
                  <Swiper
                    spaceBetween={12}
                    slidesPerView={1}
                    className="touch-pan-x"
                  >
                  {taskIds.map((taskId, index) => {
                    const task = taskDetailsById.get(taskId) ?? tasks[index];
                    const priority = formatLabel(nextActionPriority[taskId]);

                    return (
                      <SwiperSlide key={`${taskId}-priority`} className="w-full!">
                        <div className="h-full rounded-lg border border-slate-800/70 bg-slate-950/40 p-3">
                          <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
                            Task
                          </p>
                          <Link
                            href={`/projects/${projectId}/stages/${stageId}/tasks/${taskId}`}
                            className="mt-2 inline-flex items-center gap-1 text-sm font-semibold text-slate-100 transition hover:text-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
                          >
                            <span className="truncate">
                              {task?.outcome?.trim() || "Task"}
                            </span>
                            <img
                              src="/icons/external-link.svg"
                              alt=""
                              aria-hidden="true"
                              className="h-3.5 w-3.5 shrink-0 opacity-70"
                            />
                          </Link>
                          <div className="mt-3 flex flex-wrap gap-2">
                            <span className="rounded-md border border-slate-800/70 bg-slate-950/40 px-2 py-1 text-xs text-slate-200">
                              Priority: {priority}
                            </span>
                            <span
                              className={cn(
                                "rounded-md border px-2 py-1 text-xs",
                                badgeTone(task?.risk)
                              )}
                            >
                              Risk: {formatLabel(task?.risk)}
                            </span>
                          </div>
                          <div className="mt-4 grid gap-2 text-xs text-slate-300">
                            <div className="flex items-center justify-between">
                              <span>Readiness</span>
                              <span className="text-slate-400">
                                {formatPercent(task?.readiness)}
                              </span>
                            </div>
                            <div className="flex items-center justify-between">
                              <span>Coverage</span>
                              <span className="text-slate-400">
                                {formatPercent(task?.coverage)}
                              </span>
                            </div>
                            <div className="flex items-center justify-between">
                              <span>Deadline</span>
                              <span className="text-slate-400">
                                {formatDate(task?.deadline_at)}
                              </span>
                            </div>
                          </div>
                        </div>
                      </SwiperSlide>
                    );
                  })}
                  </Swiper>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <section className="space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <h2 className="text-xl font-semibold text-slate-100">Tasks</h2>
            <p className="text-sm text-slate-400">
              Task readiness and guidance for the selected stage.
            </p>
          </div>
          <Button variant="outline" onClick={() => setCreateOpen(true)}>
            New task
          </Button>
        </div>

        <div className="flex flex-col gap-3 rounded-xl border border-slate-800 bg-slate-950/40 p-4">
          <div className="flex flex-1 flex-col gap-3 lg:flex-row lg:items-center">
            <div className="flex-1">
              <label className="sr-only" htmlFor="task-search">
                Search tasks
              </label>
              <input
                id="task-search"
                value={taskSearch}
                onChange={(e) => setTaskSearch(e.target.value)}
                placeholder="Search tasks by outcome"
                className="h-10 w-full rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 placeholder:text-slate-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
              />
            </div>

            <div className="flex flex-wrap gap-2">
              <label className="sr-only" htmlFor="task-status">
                Filter by status
              </label>
              <select
                id="task-status"
                value={taskStatusFilter}
                onChange={(e) => setTaskStatusFilter(e.target.value as TaskStatusFilter)}
                className="h-10 rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
              >
                <option value="all">All statuses</option>
                <option value="backlog">Backlog</option>
                <option value="learning">Learning</option>
                <option value="baseline_passed">Baseline passed</option>
                <option value="at_risk">At risk</option>
                <option value="stable">Stable</option>
                <option value="exam_ready">Exam ready</option>
                <option value="archived">Archived</option>
              </select>

              <label className="sr-only" htmlFor="task-risk">
                Filter by risk
              </label>
              <select
                id="task-risk"
                value={taskRiskFilter}
                onChange={(e) => setTaskRiskFilter(e.target.value as TaskRiskFilter)}
                className="h-10 rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
              >
                <option value="all">All risks</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>

              <label className="sr-only" htmlFor="task-priority">
                Filter by priority
              </label>
              <select
                id="task-priority"
                value={taskPriorityFilter}
                onChange={(e) => setTaskPriorityFilter(e.target.value as TaskPriorityFilter)}
                className="h-10 rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
              >
                <option value="all">All priorities</option>
                <option value="low">Low</option>
                <option value="normal">Normal</option>
                <option value="high">High</option>
              </select>
            </div>
          </div>
        </div>

        <CreateTaskForm
          projectId={projectId}
          stageId={stageId}
          open={createOpen}
          onClose={() => setCreateOpen(false)}
        />

        <EditStageForm
          projectId={projectId}
          stageId={stageId}
          stage={stage}
          open={editOpen}
          onClose={() => setEditOpen(false)}
        />

        <ConfirmDeleteStageModal
          open={deleteOpen}
          stageName={stage?.name}
          onCancel={() => setDeleteOpen(false)}
          onConfirm={handleDeleteStage}
          isPending={deletePending}
        />

        <div className="space-y-5 sm:space-y-4">
          {tasks.length === 0 ? (
            <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
              No tasks in this stage yet.
            </div>
          ) : filteredTasks.length === 0 ? (
            <div className="rounded-xl border border-dashed border-slate-800 bg-slate-950/40 px-6 py-10 text-center text-sm text-slate-400">
              No tasks match the current filters.
            </div>
          ) : (
            filteredTasks.map((task, index) => {
              const taskId = task.id ?? "";
              const taskDetail = taskDetailsById.get(taskId) ?? task;
              const deficits = Array.isArray(blockDeficits[taskId])
                ? blockDeficits[taskId]
                : [];
              const blocks = Array.isArray(taskDetail.blocks)
                ? taskDetail.blocks
                : [];
              const deficitLabel = deficits.length
                ? deficits
                    .map((value) => {
                      const block = blocks.find((item) => item?.type === value);
                      const mastery = formatPercent(block?.mastery);
                      return `${formatLabel(value)} ${mastery}`;
                    })
                    .join(", ")
                : "None";
              const priorityLabel = formatLabel(nextActionPriority[taskId]);
              const description = `Deficits: ${deficitLabel} · Next action priority: ${priorityLabel}`;

              return (
                <TaskCard
                  key={task.id ?? `${task.outcome}-${index}`}
                  href={`/projects/${projectId}/stages/${stageId}/tasks/${taskId}`}
                  task={taskDetail}
                  description={description}
                />
              );
            })
          )}
        </div>
      </section>
    </div>
  );
}
