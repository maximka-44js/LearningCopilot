"use client";

import { useMemo } from "react";
import { useQueries } from "@tanstack/react-query";

import { Card, CardContent } from "@/components/ui/card";
import TaskCard from "@/features/projects/components/task-card";
import { formatDate, formatReason } from "@/features/projects/components/card-metrics";
import { unwrapOrvalData } from "@/features/projects/lib/orval";
import { getGetTaskByIdQueryOptions } from "@/generated/api/tasks/tasks";
import type { GithubComMaximka44jsLcInternalProjectsEntityTask as Task } from "@/generated/model/githubComMaximka44jsLcInternalProjectsEntityTask";

export type CriticalItem = {
  deadlineAt?: string;
  reason?: string;
  stageID?: string;
  taskID?: string;
};

export default function CriticalToDeadlineTasks({
  projectId,
  items,
}: {
  projectId: string;
  items: CriticalItem[];
}) {
  const normalized = useMemo(() => {
    return items
      .map((item) => ({
        ...item,
        stageID: item.stageID?.trim(),
        taskID: item.taskID?.trim(),
      }))
      .filter((item) => Boolean(item.stageID) && Boolean(item.taskID));
  }, [items]);

  const queries = useQueries({
    queries: normalized.map((item) => {
      const options = getGetTaskByIdQueryOptions(
        projectId,
        item.stageID as string,
        item.taskID as string
      );
      return {
        ...options,
        enabled: Boolean(projectId) && Boolean(item.stageID) && Boolean(item.taskID),
      };
    }),
  });

  if (!normalized.length) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Критичные элементы не найдены.
      </div>
    );
  }

  const anyLoading = queries.some((q) => q.isLoading);
  if (anyLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Загрузка критичных задач...
      </div>
    );
  }

  const anyError = queries.some((q) => q.isError);
  if (anyError) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Не удалось загрузить критичные задачи.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {normalized.map((item, index) => {
        const query = queries[index];
        const task = unwrapOrvalData<Task>(query.data) as Task | null;

        if (!task) {
          return (
            <Card key={`${item.taskID ?? "task"}-${index}`}>
              <CardContent className="px-6 py-6 text-sm text-slate-200">
                Не удалось загрузить данные задачи {item.taskID}.
              </CardContent>
            </Card>
          );
        }

        const descriptionParts = [
          item.reason ? formatReason(item.reason) : null,
          item.deadlineAt ? `Дедлайн: ${formatDate(item.deadlineAt)}` : null,
        ].filter(Boolean);

        return (
          <TaskCard
            key={task.id ?? `${item.taskID}-${index}`}
            href={`/projects/${projectId}/stages/${item.stageID}/tasks/${item.taskID}`}
            task={task}
            description={descriptionParts.join(" · ")}
          />
        );
      })}
    </div>
  );
}
