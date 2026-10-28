import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type ProjectStatusCardProps = {
  status?: string;
  action: "activate" | "pause" | "archive" | null;
  onAction: (action: "activate" | "pause" | "archive") => void;
  activating: boolean;
  pausing: boolean;
  archiving: boolean;
  onDelete: () => void;
  deleting: boolean;
};

export default function ProjectStatusCard({
  status,
  action,
  onAction,
  activating,
  pausing,
  archiving,
  onDelete,
  deleting,
}: ProjectStatusCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <CardTitle>Project status</CardTitle>
          <CardDescription>
            Current status: {status ?? "unknown"}. Update the lifecycle state.
          </CardDescription>
        </div>
        <div className="flex flex-wrap gap-2">
          {status === "active" ? (
            <>
              <Button
                variant="outline"
                onClick={() => onAction("pause")}
                disabled={pausing || action !== null}
              >
                <img
                  src="/icons/pause-button-svgrepo-com.svg"
                  alt=""
                  aria-hidden="true"
                  className="h-4 w-4 opacity-70"
                />
                {action === "pause" ? "Pausing" : "Pause"}
              </Button>
              <Button
                variant="outline"
                onClick={() => onAction("archive")}
                disabled={archiving || action !== null}
              >
                <img
                  src="/icons/archive-down-svgrepo-com.svg"
                  alt=""
                  aria-hidden="true"
                  className="h-4 w-4 opacity-70"
                />
                {action === "archive" ? "Archiving" : "Archive"}
              </Button>
            </>
          ) : (
            <Button
              variant="outline"
              onClick={() => onAction("activate")}
              disabled={activating || action !== null}
            >
              <img
                src={
                  status === "archived"
                    ? "/icons/archive-up-svgrepo-com.svg"
                    : "/icons/play-button-svgrepo-com.svg"
                }
                alt=""
                aria-hidden="true"
                className="h-4 w-4 opacity-70"
              />
              {action === "activate"
                ? "Activating"
                : status === "archived"
                  ? "Unarchive"
                  : "Activate"}
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <div className="rounded-lg border border-slate-800/70 bg-slate-950/40 px-4 py-3 text-sm text-slate-300">
          Use pause to halt updates without archiving. Archive hides the project
          from active views but keeps data.
        </div>
        <div className="mt-4 rounded-lg border border-rose-500/40 bg-rose-950/10 px-4 py-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="space-y-1">
              <p className="text-sm font-semibold text-rose-100">Danger zone</p>
              <p className="text-sm text-rose-200">
                Deleting a project removes its policies and history permanently.
              </p>
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
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
