"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import PoliciesSettingsPanel from "@/features/projects/components/policies-settings-panel";

type PoliciesSettingsModalProps = {
  open: boolean;
  onClose: () => void;
};

export default function PoliciesSettingsModal({ open, onClose }: PoliciesSettingsModalProps) {
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!open || !isMounted) {
    return null;
  }

  return createPortal(
    <div className="fixed inset-0 z-[9999] grid place-items-center bg-slate-950/70 p-4 backdrop-blur-sm">
      <Card className="w-full max-w-5xl">
        <CardHeader className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <CardTitle className="text-[1.2rem]">Policy settings</CardTitle>
            <p className="mt-1 text-sm text-slate-400">
              Adjust global defaults for aggregation, calibration, recommendation,
              and risk.
            </p>
          </div>
          <Button variant="outline" size="sm" onClick={onClose}>
            Close
          </Button>
        </CardHeader>
        <CardContent className="space-y-6">
          <PoliciesSettingsPanel />
        </CardContent>
      </Card>
    </div>,
    document.body
  );
}
