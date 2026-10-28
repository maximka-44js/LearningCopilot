import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { inputClassName } from "@/features/projects/components/settings/project-settings-utils";

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
            {saving ? "Saving" : "Save"}
          </Button>
        </CardHeader>
        <CardContent className="grid gap-3 sm:grid-cols-2">
          {Object.entries(values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {String(key).replaceAll("_", " ")}
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
