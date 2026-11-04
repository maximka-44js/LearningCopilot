import { z } from "zod";

import { ENTITY_TITLE_MAX_LENGTH } from "@/lib/validation-limits";

export const inputClassName =
  "h-10 rounded-md border border-slate-800 bg-slate-950/70 px-3 text-sm text-slate-100";

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

export const aggregationSchema = z.object({
  priority_high_weight: cappedWeight,
  priority_low_weight: cappedWeight,
  priority_normal_weight: cappedWeight,
  urgency_24h_weight: cappedWeight,
  urgency_72h_weight: cappedWeight,
  urgency_7d_weight: cappedWeight,
});

export const calibrationSchema = z.object({
  gap_threshold: percent,
  penalty_multiplier: penaltyMultiplier,
});

export const recommendationSchema = z.object({
  acceptable_mastery_threshold: percent,
  calibration_gap_threshold: percent,
  low_mastery_threshold: percent,
  max_recommendations: z
    .number()
    .int("Введите целое число.")
    .min(1, "Значение должно быть не меньше 1.")
    .max(10, "Максимум — 10 рекомендаций.")
    .optional(),
});

export const riskSchema = z.object({
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

export const projectSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Введите название проекта.")
    .max(ENTITY_TITLE_MAX_LENGTH, `Название не должно превышать ${ENTITY_TITLE_MAX_LENGTH} символов.`),
  description: z.string().trim().max(4000).optional().or(z.literal("")),
  deadline: z
    .string()
    .optional()
    .refine(
      (value) => !value || !Number.isNaN(Date.parse(value)),
      "Введите корректную дату дедлайна."
    )
    .or(z.literal("")),
  priority: z.enum(["low", "normal", "high"]),
  metadataTrack: z.string().trim().max(200).optional().or(z.literal("")),
});

export type PolicySectionState<T> = {
  values: Record<keyof T, string>;
  initial: Record<keyof T, string>;
};

export type ProjectFormValues = {
  name: string;
  description: string;
  deadline: string;
  priority: "low" | "normal" | "high";
  metadataTrack: string;
};

export type ProjectFormState = {
  values: ProjectFormValues;
  initial: ProjectFormValues;
};

export const defaultProjectForm: ProjectFormValues = {
  name: "",
  description: "",
  deadline: "",
  priority: "normal",
  metadataTrack: "",
};

export const initPolicyState = <T extends object>(
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

export const mergePolicyState = <T extends object>(
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

export const isDirty = <T,>(state: PolicySectionState<T>) => {
  return Object.keys(state.values).some((key) => {
    const typedKey = key as keyof T;
    return state.values[typedKey] !== state.initial[typedKey];
  });
};

export const isProjectDirty = (state: ProjectFormState) => {
  return (Object.keys(state.values) as Array<keyof ProjectFormValues>).some(
    (key) => state.values[key] !== state.initial[key]
  );
};

export const toNumberOrUndefined = (value: string) => {
  if (value.trim() === "") return undefined;
  const parsed = Number(value);
  return Number.isNaN(parsed) ? undefined : parsed;
};

export const toLocalDatetimeInput = (value?: string) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate()
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

export const toIsoDate = (value: string) => {
  if (!value) return undefined;
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return undefined;
  return parsed.toISOString();
};
