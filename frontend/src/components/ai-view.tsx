import { useState, useEffect } from "react";
import { useForm, Controller, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  modelsSchema,
  globalRulesSchema,
  profileSchema,
  type ModelsFormData,
  type GlobalRulesFormData,
  type ProfileFormData,
} from "@/lib/schemas";
import { FormField } from "@/components/ui/form-field";
import {
  Play,
  Loader2,
  Bot,
  Check,
  Trash2,
  Plus,
  ChevronDown,
  ChevronUp,
  RotateCcw,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Slider } from "@/components/ui/slider";
import { WarningBanner } from "@/components/ui/warning-banner";
import { AdvancedToggle } from "@/components/ui/advanced-toggle";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
import { getProfileIcon } from "@/lib/icons";
import { useModelsForConnection } from "@/lib/hooks";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";
import type { ai } from "../../wailsjs/go/models";

// ──────────────────────────────────────────────────
// Connection selector (shared)
// ──────────────────────────────────────────────────

function ConnectionSelect({
  value,
  onChange,
  connections,
  allowDefault,
  defaultLabel,
}: {
  value: string;
  onChange: (v: string) => void;
  connections: config.Connection[];
  allowDefault?: boolean;
  defaultLabel?: string;
}) {
  return (
    <Select
      value={value || "__default__"}
      onValueChange={(v) => onChange(v === "__default__" ? "" : v)}
    >
      <SelectTrigger className="text-xs">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {allowDefault && (
          <>
            <SelectItem value="__default__" className="text-xs">
              {defaultLabel ?? "Use global default"}
            </SelectItem>
            <SelectSeparator />
          </>
        )}
        {connections.length > 0 ? (
          connections.map((c) => (
            <SelectItem key={c.id} value={c.id} className="text-xs">
              {c.name}
            </SelectItem>
          ))
        ) : (
          <SelectGroup>
            <SelectLabel>No connections configured</SelectLabel>
          </SelectGroup>
        )}
      </SelectContent>
    </Select>
  );
}

// ──────────────────────────────────────────────────
// Model selector with custom option
// ──────────────────────────────────────────────────

function ModelSelect({
  value,
  onChange,
  models,
  allowDefault,
  defaultLabel,
}: {
  value: string;
  onChange: (v: string) => void;
  models: ai.ModelInfo[];
  allowDefault?: boolean;
  defaultLabel?: string;
}) {
  const [custom, setCustom] = useState(false);
  const knownIds = new Set(models.map((m) => m.id));
  const isCustomValue = value && !knownIds.has(value);

  if (custom || isCustomValue) {
    return (
      <div className="flex gap-1.5">
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="model-id..."
          className="text-xs"
        />
        <Button
          variant="outline"
          size="sm"
          className="shrink-0 text-xs"
          onClick={() => setCustom(false)}
        >
          List
        </Button>
      </div>
    );
  }

  return (
    <div className="flex gap-1.5">
      <Select
        value={value || "__default__"}
        onValueChange={(v) => onChange(v === "__default__" ? "" : v)}
      >
        <SelectTrigger className="text-xs">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {allowDefault && (
            <>
              <SelectItem value="__default__" className="text-xs">
                {defaultLabel ?? "Use global default"}
              </SelectItem>
              <SelectSeparator />
            </>
          )}
          {models.length > 0 ? (
            <SelectGroup>
              <SelectLabel>Available</SelectLabel>
              {models.map((m) => (
                <SelectItem key={m.id} value={m.id} className="text-xs">
                  {m.name}
                </SelectItem>
              ))}
            </SelectGroup>
          ) : (
            <SelectGroup>
              <SelectLabel>No models loaded</SelectLabel>
            </SelectGroup>
          )}
        </SelectContent>
      </Select>
      <Button
        variant="outline"
        size="sm"
        className="shrink-0 text-xs"
        onClick={() => setCustom(true)}
      >
        Custom
      </Button>
    </div>
  );
}

// ──────────────────────────────────────────────────
// Main component
// ──────────────────────────────────────────────────

interface AiViewProps {
  settings: config.Settings;
  configured: boolean;
  onSave: (settings: config.Settings) => void;
}

