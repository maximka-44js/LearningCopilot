import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type ProjectDangerZoneCardProps = {
  onDelete: () => void;
  deleting: boolean;
};

export default function ProjectDangerZoneCard({
  onDelete,
  deleting,
}: ProjectDangerZoneCardProps) {
  return (
    <Card className="border-rose-500/40">
      <CardHeader className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <CardTitle className="text-rose-100">Опасная зона</CardTitle>
          <CardDescription>
            Удаление проекта навсегда удалит его политики и историю.
          </CardDescription>
        </div>
        <Button
          variant="outline"
          onClick={onDelete}
          disabled={deleting}
          className={cn(
            "border-rose-500/60 text-rose-100 hover:bg-rose-500/10",
            deleting && "opacity-70"
          )}
        >
          <img
            src="/icons/delete-2-svgrepo-com.svg"
            alt=""
            aria-hidden="true"
            className="h-4 w-4 opacity-70"
          />
          {deleting ? "Удаление..." : "Удалить проект"}
        </Button>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-rose-200">
          Это действие нельзя отменить. Если проект может понадобиться позже,
          лучше его архивировать.
        </p>
      </CardContent>
    </Card>
  );
}
