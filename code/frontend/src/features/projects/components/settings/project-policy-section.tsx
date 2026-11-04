import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { inputClassName } from "@/features/projects/components/settings/project-settings-utils";

const fieldLabels: Record<string, string> = {
  priority_high_weight: "Вес высокого приоритета",
  priority_low_weight: "Вес низкого приоритета",
  priority_normal_weight: "Вес обычного приоритета",
  urgency_24h_weight: "Вес срочности: 24 часа",
  urgency_72h_weight: "Вес срочности: 72 часа",
  urgency_7d_weight: "Вес срочности: 7 дней",
  gap_threshold: "Порог расхождения",
  penalty_multiplier: "Множитель штрафа",
  acceptable_mastery_threshold: "Допустимый порог освоения",
  calibration_gap_threshold: "Порог расхождения для калибровки",
  low_mastery_threshold: "Порог низкого освоения",
  max_recommendations: "Максимум рекомендаций",
  block_critical_mastery_below: "Контекст: освоение ниже критического уровня",
  block_critical_readiness_below: "Контекст: готовность ниже критического уровня",
  block_high_mastery_below: "Контекст: освоение ниже высокого уровня",
  block_high_readiness_below: "Контекст: готовность ниже высокого уровня",
  block_medium_mastery_below: "Контекст: освоение ниже среднего уровня",
  block_medium_readiness_below: "Контекст: готовность ниже среднего уровня",
  false_confidence_gap: "Расхождение ложной уверенности",
  task_critical_readiness_below: "Задача: готовность ниже критического уровня",
  task_high_readiness_below: "Задача: готовность ниже высокого уровня",
  task_medium_readiness_below: "Задача: готовность ниже среднего уровня",
};

type ProjectPolicySectionProps = {
  title: string;
  description: string;
  values: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onSave: () => void;
  saving: boolean;
  dirty: boolean;
};

export default function ProjectPolicySection({
  title,
  description,
  values,
  onChange,
  onSave,
  saving,
  dirty,
}: ProjectPolicySectionProps) {
  return (
    <Card>
      <form
        onSubmit={(event) => {
          event.preventDefault();
          onSave();
        }}
      >
        <CardHeader className="flex items-start justify-between gap-4">
          <div>
            <CardTitle className="text-base">{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          <Button type="submit" size="sm" disabled={saving || !dirty}>
            {saving ? "Сохранение..." : "Сохранить"}
          </Button>
        </CardHeader>
        <CardContent className="grid gap-3 sm:grid-cols-2">
          {Object.entries(values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {fieldLabels[key] ?? String(key).replaceAll("_", " ")}
              </span>
              <input
                type="number"
                value={value}
                onChange={(event) => onChange(key, event.target.value)}
                className={inputClassName}
              />
            </label>
          ))}
        </CardContent>
      </form>
    </Card>
  );
}
