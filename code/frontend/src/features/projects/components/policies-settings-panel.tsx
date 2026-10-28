"use client";

import { useEffect, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import {
  getGetUserPoliciesQueryKey,
  useGetUserPolicies,
  useSetAggregationPolicy,
  useSetCalibrationPolicy,
  useSetRecommendationPolicy,
  useSetRiskPolicy,
} from "@/generated/api/policies/policies";
import type {
  GithubComMaximka44jsLcInternalProjectsEntityAggregationPolicy,
  GithubComMaximka44jsLcInternalProjectsEntityCalibrationPolicy,
  GithubComMaximka44jsLcInternalProjectsEntityRecommendationPolicy,
  GithubComMaximka44jsLcInternalProjectsEntityRiskPolicy,
} from "@/generated/model";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/providers/toast-provider";

type PolicySectionState<T> = {
  values: Record<keyof T, string>;
  initial: Record<keyof T, string>;
};

const initPolicyState = <T extends object>(
  policy?: T
): PolicySectionState<T> => {
  const values = {} as Record<keyof T, string>;
  const keys = Object.keys(policy ?? {}) as Array<keyof T>;
  keys.forEach((typedKey) => {
    const value = policy?.[typedKey] as number | undefined;
    values[typedKey] = value === null || value === undefined ? "" : String(value);
  });

  return { values: { ...values }, initial: { ...values } };
};

const mergePolicyState = <T extends object>(
  policy: T | undefined,
  prev: PolicySectionState<T>
) => {
  if (!policy) {
    return prev;
  }

  const nextValues = { ...prev.values } as Record<keyof T, string>;
  const keys = Object.keys(policy ?? {}) as Array<keyof T>;
  keys.forEach((typedKey) => {
    const value = policy?.[typedKey] as number | undefined;
    nextValues[typedKey] = value === null || value === undefined ? "" : String(value);
  });
  return { values: nextValues, initial: { ...nextValues } };
};

const isDirty = <T,>(state: PolicySectionState<T>) => {
  return Object.keys(state.values).some((key) => {
    const typedKey = key as keyof T;
    return state.values[typedKey] !== state.initial[typedKey];
  });
};

const toNumberOrUndefined = (value: string) => {
  if (value.trim() === "") return undefined;
  const parsed = Number(value);
  return Number.isNaN(parsed) ? undefined : parsed;
};

const inputClassName =
  "h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100";

const nonNegative = z.number().min(0, "Value must be 0 or higher.").optional();
const cappedWeight = z
  .number()
  .min(0, "Weight must be 0 or higher.")
  .max(10, "Weight looks too large (max 10).")
  .optional();
const penaltyMultiplier = z
  .number()
  .min(0, "Penalty must be 0 or higher.")
  .max(5, "Penalty looks too large (max 5).")
  .optional();
const percent = z
  .number()
  .min(0, "Value must be between 0 and 100.")
  .max(100, "Value must be between 0 and 100.")
  .optional();

const aggregationSchema = z.object({
  priority_high_weight: cappedWeight,
  priority_low_weight: cappedWeight,
  priority_normal_weight: cappedWeight,
  urgency_24h_weight: cappedWeight,
  urgency_72h_weight: cappedWeight,
  urgency_7d_weight: cappedWeight,
});

const calibrationSchema = z.object({
  gap_threshold: percent,
  penalty_multiplier: penaltyMultiplier,
});

const recommendationSchema = z.object({
  acceptable_mastery_threshold: percent,
  calibration_gap_threshold: percent,
  low_mastery_threshold: percent,
  max_recommendations: z
    .number()
    .int("Must be a whole number.")
    .min(1, "Must be at least 1.")
    .max(10, "Max recommendations capped at 10.")
    .optional(),
});

const riskSchema = z.object({
  block_critical_mastery_below: percent,
  block_critical_readiness_below: percent,
  block_high_mastery_below: percent,
  block_high_readiness_below: percent,
  block_medium_mastery_below: percent,
  block_medium_readiness_below: percent,
  false_confidence_gap: percent,
  task_critical_readiness_below: percent,
  task_high_readiness_below: percent,
  task_medium_readiness_below: percent,
});

export default function PoliciesSettingsPanel() {
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const { data, isLoading, isError } = useGetUserPolicies();

  const rawPolicies = useMemo(() => {
    if (Array.isArray(data)) {
      return null;
    }
    if (data && typeof (data as { status?: unknown }).status === "number") {
      return (data as { status: number; data: unknown }).status === 200
        ? (data as { data: unknown }).data
        : null;
    }
    return data ?? null;
  }, [data]);

  const [aggregation, setAggregation] = useState<
    PolicySectionState<GithubComMaximka44jsLcInternalProjectsEntityAggregationPolicy>
  >(initPolicyState());
  const [calibration, setCalibration] = useState<
    PolicySectionState<GithubComMaximka44jsLcInternalProjectsEntityCalibrationPolicy>
  >(initPolicyState());
  const [recommendation, setRecommendation] = useState<
    PolicySectionState<GithubComMaximka44jsLcInternalProjectsEntityRecommendationPolicy>
  >(initPolicyState());
  const [risk, setRisk] = useState<
    PolicySectionState<GithubComMaximka44jsLcInternalProjectsEntityRiskPolicy>
  >(initPolicyState());

  useEffect(() => {
    if (!rawPolicies || typeof rawPolicies !== "object") return;
    const typedPolicies = rawPolicies as {
      aggregation?: GithubComMaximka44jsLcInternalProjectsEntityAggregationPolicy;
      Aggregation?: GithubComMaximka44jsLcInternalProjectsEntityAggregationPolicy;
      calibration?: GithubComMaximka44jsLcInternalProjectsEntityCalibrationPolicy;
      Calibration?: GithubComMaximka44jsLcInternalProjectsEntityCalibrationPolicy;
      recommendation?: GithubComMaximka44jsLcInternalProjectsEntityRecommendationPolicy;
      Recommendation?: GithubComMaximka44jsLcInternalProjectsEntityRecommendationPolicy;
      risk?: GithubComMaximka44jsLcInternalProjectsEntityRiskPolicy;
      Risk?: GithubComMaximka44jsLcInternalProjectsEntityRiskPolicy;
    };

    const aggregationPolicy = typedPolicies.aggregation ?? typedPolicies.Aggregation;
    const calibrationPolicy = typedPolicies.calibration ?? typedPolicies.Calibration;
    const recommendationPolicy =
      typedPolicies.recommendation ?? typedPolicies.Recommendation;
    const riskPolicy = typedPolicies.risk ?? typedPolicies.Risk;

    setAggregation((current) => mergePolicyState(aggregationPolicy, current));
    setCalibration((current) => mergePolicyState(calibrationPolicy, current));
    setRecommendation((current) =>
      mergePolicyState(recommendationPolicy, current)
    );
    setRisk((current) => mergePolicyState(riskPolicy, current));
  }, [rawPolicies]);

  const { mutateAsync: saveAggregation, isPending: savingAggregation } =
    useSetAggregationPolicy({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetUserPoliciesQueryKey(),
          });
        },
      },
    });

  const { mutateAsync: saveCalibration, isPending: savingCalibration } =
    useSetCalibrationPolicy({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetUserPoliciesQueryKey(),
          });
        },
      },
    });

  const { mutateAsync: saveRecommendation, isPending: savingRecommendation } =
    useSetRecommendationPolicy({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetUserPoliciesQueryKey(),
          });
        },
      },
    });

  const { mutateAsync: saveRisk, isPending: savingRisk } = useSetRiskPolicy({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetUserPoliciesQueryKey(),
        });
      },
    },
  });

  const handleSave = async <T extends object>(
    label: string,
    section: PolicySectionState<T>,
    setSection: (state: PolicySectionState<T>) => void,
    save: (payload: { data: { policy: T } }) => Promise<unknown>
  ) => {
    if (!isDirty(section)) {
      return;
    }

    const payload = {} as T;

    for (const [key, raw] of Object.entries(section.values)) {
      const rawValue = String(raw ?? "");
      const parsed = toNumberOrUndefined(rawValue);
      if (rawValue.trim() !== "" && parsed === undefined) {
        toast({
          title: "Invalid number",
          description: `Check ${label} → ${key}.`,
          variant: "error",
        });
        return;
      }
      (payload as Record<string, number | undefined>)[key] = parsed;
    }

    const schema =
      label === "Aggregation"
        ? aggregationSchema
        : label === "Calibration"
          ? calibrationSchema
          : label === "Recommendation"
            ? recommendationSchema
            : riskSchema;

    const validation = schema.safeParse(payload);
    if (!validation.success) {
      const issue = validation.error.issues[0];
      toast({
        title: `${label} has invalid values`,
        description: issue?.message ?? "Check the fields and try again.",
        variant: "error",
      });
      return;
    }

    try {
      const response = await save({ data: { policy: payload } });
      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 200) {
          toast({
            title: `${label} not saved`,
            description: "Server rejected the update.",
            variant: "error",
          });
          return;
        }
      }
      setSection({ values: { ...section.values }, initial: { ...section.values } });
      toast({
        title: `${label} saved`,
        description: "Your policy settings were updated.",
      });
    } catch (error) {
      toast({
        title: `${label} not saved`,
        description: "Something went wrong while saving.",
        variant: "error",
      });
    }
  };

  if (!isMounted) {
    return null;
  }

  if (isLoading) {
    return (
      <div className="rounded-lg border border-slate-800 bg-slate-950/40 px-6 py-6 text-sm text-slate-400">
        Loading policies...
      </div>
    );
  }

  if (isError || !rawPolicies) {
    return (
      <div className="rounded-lg border border-slate-800 bg-slate-950/40 px-6 py-6 text-sm text-rose-200">
        Unable to load policies.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <section className="rounded-xl border border-slate-800/70 bg-slate-950/40 p-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-sm font-semibold text-slate-100">
              Aggregation
            </h3>
            <p className="text-sm text-slate-400">
              Controls how priorities and deadlines shape readiness.
            </p>
          </div>
          <Button
            size="sm"
            onClick={() =>
              handleSave(
                "Aggregation",
                aggregation,
                setAggregation,
                saveAggregation
              )
            }
            disabled={savingAggregation || !isDirty(aggregation)}
          >
            {savingAggregation ? "Saving" : "Save"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(aggregation.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {String(key).replaceAll("_", " ")}
              </span>
              <input
                type="number"
                value={value}
                onChange={(event) =>
                  setAggregation((current) => ({
                    ...current,
                    values: {
                      ...current.values,
                      [key]: event.target.value,
                    },
                  }))
                }
                className={inputClassName}
              />
            </label>
          ))}
        </div>
      </section>

      <section className="rounded-xl border border-slate-800/70 bg-slate-950/40 p-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-sm font-semibold text-slate-100">Calibration</h3>
            <p className="text-sm text-slate-400">
              Tunes how we correct overconfidence vs results.
            </p>
          </div>
          <Button
            size="sm"
            onClick={() =>
              handleSave("Calibration", calibration, setCalibration, saveCalibration)
            }
            disabled={savingCalibration || !isDirty(calibration)}
          >
            {savingCalibration ? "Saving" : "Save"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(calibration.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {String(key).replaceAll("_", " ")}
              </span>
              <input
                type="number"
                value={value}
                onChange={(event) =>
                  setCalibration((current) => ({
                    ...current,
                    values: {
                      ...current.values,
                      [key]: event.target.value,
                    },
                  }))
                }
                className={inputClassName}
              />
            </label>
          ))}
        </div>
      </section>

      <section className="rounded-xl border border-slate-800/70 bg-slate-950/40 p-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-sm font-semibold text-slate-100">
              Recommendation
            </h3>
            <p className="text-sm text-slate-400">
              Controls when and how many next actions we suggest.
            </p>
          </div>
          <Button
            size="sm"
            onClick={() =>
              handleSave(
                "Recommendation",
                recommendation,
                setRecommendation,
                saveRecommendation
              )
            }
            disabled={savingRecommendation || !isDirty(recommendation)}
          >
            {savingRecommendation ? "Saving" : "Save"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(recommendation.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {String(key).replaceAll("_", " ")}
              </span>
              <input
                type="number"
                value={value}
                onChange={(event) =>
                  setRecommendation((current) => ({
                    ...current,
                    values: {
                      ...current.values,
                      [key]: event.target.value,
                    },
                  }))
                }
                className={inputClassName}
              />
            </label>
          ))}
        </div>
      </section>

      <section className="rounded-xl border border-slate-800/70 bg-slate-950/40 p-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-sm font-semibold text-slate-100">Risk</h3>
            <p className="text-sm text-slate-400">
              Controls when a block or task is considered at risk.
            </p>
          </div>
          <Button
            size="sm"
            onClick={() => handleSave("Risk", risk, setRisk, saveRisk)}
            disabled={savingRisk || !isDirty(risk)}
          >
            {savingRisk ? "Saving" : "Save"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(risk.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {String(key).replaceAll("_", " ")}
              </span>
              <input
                type="number"
                value={value}
                onChange={(event) =>
                  setRisk((current) => ({
                    ...current,
                    values: {
                      ...current.values,
                      [key]: event.target.value,
                    },
                  }))
                }
                className={inputClassName}
              />
            </label>
          ))}
        </div>
      </section>
    </div>
  );
}
