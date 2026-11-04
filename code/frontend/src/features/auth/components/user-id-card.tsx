"use client";

import { Check, Copy } from "lucide-react";
import { useState } from "react";

import { useToast } from "@/components/providers/toast-provider";
import { Button } from "@/components/ui/button";

type UserIdCardProps = {
  userId: string;
};

export function UserIdCard({ userId }: UserIdCardProps) {
  const [copied, setCopied] = useState(false);
  const { toast } = useToast();

  const copyUserId = async () => {
    try {
      await navigator.clipboard.writeText(userId);
      setCopied(true);
      toast({
        title: "ID скопирован",
        description: "Ваш ID скопирован в буфер обмена.",
      });
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      toast({
        title: "Не удалось скопировать",
        description: "Скопируйте ID вручную.",
        variant: "error",
      });
    }
  };

  return (
    <section className="rounded-xl border border-slate-800/80 bg-slate-900/40 p-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="space-y-1">
          <p className="text-sm font-medium text-slate-200">Ваш ID</p>
          <p className="break-all font-mono text-sm text-slate-400">{userId}</p>
        </div>
        <Button type="button" variant="outline" onClick={copyUserId}>
          {copied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
          {copied ? "Скопировано" : "Копировать"}
        </Button>
      </div>
    </section>
  );
}
