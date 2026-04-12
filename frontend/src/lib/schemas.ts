import { z } from "zod";
import { providerPresets } from "@/lib/providers";

export const connectionSchema = z
  .object({
    name: z.string().min(1, "Name is required"),
    providerId: z.string().min(1),
    apiKey: z.string(),
    baseUrl: z.string().url("Must be a valid URL").or(z.literal("")).optional(),
  })
  .superRefine((data, ctx) => {
    const preset = providerPresets[data.providerId];
    const needsBaseUrl = preset?.requiresBaseUrl ?? false;
    if (needsBaseUrl && !data.baseUrl?.trim()) {
      ctx.addIssue({
        code: "custom",
        path: ["baseUrl"],
        message: "Base URL is required for this provider",
      });
    }
    if (!needsBaseUrl && !data.apiKey.trim()) {
      ctx.addIssue({
        code: "custom",
        path: ["apiKey"],
        message: "API key is required",
      });
    }
  });

export type ConnectionFormData = z.infer<typeof connectionSchema>;

export const modelsSchema = z.object({
  transcriptionConnectionId: z.string().min(1, "Select a connection"),
  transcriptionModel: z.string(),
  transcriptionLanguage: z.string().optional(),
  refinementConnectionId: z.string().min(1, "Select a connection"),
  refinementModel: z.string(),
});

export type ModelsFormData = z.infer<typeof modelsSchema>;

export const globalRulesSchema = z.object({
  globalRules: z.string(),
  builtInRules: z.string(),
});

export type GlobalRulesFormData = z.infer<typeof globalRulesSchema>;

export const fewShotExampleSchema = z.object({
  input: z.string(),
  output: z.string(),
});

export const profileSchema = z.object({
  name: z.string().min(1, "Name is required"),
  connectionId: z.string().optional(),
  model: z.string(),
  prompt: z.string(),
  temperature: z.number().min(0).max(1).optional(),
  topP: z.number().min(0.1).max(1).optional(),
  examples: z.array(fewShotExampleSchema).optional(),
});

export type ProfileFormData = z.infer<typeof profileSchema>;

export const wizardConnectionSchema = z
  .object({
    providerId: z.string().min(1),
    apiKey: z.string(),
    baseUrl: z.string().url("Must be a valid URL").or(z.literal("")).optional(),
  })
  .superRefine((data, ctx) => {
    const preset = providerPresets[data.providerId];
    const needsBaseUrl = preset?.requiresBaseUrl ?? false;
    if (needsBaseUrl && !data.baseUrl?.trim()) {
      ctx.addIssue({
        code: "custom",
        path: ["baseUrl"],
        message: "Base URL is required for this provider",
      });
    }
  });

export type WizardConnectionFormData = z.infer<typeof wizardConnectionSchema>;
