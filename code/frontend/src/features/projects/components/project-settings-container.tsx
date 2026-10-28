"use client";

import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import ProjectPolicySection from "@/features/projects/components/settings/project-policy-section";
import ProjectProfileCard from "@/features/projects/components/settings/project-profile-card";
import ProjectStatusCard from "@/features/projects/components/settings/project-status-card";
import {
  aggregationSchema,
  calibrationSchema,
  defaultProjectForm,
  initPolicyState,
  isDirty,
  isProjectDirty,
  mergePolicyState,
  projectSchema,
  recommendationSchema,
  riskSchema,
  toIsoDate,
  toLocalDatetimeInput,
  toNumberOrUndefined,
  type PolicySectionState,
  type ProjectFormState,
  type ProjectFormValues,
} from "@/features/projects/components/settings/project-settings-utils";
import { useToast } from "@/components/providers/toast-provider";
import {
  getGetProjectPoliciesQueryKey,
  useGetProjectPolicies,
  useSetProjectAggregationPolicy,
  useSetProjectCalibrationPolicy,
  useSetProjectRecommendationPolicy,
  useSetProjectRiskPolicy,
} from "@/generated/api/policies/policies";
import {
  getGetProjectByIdQueryKey,
  getListProjectsQueryKey,
  useActivateProject,
  useArchiveProject,
  useGetProjectById,
  usePauseProject,
  useUpdateProject,
} from "@/generated/api/projects/projects";
import type {
  GithubComMaximka44jsLcInternalProjectsEntityAggregationPolicy,
  GithubComMaximka44jsLcInternalProjectsEntityCalibrationPolicy,
  GithubComMaximka44jsLcInternalProjectsEntityRecommendationPolicy,
  GithubComMaximka44jsLcInternalProjectsEntityRiskPolicy,
  GithubComMaximka44jsLcInternalProjectsEntityProject as Project,
} from "@/generated/model";
import type { InternalProjectsControllerRestapiV1GetProjectPoliciesResponse } from "@/generated/model/internalProjectsControllerRestapiV1GetProjectPoliciesResponse";
import { getOrvalStatus, unwrapOrvalData } from "@/features/projects/lib/orval";
import { orvalAxios } from "@/lib/orval/mutator";

