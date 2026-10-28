"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import ConfirmDeleteTaskModal from "@/features/projects/components/confirm-delete-task-modal";
import EditTaskForm from "@/features/projects/components/edit-task-form";
import EditTaskWeightsForm from "@/features/projects/components/edit-task-weights-form";
import RecordAssessmentForm from "@/features/projects/components/record-assessment-form";
import {
  badgeTone,
  formatDate,
  formatLabel,
  formatPercent,
  formatReason,
  Metric,
  metricTone,
  riskTone,
  statusDotClass,
  taskStatusDotClass,
} from "@/features/projects/components/card-metrics";
import { getOrvalStatus, unwrapOrvalData } from "@/features/projects/lib/orval";
import {
  getGetTaskReadinessViewQueryKey,
  useGetTaskReadinessView,
} from "@/generated/api/projects-read-model/projects-read-model";
import { getGetStageBoardQueryKey } from "@/generated/api/projects-read-model/projects-read-model";
import {
  getGetTaskByIdQueryKey,
  useGetTaskById,
  useTransitionTaskStatus,
} from "@/generated/api/tasks/tasks";
import type { InternalProjectsControllerRestapiV1GetTaskReadinessViewResponse } from "@/generated/model/internalProjectsControllerRestapiV1GetTaskReadinessViewResponse";
import type { GithubComMaximka44jsLcInternalProjectsEntityTask as Task } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityTask";
import type { V1TransitionTaskStatusNextStatus } from "@/generated/model/v1TransitionTaskStatusNextStatus";
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

const blockOrder = ["comprehension", "memorization", "practice"] as const;

const formatFractionPercent = (value?: number) => {
  if (value === null || value === undefined) return "N/A";
  if (value > 1) return formatPercent(value);
  return `${Math.round(value * 100)}%`;
};

const sortBlockKeys = (a: string, b: string) => {
  const aIndex = blockOrder.indexOf(a as (typeof blockOrder)[number]);
  const bIndex = blockOrder.indexOf(b as (typeof blockOrder)[number]);
  if (aIndex === -1 && bIndex === -1) return a.localeCompare(b);
  if (aIndex === -1) return 1;
  if (bIndex === -1) return -1;
  return aIndex - bIndex;
};

const taskStatusTransitions: Record<V1TransitionTaskStatusNextStatus, V1TransitionTaskStatusNextStatus[]> = {
  backlog: ["learning", "archived"],
  learning: ["baseline_passed", "at_risk", "archived"],
  baseline_passed: ["learning", "at_risk", "stable", "archived"],
  at_risk: ["learning", "stable", "archived"],
  stable: ["learning", "at_risk", "exam_ready", "archived"],
  exam_ready: ["at_risk", "archived"],
  archived: ["backlog"],
};

const isTaskStatus = (value?: string): value is V1TransitionTaskStatusNextStatus => {
  if (!value) return false;
  return value in taskStatusTransitions;
};

