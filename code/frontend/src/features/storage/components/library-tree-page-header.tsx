import Link from "next/link";

type LibraryTreePageHeaderProps = {
  title: string;
  libraryId: string;
  folderLabel: string;
};

export default function LibraryTreePageHeader({
  title,
  libraryId,
  folderLabel,
}: LibraryTreePageHeaderProps) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3">
      <div className="space-y-2">
        <h1 className="break-all text-3xl font-semibold text-slate-100">{title}</h1>
        <p className="break-all text-sm text-slate-400">Library ID: {libraryId}</p>
        <p className="text-sm text-slate-500">{folderLabel}</p>
      </div>
      <Link
        href="/storage/libraries"
        className="inline-flex h-9 items-center rounded-md border border-slate-700 px-3 text-sm font-medium text-slate-100 transition hover:bg-slate-900/40"
      >
        Back to libraries
      </Link>
    </div>
  );
}