export default function ProjectSettingsContainer({
  projectId,
}: {
  projectId: string;
}) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { toast } = useToast();
  const [projectState, setProjectState] = useState<ProjectFormState>({
    values: defaultProjectForm,
    initial: defaultProjectForm,
  });
  const [projectError, setProjectError] = useState<string | null>(null);
  const [projectAction, setProjectAction] = useState<
    "activate" | "pause" | "archive" | null
  >(null);
  const [sectionsOpen, setSectionsOpen] = useState({
    profile: false,
    policies: false,
    status: false,
  });

  const { data: projectData, isLoading: projectLoading, isError: projectErrorFlag } =
    useGetProjectById(projectId);
  const {
    data: policiesData,
    isLoading: policiesLoading,
    isError: policiesError,
  } = useGetProjectPolicies(projectId);

  const project = unwrapOrvalData<Project>(projectData);
  const policies = unwrapOrvalData<InternalProjectsControllerRestapiV1GetProjectPoliciesResponse>(
    policiesData
  );

  useEffect(() => {
    if (!project) return;
    const nextValues: ProjectFormValues = {
      name: project.name ?? "",
      description: project.description ?? "",
      deadline: toLocalDatetimeInput(project.deadline_at),
      priority: (project.priority as ProjectFormValues["priority"]) ?? "normal",
      metadataTrack: project.metadata?.track ?? "",
    };
    setProjectState({ values: nextValues, initial: nextValues });
  }, [project?.id, project?.updated_at]);

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
    if (!policies) return;
    const typedPolicies = policies as InternalProjectsControllerRestapiV1GetProjectPoliciesResponse & {
      Aggregation?: GithubComMaximka44jsLcInternalProjectsEntityAggregationPolicy;
      Calibration?: GithubComMaximka44jsLcInternalProjectsEntityCalibrationPolicy;
      Recommendation?: GithubComMaximka44jsLcInternalProjectsEntityRecommendationPolicy;
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
  }, [policies]);

  const { mutateAsync: updateProject, isPending: projectSaving } =
    useUpdateProject({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetProjectByIdQueryKey(projectId),
          });
          queryClient.invalidateQueries({
            queryKey: getListProjectsQueryKey(),
          });
        },
      },
    });

  const { mutateAsync: activateProject, isPending: activating } =
    useActivateProject({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetProjectByIdQueryKey(projectId),
          });
          queryClient.invalidateQueries({
            queryKey: getListProjectsQueryKey(),
          });
        },
      },
    });

  const { mutateAsync: pauseProject, isPending: pausing } = usePauseProject({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetProjectByIdQueryKey(projectId),
        });
        queryClient.invalidateQueries({
          queryKey: getListProjectsQueryKey(),
        });
      },
    },
  });

  const { mutateAsync: archiveProject, isPending: archiving } =
    useArchiveProject({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetProjectByIdQueryKey(projectId),
          });
          queryClient.invalidateQueries({
            queryKey: getListProjectsQueryKey(),
          });
        },
      },
    });

  const { mutateAsync: saveAggregation, isPending: savingAggregation } =
    useSetProjectAggregationPolicy({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetProjectPoliciesQueryKey(projectId),
          });
        },
      },
    });

  const { mutateAsync: saveCalibration, isPending: savingCalibration } =
    useSetProjectCalibrationPolicy({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetProjectPoliciesQueryKey(projectId),
          });
        },
      },
    });

  const { mutateAsync: saveRecommendation, isPending: savingRecommendation } =
    useSetProjectRecommendationPolicy({
      mutation: {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: getGetProjectPoliciesQueryKey(projectId),
          });
        },
      },
    });

  const { mutateAsync: saveRisk, isPending: savingRisk } = useSetProjectRiskPolicy({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetProjectPoliciesQueryKey(projectId),
        });
      },
    },
  });

  const { mutateAsync: deleteProject, isPending: deleting } = useMutation({
    mutationFn: async () => {
      return orvalAxios(`/projects/${projectId}`, { method: "DELETE" });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getListProjectsQueryKey(),
      });
      toast({
        title: "Project deleted",
        description: "The project and its settings were removed.",
      });
      router.push("/projects");
    },
    onError: () => {
      toast({
        title: "Project not deleted",
        description: "Unable to delete the project right now.",
        variant: "error",
      });
    },
  });

  const handleProjectSave = async () => {
    setProjectError(null);

    if (!isProjectDirty(projectState)) {
      return;
    }

    const parsed = projectSchema.safeParse(projectState.values);
    if (!parsed.success) {
      const issue = parsed.error.issues[0];
      const message = issue?.message ?? "Check the form values.";
      setProjectError(message);
      toast({
        title: "Project settings invalid",
        description: message,
        variant: "error",
      });
      return;
    }

    try {
      const deadline_at = parsed.data.deadline
        ? toIsoDate(parsed.data.deadline)
        : undefined;
      const nextTrack = parsed.data.metadataTrack?.trim() ?? "";
      const initialTrack = projectState.initial.metadataTrack.trim();
      let metadata: Record<string, string> | undefined;
      if (nextTrack) {
        metadata = { source: "settings", track: nextTrack };
      } else if (initialTrack) {
        metadata = { source: "settings", track: "" };
      }

      const response = await updateProject({
        id: projectId,
        data: {
          name: parsed.data.name.trim(),
          description: parsed.data.description?.trim() || undefined,
          deadline_at,
          priority: parsed.data.priority,
          metadata,
        },
      });

      const status = getOrvalStatus(response);
      if (status && status !== 200) {
        setProjectError("Project could not be updated.");
        toast({
          title: "Project not saved",
          description: "Server rejected the update.",
          variant: "error",
        });
        return;
      }

      setProjectState((current) => ({
        values: { ...current.values },
        initial: { ...current.values },
      }));
      toast({
        title: "Project updated",
        description: "Your project details were saved.",
      });
    } catch (error) {
      setProjectError("Project could not be updated.");
      toast({
        title: "Project not saved",
        description: "Something went wrong while saving.",
        variant: "error",
      });
    }
  };

  const handlePolicySave = async <T extends object>(
    label: string,
    section: PolicySectionState<T>,
    setSection: (state: PolicySectionState<T>) => void,
    save: (payload: { id: string; data: { policy: T } }) => Promise<unknown>
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
      const response = await save({ id: projectId, data: { policy: payload } });
      const status = getOrvalStatus(response);
      if (status && status !== 200) {
        toast({
          title: `${label} not saved`,
          description: "Server rejected the update.",
          variant: "error",
        });
        return;
      }
      setSection({ values: { ...section.values }, initial: { ...section.values } });
      toast({
        title: `${label} saved`,
        description: "Project policy overrides were updated.",
      });
    } catch (error) {
      toast({
        title: `${label} not saved`,
        description: "Something went wrong while saving.",
        variant: "error",
      });
    }
  };

  const handleProjectAction = async (
    action: "activate" | "pause" | "archive"
  ) => {
    try {
      setProjectAction(action);
      if (action === "activate") {
        await activateProject({ id: projectId });
      } else if (action === "pause") {
        await pauseProject({ id: projectId });
      } else {
        await archiveProject({ id: projectId });
      }
    } finally {
      setProjectAction(null);
    }
  };

  const handleDeleteProject = async () => {
    const confirmed = window.confirm(
      "Delete this project? This cannot be undone."
    );
    if (!confirmed) {
      return;
    }
    await deleteProject();
  };

  const isProjectLoading = projectLoading;
  const hasProjectError = projectErrorFlag || !project;

  if (isProjectLoading) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-slate-400">
        Loading project settings...
      </div>
    );
  }

  if (hasProjectError) {
    return (
      <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-12 text-center text-sm text-rose-200">
        Unable to load this project right now.
      </div>
    );
  }

  return (
    <section className="space-y-10">
      <header className="space-y-3">
        <p className="text-xs uppercase tracking-[0.3em] text-slate-500">
          Project settings
        </p>
        <h1 className="text-3xl font-semibold">
          {project?.name || "Project settings"}
        </h1>
        <p className="max-w-2xl text-sm text-slate-400">
          Manage policies, core fields, and lifecycle controls for this project.
        </p>
      </header>

      <div className="rounded-xl border border-slate-800/70 bg-slate-950/40 px-5 py-4">
        <button
          type="button"
          onClick={() =>
            setSectionsOpen((current) => ({
              ...current,
              profile: !current.profile,
            }))
          }
          className="flex w-full items-center justify-between text-left text-sm font-semibold text-slate-100"
        >
          <span>Project profile</span>
          <span className="text-lg text-slate-400">
            {sectionsOpen.profile ? "-" : "+"}
          </span>
        </button>
      </div>
      {sectionsOpen.profile ? (
        <ProjectProfileCard
          values={projectState.values}
          onChange={(next) =>
            setProjectState((current) => ({
              ...current,
              values: { ...current.values, ...next },
            }))
          }
          onSave={handleProjectSave}
          saving={projectSaving}
          dirty={isProjectDirty(projectState)}
          error={projectError}
        />
      ) : null}

      <div className="space-y-4">
        <div className="rounded-xl border border-slate-800/70 bg-slate-950/40 px-5 py-4">
          <button
            type="button"
            onClick={() =>
              setSectionsOpen((current) => ({
                ...current,
                policies: !current.policies,
              }))
            }
            className="flex w-full items-center justify-between text-left"
          >
            <div>
              <h2 className="text-sm font-semibold text-slate-100">Policy overrides</h2>
              <p className="text-sm text-slate-400">
                Adjust policy weights and thresholds for this specific project.
              </p>
            </div>
            <span className="text-lg text-slate-400">
              {sectionsOpen.policies ? "-" : "+"}
            </span>
          </button>
        </div>
        {sectionsOpen.policies ? (
          policiesLoading ? (
            <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-8 text-sm text-slate-400">
              Loading project policies...
            </div>
          ) : policiesError || !policies ? (
            <div className="rounded-xl border border-slate-800 bg-slate-950/40 px-6 py-8 text-sm text-rose-200">
              Unable to load project policies.
            </div>
          ) : (
            <div className="grid gap-6 lg:grid-cols-2">
            <ProjectPolicySection
              title="Aggregation"
              description="Controls how priorities and deadlines shape readiness."
              values={aggregation.values as Record<string, string>}
              onChange={(key, value) =>
                setAggregation((current) => ({
                  ...current,
                  values: {
                    ...current.values,
                    [key]: value,
                  },
                }))
              }
              onSave={() =>
                handlePolicySave(
                  "Aggregation",
                  aggregation,
                  setAggregation,
                  saveAggregation
                )
              }
              saving={savingAggregation}
              dirty={isDirty(aggregation)}
            />
            <ProjectPolicySection
              title="Calibration"
              description="Tunes how we correct overconfidence vs results."
              values={calibration.values as Record<string, string>}
              onChange={(key, value) =>
                setCalibration((current) => ({
                  ...current,
                  values: {
                    ...current.values,
                    [key]: value,
                  },
                }))
              }
              onSave={() =>
                handlePolicySave(
                  "Calibration",
                  calibration,
                  setCalibration,
                  saveCalibration
                )
              }
              saving={savingCalibration}
              dirty={isDirty(calibration)}
            />
            <ProjectPolicySection
              title="Recommendation"
              description="Configures thresholds for recommendation selection."
              values={recommendation.values as Record<string, string>}
              onChange={(key, value) =>
                setRecommendation((current) => ({
                  ...current,
                  values: {
                    ...current.values,
                    [key]: value,
                  },
                }))
              }
              onSave={() =>
                handlePolicySave(
                  "Recommendation",
                  recommendation,
                  setRecommendation,
                  saveRecommendation
                )
              }
              saving={savingRecommendation}
              dirty={isDirty(recommendation)}
            />
            <ProjectPolicySection
              title="Risk"
              description="Policy rules for risk thresholds and gating."
              values={risk.values as Record<string, string>}
              onChange={(key, value) =>
                setRisk((current) => ({
                  ...current,
                  values: {
                    ...current.values,
                    [key]: value,
                  },
                }))
              }
              onSave={() => handlePolicySave("Risk", risk, setRisk, saveRisk)}
              saving={savingRisk}
              dirty={isDirty(risk)}
            />
            </div>
          )
        ) : null}
      </div>
      <div className="rounded-xl border border-slate-800/70 bg-slate-950/40 px-5 py-4">
        <button
          type="button"
          onClick={() =>
            setSectionsOpen((current) => ({
              ...current,
              status: !current.status,
            }))
          }
          className="flex w-full items-center justify-between text-left text-sm font-semibold text-slate-100"
        >
          <span>Project status</span>
          <span className="text-lg text-slate-400">
            {sectionsOpen.status ? "-" : "+"}
          </span>
        </button>
      </div>
      {sectionsOpen.status ? (
        <ProjectStatusCard
          status={project?.status}
          action={projectAction}
          onAction={handleProjectAction}
          activating={activating}
          pausing={pausing}
          archiving={archiving}
          onDelete={handleDeleteProject}
          deleting={deleting}
        />
      ) : null}
    </section>
  );
}
