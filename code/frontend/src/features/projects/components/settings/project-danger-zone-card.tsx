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
          <CardTitle className="text-rose-100">Danger zone</CardTitle>
          <CardDescription>
            Deleting a project removes its policies and history permanently.
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
          {deleting ? "Deleting" : "Delete project"}
        </Button>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-rose-200">
          This action cannot be undone. Consider archiving if you might need it
          later.
        </p>
      </CardContent>
    </Card>
  );
}
