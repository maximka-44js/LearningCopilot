import { requireUserSession } from "@/features/auth/server/require-user-session";
import { redirect } from "next/navigation";

type LibraryTreePageProps = {
  params: Promise<{ id: string }>;
};

export default async function LibraryTreePage({ params }: LibraryTreePageProps) {
  const { id } = await params;

  await requireUserSession(`/storage/libraries/${id}`);
  redirect(`/storage/libraries/${id}`);
}