export function AiView({ settings, configured, onSave }: AiViewProps) {
  const connections = settings.connections ?? [];

  // ── Models form ──
  const modelsForm = useForm<ModelsFormData>({
    resolver: zodResolver(modelsSchema),
    defaultValues: {
      transcriptionConnectionId: settings.transcriptionConnectionId,
      transcriptionModel: settings.transcriptionModel,
      refinementConnectionId: settings.refinementConnectionId,
      refinementModel: settings.refinementModel,
    },
  });
  const transConnId = modelsForm.watch("transcriptionConnectionId");
  const refConnId = modelsForm.watch("refinementConnectionId");
  const refModel = modelsForm.watch("refinementModel");
  const { transcriptionModels } = useModelsForConnection(transConnId);
  const { chatModels: refChatModels } = useModelsForConnection(refConnId);

  const saveModels = (data: ModelsFormData) => {
    onSave(config.Settings.createFrom({ ...settings, ...data }));
  };

  // ── Global rules form ──
  const rulesForm = useForm<GlobalRulesFormData>({
    resolver: zodResolver(globalRulesSchema),
    defaultValues: {
      globalRules: settings.globalRules ?? "",
      builtInRules: settings.builtInRules ?? "",
    },
  });
  const [globalAdvancedOpen, setGlobalAdvancedOpen] = useState(false);

  useEffect(() => {
    if (!settings.builtInRules) {
      AppBridge.GetDefaultBuiltInRules().then((rules) =>
        rulesForm.setValue("builtInRules", rules),
      );
    }
  }, [settings.builtInRules, rulesForm]);

  const saveRules = (data: GlobalRulesFormData) => {
    onSave(
      config.Settings.createFrom({
        ...settings,
        globalRules: data.globalRules,
        builtInRules: data.builtInRules || undefined,
      }),
    );
  };

  // ── Profiles (list-level state, per-card forms in Phase 4) ──
  const [draftProfiles, setDraftProfiles] = useState(
    settings.refinementProfiles ?? [],
  );
  const [activeId, setActiveId] = useState(settings.activeProfileId);
  const [expandedProfile, setExpandedProfile] = useState<string | null>(null);
  const [advancedOpen, setAdvancedOpen] = useState<string | null>(null);

  // ── Test playground (UI state only) ──
  const [sampleText, setSampleText] = useState("");
  const [testResult, setTestResult] = useState("");
  const [testing, setTesting] = useState(false);

  const saveProfiles = (
    profiles?: typeof draftProfiles,
    newActiveId?: string,
  ) => {
    const p = profiles ?? draftProfiles;
    const a = newActiveId ?? activeId;
    onSave(
      config.Settings.createFrom({
        ...settings,
        refinementProfiles: p,
        activeProfileId: a,
      }),
    );
  };

  const saveProfile = (id: string, data: ProfileFormData) => {
    const updated = draftProfiles.map((p) =>
      p.id === id
        ? config.RefinementProfile.createFrom({ ...p, ...data, connectionId: data.connectionId || undefined })
        : p,
    );
    setDraftProfiles(updated);
    saveProfiles(updated);
  };

  const deleteProfile = (id: string) => {
    const updated = draftProfiles.filter((p) => p.id !== id);
    const newActive = activeId === id ? updated[0]?.id ?? "" : activeId;
    setDraftProfiles(updated);
    setActiveId(newActive);
    saveProfiles(updated, newActive);
  };

  const addProfile = () => {
    const id = `custom-${Date.now()}`;
    const newProfile = config.RefinementProfile.createFrom({
      id,
      name: "New Style",
      icon: "pen-tool",
      model: "",
      prompt: "",
    });
    const updated = [...draftProfiles, newProfile];
    setDraftProfiles(updated);
    setExpandedProfile(id);
  };

  const setActive = (id: string) => {
    setActiveId(id);
    saveProfiles(undefined, id);
  };

  const applyDefaultsToAll = () => {
    const updated = draftProfiles.map((p) =>
      config.RefinementProfile.createFrom({
        ...p,
        connectionId: undefined,
        model: "",
      }),
    );
    setDraftProfiles(updated);
    saveProfiles(updated);
    toast.success("All styles now use global defaults");
  };

  const restoreDefaultProfiles = async () => {
    const defaults = await AppBridge.GetDefaultProfiles();
    const defaultIds = new Set(defaults.map((d) => d.id));
    const custom = draftProfiles.filter((p) => !defaultIds.has(p.id));
    const updated = [...defaults, ...custom];
    setDraftProfiles(updated);
    const newActive = defaultIds.has(activeId)
      ? activeId
      : updated[0]?.id ?? "";
    setActiveId(newActive);
    saveProfiles(updated, newActive);
    toast.success("Default styles restored");
  };

  const restoreProfile = async (id: string) => {
    const defaults = await AppBridge.GetDefaultProfiles();
    const def = defaults.find((d) => d.id === id);
    if (!def) return;
    const updated = draftProfiles.map((p) => (p.id === id ? def : p));
    setDraftProfiles(updated);
    saveProfiles(updated);
    toast.success(`"${def.name}" restored to default`);
  };

  const restoreBuiltInRules = async () => {
    const rules = await AppBridge.GetDefaultBuiltInRules();
    rulesForm.setValue("builtInRules", rules, { shouldDirty: true });
    toast.success("Built-in rules restored to default");
  };

  const placeholderText =
    "hey um so I was thinking we should probably move the meeting to thursday because like john cant make it on wednesday and I think it would be better if we all met together you know";

  const handleTest = async () => {
    const text = sampleText.trim() || placeholderText;
    const profile = draftProfiles.find((p) => p.id === activeId);
    if (!profile?.prompt) {
      toast.error("Active profile has no prompt");
      return;
    }
    setTesting(true);
    setTestResult("");
    try {
      const res = await AppBridge.TestPrompt(text, profile.prompt);
      if (res.error) toast.error(res.error);
      else setTestResult(res.refined);
    } catch (err) {
      toast.error(`Test failed: ${err}`);
    } finally {
      setTesting(false);
    }
  };

  const presetIds = new Set(["casual", "professional", "concise"]);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <WarningBanner>
            Set up an AI connection first.
          </WarningBanner>
        )}

        {/* Models */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Bot className="size-4" /> Default Models{" "}
              <InfoTip text="Global defaults for transcription and refinement. Styles can override the refinement connection and model." />
            </CardTitle>
            <CardDescription>
              Used unless a style overrides them.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Transcription */}
            <div className="space-y-2">
              <Label className="text-xs font-semibold">Transcription</Label>
              <FormField label={<span className="text-xs">Connection</span>} error={modelsForm.formState.errors.transcriptionConnectionId?.message}>
                <Controller name="transcriptionConnectionId" control={modelsForm.control} render={({ field }) => (
                  <ConnectionSelect value={field.value} onChange={field.onChange} connections={connections} />
                )} />
              </FormField>
              <FormField label={<span className="text-xs">Model</span>}>
                <Controller name="transcriptionModel" control={modelsForm.control} render={({ field }) => (
                  <ModelSelect value={field.value} onChange={field.onChange} models={transcriptionModels} />
                )} />
              </FormField>
            </div>
            <Separator />
            {/* Refinement */}
            <div className="space-y-2">
              <Label className="text-xs font-semibold">Refinement</Label>
              <FormField label={<span className="text-xs">Connection</span>} error={modelsForm.formState.errors.refinementConnectionId?.message}>
                <Controller name="refinementConnectionId" control={modelsForm.control} render={({ field }) => (
                  <ConnectionSelect value={field.value} onChange={field.onChange} connections={connections} />
                )} />
              </FormField>
              <FormField label={<span className="text-xs">Model</span>}>
                <Controller name="refinementModel" control={modelsForm.control} render={({ field }) => (
                  <ModelSelect value={field.value} onChange={field.onChange} models={refChatModels} />
                )} />
              </FormField>
            </div>
            <Button
              size="sm"
              onClick={modelsForm.handleSubmit(saveModels)}
              disabled={!modelsForm.formState.isDirty}
            >
              Save
            </Button>
          </CardContent>
        </Card>

        <Separator />

        {/* Global Rules */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              Global Rules{" "}
              <InfoTip text="Rules applied to every refinement style. Use for preferences like 'don't use em dashes' or 'use British English'." />
            </CardTitle>
            <CardDescription>
              Applied to all styles, after the style prompt.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <textarea
              {...rulesForm.register("globalRules")}
              rows={3}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
              placeholder={
                "- Don't use em dashes\n- Use British English spelling\n- Keep sentences under 20 words"
              }
            />

            <AdvancedToggle
              label="Built-in rules"
              open={globalAdvancedOpen}
              onToggle={setGlobalAdvancedOpen}
            >
              <p className="text-xs text-muted-foreground">
                Core rules always applied. Edit with care.
              </p>
              <textarea
                {...rulesForm.register("builtInRules")}
                rows={6}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-xs leading-relaxed font-mono placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                placeholder="Built-in rules (leave empty for defaults)"
              />
              <ConfirmDialog
                trigger={
                  <Button variant="outline" size="sm" className="text-xs">
                    <RotateCcw className="size-3" /> Restore Default Rules
                  </Button>
                }
                title="Restore built-in rules?"
                description="This will reset the built-in rules to their factory defaults."
                confirmLabel="Restore"
                onConfirm={restoreBuiltInRules}
              />
            </AdvancedToggle>

            <Button
              size="sm"
              onClick={rulesForm.handleSubmit(saveRules)}
              disabled={!rulesForm.formState.isDirty}
            >
              Save
            </Button>
          </CardContent>
        </Card>

        <Separator />

        {/* Profiles */}
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            Refinement Styles{" "}
            <InfoTip text="Each style defines how your speech gets rewritten. Styles can override the connection and model." />
          </h3>
          <div className="flex items-center gap-1.5">
            <ConfirmDialog
              trigger={
                <Button variant="ghost" size="sm">
                  Apply Defaults to All
                </Button>
              }
              title="Apply defaults to all styles?"
              description="This clears connection and model overrides on every style so they all use the global defaults above."
              confirmLabel="Apply"
              variant="default"
              onConfirm={applyDefaultsToAll}
            />
            <ConfirmDialog
              trigger={
                <Button variant="ghost" size="sm">
                  <RotateCcw className="size-3.5" /> Restore Defaults
                </Button>
              }
              title="Restore default styles?"
              description="This will replace the built-in styles with their defaults. Custom styles will be kept."
              confirmLabel="Restore"
              onConfirm={restoreDefaultProfiles}
            />
            <Button variant="outline" size="sm" onClick={addProfile}>
              <Plus className="size-3.5" /> Add
            </Button>
          </div>
        </div>

        {draftProfiles.map((profile) => {
          const Icon = getProfileIcon(profile.icon);
          const isActive = profile.id === activeId;
          const isExpanded = expandedProfile === profile.id;
          const isPreset = presetIds.has(profile.id);

          return (
            <ProfileCard
              key={profile.id}
              profile={profile}
              Icon={Icon}
              isActive={isActive}
              isExpanded={isExpanded}
              isPreset={isPreset}
              connections={connections}
              defaultRefConnId={refConnId}
              defaultRefModel={refModel}
              advancedOpen={advancedOpen === profile.id}
              onToggleExpand={() =>
                setExpandedProfile(isExpanded ? null : profile.id)
              }
              onToggleAdvanced={(v) =>
                setAdvancedOpen(v ? profile.id : null)
              }
              onActivate={() => setActive(profile.id)}
              onSave={(data) => saveProfile(profile.id, data)}
              onRestore={() => restoreProfile(profile.id)}
              onDelete={() => deleteProfile(profile.id)}
            />
          );
        })}

        <Separator />

        {/* Test */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              Test Playground{" "}
              <InfoTip text="Paste sample text to preview how the active style transforms it." />
            </CardTitle>
            <CardDescription>Tests with the active style.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <textarea
              value={sampleText}
              onChange={(e) => setSampleText(e.target.value)}
              rows={3}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
              placeholder={placeholderText}
            />
            <Button
              onClick={handleTest}
              disabled={testing}
              className="w-full"
            >
              {testing ? (
                <>
                  <Loader2 className="size-4 animate-spin" /> Running...
                </>
              ) : (
                <>
                  <Play className="size-4" /> Run Test
                </>
              )}
            </Button>
            {testResult && (
              <div className="rounded-lg bg-muted/50 p-3">
                <p className="text-xs text-muted-foreground mb-1">Result</p>
                <p className="text-sm leading-relaxed whitespace-pre-wrap">
                  {testResult}
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────
// Profile card (extracted for readability)
// ──────────────────────────────────────────────────

function ProfileCard({
  profile,
  Icon,
  isActive,
  isExpanded,
  isPreset,
  connections,
  defaultRefConnId,
  defaultRefModel,
  advancedOpen,
  onToggleExpand,
  onToggleAdvanced,
  onActivate,
  onSave,
  onRestore,
  onDelete,
}: {
  profile: config.RefinementProfile;
  Icon: React.FC<{ className?: string }>;
  isActive: boolean;
  isExpanded: boolean;
  isPreset: boolean;
  connections: config.Connection[];
  defaultRefConnId: string;
  defaultRefModel: string;
  advancedOpen: boolean;
  onToggleExpand: () => void;
  onToggleAdvanced: (v: boolean) => void;
  onActivate: () => void;
  onSave: (data: ProfileFormData) => void;
  onRestore: () => void;
  onDelete: () => void;
}) {
  const {
    register,
    handleSubmit,
    watch,
    reset,
    control,
    formState: { errors, isDirty },
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: profile.name,
      connectionId: profile.connectionId ?? "",
      model: profile.model,
      prompt: profile.prompt,
      temperature: profile.temperature ?? 0.3,
      topP: profile.topP ?? 0.9,
      examples: profile.examples ?? [],
    },
  });

  const { fields: exampleFields, append: appendExample, remove: removeExample } =
    useFieldArray({ control, name: "examples" });

  // Reset when profile changes externally (restore default, etc.)
  useEffect(() => {
    reset({
      name: profile.name,
      connectionId: profile.connectionId ?? "",
      model: profile.model,
      prompt: profile.prompt,
      temperature: profile.temperature ?? 0.3,
      topP: profile.topP ?? 0.9,
      examples: profile.examples ?? [],
    });
  }, [profile, reset]);

  const watchedName = watch("name");
  const watchedConnId = watch("connectionId");
  const watchedModel = watch("model");

  const effectiveConnId = watchedConnId || defaultRefConnId;
  const { chatModels: profileModels } = useModelsForConnection(effectiveConnId);

  const defaultConnName =
    connections.find((c) => c.id === defaultRefConnId)?.name ?? "default";
  const displayModel = watchedModel || defaultRefModel || "default";
  const displayConn = watchedConnId
    ? connections.find((c) => c.id === watchedConnId)?.name ?? "custom"
    : "default";

  return (
    <Card className={isActive ? "border-primary" : ""}>
      <CardHeader className="pb-2">
        <div className="flex items-center gap-3">
          <div
            className={`flex size-8 items-center justify-center rounded-md ${
              isActive
                ? "bg-primary text-primary-foreground"
                : "bg-muted text-muted-foreground"
            }`}
          >
            <Icon className="size-4" />
          </div>
          <div className="flex-1">
            <CardTitle className="flex items-center gap-2 text-sm">
              {watchedName || "Untitled"}
            </CardTitle>
            <CardDescription className="text-xs">
              {displayConn} / {displayModel}
            </CardDescription>
          </div>
          <div className="flex items-center gap-1">
            {!isActive && (
              <Button variant="ghost" size="sm" onClick={onActivate}>
                <Check className="size-3.5" /> Use
              </Button>
            )}
            {isActive && (
              <span className="text-xs font-medium text-primary">active</span>
            )}
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              onClick={onToggleExpand}
            >
              {isExpanded ? (
                <ChevronUp className="size-3.5" />
              ) : (
                <ChevronDown className="size-3.5" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>
      {isExpanded && (
        <CardContent className="space-y-3 pt-0">
          <FormField label="Name" error={errors.name?.message}>
            <Input {...register("name")} />
          </FormField>

          <FormField
            label={<>Connection override <InfoTip text="Use a different AI connection for this style. Leave on default to inherit." /></>}
          >
            <Controller name="connectionId" control={control} render={({ field }) => (
              <ConnectionSelect
                value={field.value ?? ""}
                onChange={field.onChange}
                connections={connections}
                allowDefault
                defaultLabel={`Use default (${defaultConnName})`}
              />
            )} />
          </FormField>

          <FormField
            label={<>Model override <InfoTip text="Override the refinement model for this style." /></>}
          >
            <Controller name="model" control={control} render={({ field }) => (
              <ModelSelect
                value={field.value}
                onChange={field.onChange}
                models={profileModels}
                allowDefault
                defaultLabel={`Use default (${defaultRefModel})`}
              />
            )} />
          </FormField>

          <FormField label="Prompt" error={errors.prompt?.message}>
            <textarea
              {...register("prompt")}
              rows={6}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
            />
          </FormField>

          <AdvancedToggle open={advancedOpen} onToggle={onToggleAdvanced}>
            <div className="space-y-4">
              {/* Temperature */}
              <Controller name="temperature" control={control} render={({ field }) => (
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label className="text-xs flex items-center gap-1">
                      Temperature <InfoTip text="Lower = consistent, higher = creative." />
                    </Label>
                    <Input
                      type="number" min={0} max={1} step={0.05}
                      value={field.value ?? 0.3}
                      onChange={(e) => { const v = parseFloat(e.target.value); if (!isNaN(v) && v >= 0 && v <= 1) field.onChange(v); }}
                      className="h-6 w-16 text-xs tabular-nums px-1.5 text-right"
                    />
                  </div>
                  <Slider min={0} max={1} step={0.05} value={[field.value ?? 0.3]} onValueChange={([v]) => field.onChange(v)} />
                  <div className="flex justify-between text-[10px] text-muted-foreground"><span>Consistent</span><span>Creative</span></div>
                </div>
              )} />

              {/* Top P */}
              <Controller name="topP" control={control} render={({ field }) => (
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label className="text-xs flex items-center gap-1">
                      Top P <InfoTip text="Nucleus sampling threshold." />
                    </Label>
                    <Input
                      type="number" min={0.1} max={1} step={0.05}
                      value={field.value ?? 0.9}
                      onChange={(e) => { const v = parseFloat(e.target.value); if (!isNaN(v) && v >= 0.1 && v <= 1) field.onChange(v); }}
                      className="h-6 w-16 text-xs tabular-nums px-1.5 text-right"
                    />
                  </div>
                  <Slider min={0.1} max={1} step={0.05} value={[field.value ?? 0.9]} onValueChange={([v]) => field.onChange(v)} />
                  <div className="flex justify-between text-[10px] text-muted-foreground"><span>Focused</span><span>Diverse</span></div>
                </div>
              )} />

              {/* Examples */}
              <div className="space-y-2">
                <Label className="text-xs flex items-center gap-1">
                  Examples <InfoTip text="Before/after pairs that anchor the model's style." />
                </Label>
                {exampleFields.map((field, i) => (
                  <div key={field.id} className="space-y-1 rounded border border-border bg-background p-2">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-medium text-muted-foreground">Example {i + 1}</span>
                      <Button variant="ghost" size="icon" className="size-5" onClick={() => removeExample(i)}>
                        <Trash2 className="size-2.5" />
                      </Button>
                    </div>
                    <textarea
                      {...register(`examples.${i}.input`)}
                      rows={2}
                      placeholder="Speech transcript (before)..."
                      className="w-full rounded border border-input bg-background px-2 py-1 text-xs leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                    />
                    <textarea
                      {...register(`examples.${i}.output`)}
                      rows={2}
                      placeholder="Desired output (after)..."
                      className="w-full rounded border border-input bg-background px-2 py-1 text-xs leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                    />
                  </div>
                ))}
                <Button
                  variant="outline" size="sm" className="w-full text-xs"
                  onClick={() => appendExample({ input: "", output: "" })}
                >
                  <Plus className="size-3" /> Add Example
                </Button>
              </div>
            </div>
          </AdvancedToggle>

          <div className="flex items-center gap-2">
            <Button size="sm" onClick={handleSubmit(onSave)} disabled={!isDirty}>
              Save
            </Button>
            {isPreset && (
              <ConfirmDialog
                trigger={<Button variant="outline" size="sm"><RotateCcw className="size-3.5" /> Restore Default</Button>}
                title={`Restore "${watchedName}" to default?`}
                description="This will reset the prompt, examples, and sampling parameters."
                confirmLabel="Restore"
                onConfirm={onRestore}
              />
            )}
            {!isPreset && (
              <ConfirmDialog
                trigger={<Button variant="destructive" size="sm"><Trash2 className="size-3.5" /> Delete</Button>}
                title={`Delete "${watchedName}"?`}
                description="This style will be permanently removed."
                confirmLabel="Delete"
                onConfirm={onDelete}
              />
            )}
          </div>
        </CardContent>
      )}
    </Card>
  );
}