export default function TaskReadinessContainer({
  projectId,
  stageId,
  taskId,
}: {
  projectId: string;
  stageId: string;
  taskId: string;
}) {
  const readinessQuery = useGetTaskReadinessView(projectId, stageId, taskId);
  const taskQuery = useGetTaskById(projectId, stageId, taskId);
  const queryClient = useQueryClient();
  const router = useRouter();
  const { toast } = useToast();
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deletePending, setDeletePending] = useState(false);
  const [weightsOpen, setWeightsOpen] = useState(false);
  const [assessmentOpen, setAssessmentOpen] = useState(false);
  const [nextStatus, setNextStatus] = useState<V1TransitionTaskStatusNextStatus | "">("");

  const readiness = unwrapOrvalData<InternalProjectsControllerRestapiV1GetTaskReadinessViewResponse>(
    readinessQuery.data
  );
  const task = readiness?.task ?? unwrapOrvalData<Task>(taskQuery.data);

  const lastAssessments = readiness?.lastAssessments ?? {};
  const lastAssessmentEntries = useMemo(() => {
    return Object.entries(lastAssessments)
      .filter(([key]) => key)
      .sort(([a], [b]) => sortBlockKeys(a, b));
  }, [lastAssessments]);

  const allowedNextStatuses = useMemo(() => {
    const current = task?.status;
    return isTaskStatus(current) ? taskStatusTransitions[current] : [];
  }, [task?.status]);

  const { mutateAsync: transitionTaskStatus, isPending: statusPending } = useTransitionTaskStatus({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetTaskReadinessViewQueryKey(projectId, stageId, taskId),
        });
        queryClient.invalidateQueries({
          queryKey: getGetTaskByIdQueryKey(projectId, stageId, taskId),
        });
        queryClient.invalidateQueries({
          queryKey: getGetStageBoardQueryKey(projectId, stageId),
        });
        setNextStatus("");
      },
      onError: () => {
        toast({
          title: "Status not updated",
          description: "Unable to update the task status right now.",
          variant: "error",
        });
      },
    },
  });

  const handleStatusChange = async () => {
    if (!nextStatus) return;
    try {
      await transitionTaskStatus({
        projectId,
        stageId,
        taskId,
        data: {
          next_status: nextStatus,
        },
      });
      toast({
        title: "Status updated",
        description: `Task status set to ${formatLabel(nextStatus)}.`,
      });
    } catch {
      return;
    }
  };

  if (readinessQuery.isLoading || taskQuery.isLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading task readiness...
      </div>
    );
  }

  if (readinessQuery.isError || !readiness) {
    const status = getOrvalStatus(readinessQuery.data);
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Unable to load the task readiness right now.
        {status ? (
          <span className="mt-2 block text-xs text-slate-400">
            API returned status {status}.
          </span>
        ) : null}
      </div>
    );
  }

  const blockScoresNow = Array.isArray(readiness.blockScoresNow)
    ? readiness.blockScoresNow
    : [];
  const dropForecast = Array.isArray(readiness.dropForecast)
    ? readiness.dropForecast
    : [];
  const recommendations = Array.isArray(readiness.recommendations)
    ? readiness.recommendations
    : [];
  const taskBlocks = Array.isArray(task?.blocks) ? task.blocks : [];

  const handleDeleteTask = async () => {
    if (!projectId || !stageId || !taskId) return;
    try {
      setDeletePending(true);
      await orvalAxios(`/projects/${projectId}/stages/${stageId}/tasks/${taskId}`,
        {
          method: "DELETE",
        }
      );
      toast({
        title: "Task deleted",
        description: "The task was removed.",
      });
      setDeleteOpen(false);
      queryClient.invalidateQueries({
        queryKey: getGetStageBoardQueryKey(projectId, stageId),
      });
      queryClient.invalidateQueries({
        queryKey: getGetTaskReadinessViewQueryKey(projectId, stageId, taskId),
      });
      queryClient.invalidateQueries({
        queryKey: getGetTaskByIdQueryKey(projectId, stageId, taskId),
      });
      router.push(`/projects/${projectId}/stages/${stageId}`);
    } catch {
      toast({
        title: "Task not deleted",
        description: "Unable to delete the task right now.",
        variant: "error",
      });
    } finally {
      setDeletePending(false);
    }
  };

  return (
    <div className="space-y-8">
      <div className="space-y-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
              Task
            </p>
            <h1 className="text-3xl font-semibold">
              {task?.outcome?.trim() || "Task readiness"}
            </h1>
            <p className="max-w-2xl text-sm text-slate-400">
              Readiness, block health, and next actions for this task.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2 text-sm">
            <select
              value={nextStatus}
              onChange={(event) => {
                const value = event.target.value;
                setNextStatus(isTaskStatus(value) ? value : "");
              }}
              disabled={!task || allowedNextStatuses.length === 0 || statusPending}
              className="h-10 min-w-[220px] rounded-md border border-slate-800/70 bg-slate-950/60 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500/60"
            >
              <option value="">Change status</option>
              {allowedNextStatuses.map((status) => (
                <option key={status} value={status}>
                  {formatLabel(status)}
                </option>
              ))}
            </select>
            <Button
              variant="outline"
              onClick={handleStatusChange}
              disabled={!nextStatus || statusPending}
            >
              {statusPending ? "Updating..." : "Update status"}
            </Button>
          </div>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <Metric
            label="Readiness"
            value={formatPercent(task?.readiness)}
            tone={metricTone(task?.readiness)}
            progress={task?.readiness}
          />
          <Metric
            label="Coverage"
            value={formatPercent(task?.coverage)}
            tone={metricTone(task?.coverage)}
            progress={task?.coverage}
          />
          <Metric
            label="Task risk"
            value={formatLabel(task?.risk)}
            tone={riskTone(task?.risk)}
          />
          <Metric
            label="Deadline"
            value={formatDate(task?.deadline_at)}
            tone="slate"
          />
        </div>
      </div>

      <EditTaskForm
        projectId={projectId}
        stageId={stageId}
        taskId={taskId}
        task={task}
        open={editOpen}
        onClose={() => setEditOpen(false)}
      />

      <ConfirmDeleteTaskModal
        open={deleteOpen}
        taskOutcome={task?.outcome}
        onCancel={() => setDeleteOpen(false)}
        onConfirm={handleDeleteTask}
        isPending={deletePending}
      />

      <EditTaskWeightsForm
        projectId={projectId}
        stageId={stageId}
        taskId={taskId}
        task={task}
        open={weightsOpen}
        onClose={() => setWeightsOpen(false)}
      />

      <RecordAssessmentForm
        projectId={projectId}
        stageId={stageId}
        taskId={taskId}
        task={task}
        open={assessmentOpen}
        onClose={() => setAssessmentOpen(false)}
      />

      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="lg:flex lg:flex-col">
          <CardHeader className="space-y-2">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="space-y-1">
                <CardTitle>Task details</CardTitle>
                <CardDescription>
                  Status, policy, and scheduling details for the task.
                </CardDescription>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setEditOpen(true)}
                  disabled={!task}
                  aria-label="Edit task"
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
                  disabled={!task || deletePending}
                  aria-label="Delete task"
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
          <CardContent className="space-y-4">
            <dl className="space-y-3">
              <ValueRow label="Status" value={formatLabel(task?.status)} />
              <ValueRow label="Priority" value={formatLabel(task?.priority)} />
              <ValueRow label="Risk" value={formatLabel(task?.risk)} />
              <ValueRow
                label="Difficulty"
                value={task?.difficulty !== undefined ? `${task.difficulty}` : "N/A"}
              />
              <ValueRow
                label="Readiness"
                value={formatPercent(task?.readiness)}
              />
              <ValueRow
                label="Coverage"
                value={formatPercent(task?.coverage)}
              />
              <ValueRow
                label="Exam-ready mastery"
                value={
                  task?.policy?.exam_ready_min_mastery !== undefined
                    ? formatPercent(task.policy.exam_ready_min_mastery)
                    : "N/A"
                }
              />
              <ValueRow
                label="Require non-critical risk"
                value={
                  task?.policy?.require_non_critical_block_risk !== undefined
                    ? task.policy.require_non_critical_block_risk
                      ? "Yes"
                      : "No"
                    : "N/A"
                }
              />
              <ValueRow label="Created" value={formatDate(task?.created_at)} />
              <ValueRow label="Updated" value={formatDate(task?.updated_at)} />
              <ValueRow label="Archived" value={formatDate(task?.archived_at)} />
            </dl>

            {task ? (
              <div className="flex flex-wrap gap-2 text-sm">
                <span
                  className={cn(
                    "inline-flex items-center rounded-md border px-2 py-1 font-medium",
                    badgeTone(task.risk)
                  )}
                >
                  Risk: {formatLabel(task.risk)}
                </span>
                <span className="rounded-md border border-slate-800/70 bg-slate-900/60 px-2 py-1 text-slate-200">
                  <span className="inline-flex items-center gap-2">
                    <span
                      className={cn(
                        "h-2 w-2 rounded-full",
                        taskStatusDotClass(task.status)
                      )}
                    />
                    Status: {formatLabel(task.status)}
                  </span>
                </span>
              </div>
            ) : null}
          </CardContent>
        </Card>

        <Card className="lg:flex lg:flex-col">
          <CardHeader className="space-y-2">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="space-y-1">
                <CardTitle>Block scores now</CardTitle>
                <CardDescription>
                  Current mastery, freshness, and readiness score per block.
                </CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAssessmentOpen(true)}
                disabled={!task}
                aria-label="Record assessment"
              >
                <img
                  src="/icons/edit-svgrepo-com.svg"
                  alt=""
                  aria-hidden="true"
                  className="h-4 w-4 opacity-70"
                />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {blockScoresNow.length === 0 ? (
              <div className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-4 py-6 text-center text-sm text-slate-400">
                No block scores available yet.
              </div>
            ) : (
              <div className="space-y-4">
                {blockScoresNow.map((block, index) => (
                  <div
                    key={`${block.type ?? "block"}-${index}`}
                    className="rounded-lg border border-slate-800/70 bg-slate-950/40 p-4"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <h3 className="text-sm font-semibold text-slate-100">
                        {formatLabel(block.type)}
                      </h3>
                      <span
                        className={cn(
                          "inline-flex items-center rounded-md border px-2 py-1 text-xs font-medium",
                          badgeTone(block.risk)
                        )}
                      >
                        Risk: {formatLabel(block.risk)}
                      </span>
                    </div>
                    <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                      <Metric
                        label="Weight"
                        value={formatPercent(block.weight)}
                        tone="slate"
                      />
                      <Metric
                        label="Mastery"
                        value={formatPercent(block.mastery)}
                        tone={metricTone(block.mastery)}
                        progress={block.mastery}
                      />
                      <Metric
                        label="Freshness"
                        value={formatFractionPercent(block.freshness)}
                        tone="slate"
                      />
                      <Metric
                        label="Score now"
                        value={formatPercent(block.scoreNow)}
                        tone={metricTone(block.scoreNow)}
                        progress={block.scoreNow}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="lg:flex lg:flex-col">
          <CardHeader className="space-y-2">
            <CardTitle>Recommendations</CardTitle>
            <CardDescription>
              Targeted next actions based on the current deficits.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {recommendations.length === 0 ? (
              <div className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-4 py-6 text-center text-sm text-slate-400">
                No recommendations available.
              </div>
            ) : (
              <div className="space-y-4">
                {recommendations.map((rec, index) => (
                  <div
                    key={`${rec.task_id ?? taskId}-${rec.block_type ?? "rec"}-${index}`}
                    className="rounded-lg border border-slate-800/70 bg-slate-950/40 p-4"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <h3 className="text-sm font-semibold text-slate-100">
                        {formatLabel(rec.block_type)} focus
                      </h3>
                      <span
                        className={cn(
                          "inline-flex items-center rounded-md border px-2 py-1 text-xs font-medium",
                          badgeTone(rec.priority_hint)
                        )}
                      >
                        Priority: {formatLabel(rec.priority_hint)}
                      </span>
                    </div>
                    <p className="mt-2 text-sm text-slate-300">
                      {formatReason(rec.reason)}
                    </p>
                    {Array.isArray(rec.next_actions) && rec.next_actions.length > 0 ? (
                      <ul className="mt-3 space-y-1 text-sm text-slate-200">
                        {rec.next_actions.map((action, actionIndex) => (
                          <li key={`${action}-${actionIndex}`}>• {action}</li>
                        ))}
                      </ul>
                    ) : (
                      <p className="mt-3 text-sm text-slate-400">No actions listed.</p>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="lg:flex lg:flex-col">
          <CardHeader className="space-y-2">
            <CardTitle>Readiness drop forecast</CardTitle>
            <CardDescription>
              Expected readiness and risk over the next checkpoints.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {dropForecast.length === 0 ? (
              <div className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-4 py-6 text-center text-sm text-slate-400">
                No forecast points returned.
              </div>
            ) : (
              dropForecast.map((point, index) => (
                <div
                  key={`${point.at ?? "point"}-${index}`}
                  className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-slate-800/70 bg-slate-950/40 px-4 py-3 text-sm"
                >
                  <span className="text-slate-200">
                    {formatDate(point.at)}
                  </span>
                  <span className="text-slate-100">
                    {formatPercent(point.readiness)} readiness
                  </span>
                  <span
                    className={cn(
                      "inline-flex items-center rounded-md border px-2 py-1 text-xs font-medium",
                      badgeTone(point.risk)
                    )}
                  >
                    {formatLabel(point.risk)}
                  </span>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="lg:flex lg:flex-col">
          <CardHeader className="space-y-2">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="space-y-1">
                <CardTitle>Block configuration</CardTitle>
                <CardDescription>
                  Weights, decay profiles, and last confirmations by block.
                </CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setWeightsOpen(true)}
                disabled={!task}
                aria-label="Edit block weights"
              >
                <img
                  src="/icons/edit-svgrepo-com.svg"
                  alt=""
                  aria-hidden="true"
                  className="h-4 w-4 opacity-70"
                />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {taskBlocks.length === 0 ? (
              <div className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-4 py-6 text-center text-sm text-slate-400">
                No block configuration yet.
              </div>
            ) : (
              <div className="space-y-4">
                {taskBlocks.map((block, index) => (
                  <div
                    key={`${block.type ?? "config"}-${index}`}
                    className="rounded-lg border border-slate-800/70 bg-slate-950/40 p-4"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <h3 className="text-sm font-semibold text-slate-100">
                        {formatLabel(block.type)}
                      </h3>
                      <span
                        className={cn(
                          "inline-flex items-center rounded-md border px-2 py-1 text-xs font-medium",
                          badgeTone(block.risk_state)
                        )}
                      >
                        Risk: {formatLabel(block.risk_state)}
                      </span>
                    </div>
                    <div className="mt-3 grid gap-2 text-sm text-slate-300 sm:grid-cols-2">
                      <span>Weight: {formatPercent(block.weight)}</span>
                      <span>Mastery: {formatPercent(block.mastery)}</span>
                      <span>Decay profile: {block.decay_profile || "N/A"}</span>
                      <span>Last assessed: {formatDate(block.last_assessed_at)}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="lg:flex lg:flex-col">
          <CardHeader className="space-y-2">
            <CardTitle>Last assessments</CardTitle>
            <CardDescription>
              Most recent confirmations per block from the read-model.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {lastAssessmentEntries.length === 0 ? (
              <div className="rounded-lg border border-dashed border-slate-800 bg-slate-950/40 px-4 py-6 text-center text-sm text-slate-400">
                No assessments recorded yet.
              </div>
            ) : (
              lastAssessmentEntries.map(([key, value]) => (
                <div
                  key={key}
                  className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-slate-800/70 bg-slate-950/40 px-4 py-3 text-sm"
                >
                  <span className="text-slate-200">
                    {formatLabel(key)}
                  </span>
                  <span className="text-slate-100">
                    {formatDate(value)}
                  </span>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
