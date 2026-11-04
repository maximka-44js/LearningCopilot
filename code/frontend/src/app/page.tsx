import { auth } from "@/auth";
import { UserIdCard } from "@/features/auth/components/user-id-card";

export default async function Home() {
  const session = await auth();

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-20 text-slate-100">
      <div className="mx-auto max-w-3xl space-y-4">
        <p className="text-sm uppercase tracking-[0.2em] text-slate-500">
          Learning Copilot
        </p>
        <h1 className="text-4xl font-semibold">Учитесь системно</h1>
        <p className="text-slate-400">
          Планируйте подготовку, работайте с материалами и отслеживайте
          готовность к цели.
        </p>
        {session?.user?.id ? <UserIdCard userId={session.user.id} /> : null}
      </div>
    </main>
  );
}
