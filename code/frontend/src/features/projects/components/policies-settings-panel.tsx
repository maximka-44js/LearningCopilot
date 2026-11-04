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

const sectionTitles = {
  Aggregation: "Агрегация",
  Calibration: "Калибровка",
  Recommendation: "Рекомендации",
  Risk: "Риски",
} as const;

const fieldTitles: Record<string, string> = {
  priority_high_weight: "Вес высокого приоритета",
  priority_low_weight: "Вес низкого приоритета",
  priority_normal_weight: "Вес обычного приоритета",
  urgency_24h_weight: "Вес срочности: 24 часа",
  urgency_72h_weight: "Вес срочности: 72 часа",
  urgency_7d_weight: "Вес срочности: 7 дней",
  gap_threshold: "Порог расхождения",
  penalty_multiplier: "Множитель штрафа",
  acceptable_mastery_threshold: "Допустимый порог освоения",
  calibration_gap_threshold: "Порог расхождения для калибровки",
  low_mastery_threshold: "Порог низкого уровня освоения",
  max_recommendations: "Максимум рекомендаций",
  block_critical_mastery_below: "Контекст: освоение ниже критического уровня",
  block_critical_readiness_below: "Контекст: готовность ниже критического уровня",
  block_high_mastery_below: "Контекст: освоение ниже высокого уровня",
  block_high_readiness_below: "Контекст: готовность ниже высокого уровня",
  block_medium_mastery_below: "Контекст: освоение ниже среднего уровня",
  block_medium_readiness_below: "Контекст: готовность ниже среднего уровня",
  false_confidence_gap: "Расхождение ложной уверенности",
  task_critical_readiness_below: "Задача: готовность ниже критического уровня",
  task_high_readiness_below: "Задача: готовность ниже высокого уровня",
  task_medium_readiness_below: "Задача: готовность ниже среднего уровня",
};

const nonNegative = z
  .number()
  .min(0, "Значение не может быть отрицательным.")
  .optional();
const cappedWeight = z
  .number()
  .min(0, "Вес не может быть отрицательным.")
  .max(10, "Вес слишком большой (максимум — 10).")
  .optional();
const penaltyMultiplier = z
  .number()
  .min(0, "Штраф не может быть отрицательным.")
  .max(5, "Штраф слишком большой (максимум — 5).")
  .optional();
const percent = z
  .number()
  .min(0, "Значение должно быть от 0 до 100.")
  .max(100, "Значение должно быть от 0 до 100.")
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
    .int("Введите целое число.")
    .min(1, "Значение должно быть не меньше 1.")
    .max(10, "Максимальное количество рекомендаций — 10.")
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
          title: "Некорректное число",
          description: `Проверьте поле «${fieldTitles[key] ?? key}» в разделе «${
            sectionTitles[label as keyof typeof sectionTitles]
          }».`,
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
        title: `Некорректные значения: ${
          sectionTitles[label as keyof typeof sectionTitles]
        }`,
        description: issue?.message ?? "Проверьте поля и попробуйте ещё раз.",
        variant: "error",
      });
      return;
    }

    try {
      const response = await save({ data: { policy: payload } });
      if (typeof (response as { status?: unknown }).status === "number") {
        if ((response as { status: number }).status !== 200) {
          toast({
            title: `Не удалось сохранить: ${
              sectionTitles[label as keyof typeof sectionTitles]
            }`,
            description: "Сервер отклонил обновление.",
            variant: "error",
          });
          return;
        }
      }
      setSection({
        values: { ...section.values },
        initial: { ...section.values },
      });
      toast({
        title: `${
          sectionTitles[label as keyof typeof sectionTitles]
        } сохранены`,
        description: "Настройки политик обновлены.",
      });
    } catch (error) {
      toast({
        title: `Не удалось сохранить: ${
          sectionTitles[label as keyof typeof sectionTitles]
        }`,
        description: "При сохранении произошла ошибка.",
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
        Загрузка политик...
      </div>
    );
  }

  if (isError || !rawPolicies) {
    return (
      <div className="rounded-lg border border-slate-800 bg-slate-950/40 px-6 py-6 text-sm text-rose-200">
        Не удалось загрузить политики.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <section className="rounded-xl border border-slate-800/70 bg-slate-950/40 p-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-sm font-semibold text-slate-100">
              {sectionTitles.Aggregation}
            </h3>
            <p className="text-sm text-slate-400">
              Определяет, как приоритеты и сроки влияют на готовность.
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
            {savingAggregation ? "Сохранение..." : "Сохранить"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(aggregation.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {fieldTitles[key] ?? String(key).replaceAll("_", " ")}
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
            <h3 className="text-sm font-semibold text-slate-100">
              {sectionTitles.Calibration}
            </h3>
            <p className="text-sm text-slate-400">
              Настраивает коррекцию излишней уверенности по сравнению с результатами.
            </p>
          </div>
          <Button
            size="sm"
            onClick={() =>
              handleSave("Calibration", calibration, setCalibration, saveCalibration)
            }
            disabled={savingCalibration || !isDirty(calibration)}
          >
            {savingCalibration ? "Сохранение..." : "Сохранить"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(calibration.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {fieldTitles[key] ?? String(key).replaceAll("_", " ")}
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
              {sectionTitles.Recommendation}
            </h3>
            <p className="text-sm text-slate-400">
              Определяет, когда и сколько следующих действий мы предлагаем.
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
            {savingRecommendation ? "Сохранение..." : "Сохранить"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(recommendation.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {fieldTitles[key] ?? String(key).replaceAll("_", " ")}
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
            <h3 className="text-sm font-semibold text-slate-100">
              {sectionTitles.Risk}
            </h3>
            <p className="text-sm text-slate-400">
              Определяет, когда ограниченный контекст или задача считаются рискованными.
            </p>
          </div>
          <Button
            size="sm"
            onClick={() => handleSave("Risk", risk, setRisk, saveRisk)}
            disabled={savingRisk || !isDirty(risk)}
          >
            {savingRisk ? "Сохранение..." : "Сохранить"}
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {Object.entries(risk.values).map(([key, value]) => (
            <label key={key} className="text-sm text-slate-300">
              <span className="block text-xs uppercase tracking-wide text-slate-500">
                {fieldTitles[key] ?? String(key).replaceAll("_", " ")}
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
