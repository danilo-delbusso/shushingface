import { useState, useEffect } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  modelsSchema,
  globalRulesSchema,
  type ModelsFormData,
  type GlobalRulesFormData,
  type ProfileFormData,
} from "@/lib/schemas";
import { FormField } from "@/components/ui/form-field";
import {
  Play,
  Loader2,
  Bot,
  Plus,
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
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { WarningBanner } from "@/components/ui/warning-banner";
import { AdvancedToggle } from "@/components/ui/advanced-toggle";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
import { ConnectionSelect, ModelSelect, ProfileCard } from "@/components/profile-card";
import { getProfileIcon } from "@/lib/icons";
import { whisperLanguages } from "@/lib/languages";
import { cn, textareaClass, textareaCompactClass } from "@/lib/utils";
import { useModelsForConnection } from "@/lib/hooks";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";

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
      transcriptionLanguage: settings.transcriptionLanguage ?? "",
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
    modelsForm.reset(data);
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
    rulesForm.reset(data);
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
              <FormField
                label={
                  <span className="text-xs">
                    Language{" "}
                    <InfoTip text="Tell the model what language you're speaking in. This improves accuracy but will produce gibberish if set to the wrong language. Auto-detect works well for most cases." />
                  </span>
                }
              >
                <Controller name="transcriptionLanguage" control={modelsForm.control} render={({ field }) => (
                  <Select value={field.value || "__auto__"} onValueChange={(v) => field.onChange(v === "__auto__" ? "" : v)}>
                    <SelectTrigger className="text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__auto__" className="text-xs">
                        Auto-detect
                      </SelectItem>
                      <SelectSeparator />
                      {whisperLanguages.map((lang) => (
                        <SelectItem key={lang.code} value={lang.code} className="text-xs">
                          {lang.flag} {lang.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
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
              className={textareaClass}
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
                className={cn(textareaCompactClass, "font-mono")}
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
              className={textareaClass}
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

