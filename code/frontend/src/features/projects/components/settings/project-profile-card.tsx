import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { ProjectFormValues } from "@/features/projects/components/settings/project-settings-utils";
import { inputClassName } from "@/features/projects/components/settings/project-settings-utils";
import { ENTITY_TITLE_MAX_LENGTH } from "@/lib/validation-limits";

type ProjectProfileCardProps = {
  values: ProjectFormValues;
  onChange: (next: Partial<ProjectFormValues>) => void;
  onSave: () => void;
  saving: boolean;
  dirty: boolean;
  error?: string | null;
};

export default function ProjectProfileCard({
  values,
  onChange,
  onSave,
  saving,
  dirty,
  error,
}: ProjectProfileCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <CardTitle>Профиль проекта</CardTitle>
          <CardDescription>
            Название, приоритет, дедлайн и направление проекта.
          </CardDescription>
        </div>
        <Button size="sm" onClick={onSave} disabled={saving || !dirty}>
          {saving ? "Сохранение..." : "Сохранить"}
        </Button>
      </CardHeader>
      <CardContent className="grid gap-4">
        <div className="grid gap-2">
          <label className="text-sm text-slate-300">Название</label>
          <input
            value={values.name}
            onChange={(event) => onChange({ name: event.target.value })}
            className={inputClassName}
            maxLength={ENTITY_TITLE_MAX_LENGTH}
          />
        </div>
        <div className="grid gap-2">
          <label className="text-sm text-slate-300">Описание</label>
          <textarea
            value={values.description}
            onChange={(event) => onChange({ description: event.target.value })}
            className="min-h-24 rounded-md border border-slate-800 bg-slate-950/70 px-3 py-2 text-sm text-slate-100"
          />
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="grid gap-2">
            <label className="text-sm text-slate-300">Дедлайн</label>
            <input
              type="datetime-local"
              value={values.deadline}
              onFocus={(event) => event.currentTarget.showPicker?.()}
              onChange={(event) => onChange({ deadline: event.target.value })}
              className={inputClassName}
              style={{ colorScheme: "dark" }}
            />
          </div>
          <div className="grid gap-2">
            <label className="text-sm text-slate-300">Приоритет</label>
            <select
              value={values.priority}
              onChange={(event) =>
                onChange({
                  priority: event.target.value as ProjectFormValues["priority"],
                })
              }
              className={inputClassName}
            >
              <option value="high">Высокий</option>
              <option value="normal">Обычный</option>
              <option value="low">Низкий</option>
            </select>
          </div>
        </div>
        <div className="grid gap-2">
          <label className="text-sm text-slate-300">Направление</label>
          <input
            value={values.metadataTrack}
            onChange={(event) => onChange({ metadataTrack: event.target.value })}
            className={inputClassName}
            placeholder="например, биология"
          />
        </div>
        {error ? <p className="text-sm text-rose-200">{error}</p> : null}
      </CardContent>
    </Card>
  );
}
