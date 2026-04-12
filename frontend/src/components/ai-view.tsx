import { useState, useEffect } from "react";
import {
  Play,
  Loader2,
  Bot,
  AlertTriangle,
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
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";
import type { ai } from "../../wailsjs/go/models";

// ──────────────────────────────────────────────────
// Model selector shared by transcription, refinement, and per-profile
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
  const isCustomValue = value && !knownIds.has(value) && value !== "";

  if (custom || (isCustomValue && value !== "")) {
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
      <Select value={value || "__default__"} onValueChange={(v) => onChange(v === "__default__" ? "" : v)}>
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
  transcriptionModels: ai.ModelInfo[];
  chatModels: ai.ModelInfo[];
}

export function AiView({
  settings,
  configured,
  onSave,
  transcriptionModels,
  chatModels,
}: AiViewProps) {
  const [transModel, setTransModel] = useState(settings.transcriptionModel);
  const [refModel, setRefModel] = useState(settings.refinementModel);
  const [sampleText, setSampleText] = useState("");
  const [testResult, setTestResult] = useState("");
  const [testing, setTesting] = useState(false);
  const [expandedProfile, setExpandedProfile] = useState<string | null>(null);
  const [advancedOpen, setAdvancedOpen] = useState<string | null>(null);
  const [globalAdvancedOpen, setGlobalAdvancedOpen] = useState(false);
  const [draftProfiles, setDraftProfiles] = useState(
    settings.refinementProfiles ?? [],
  );
  const [activeId, setActiveId] = useState(settings.activeProfileId);
  const [globalRules, setGlobalRules] = useState(settings.globalRules ?? "");
  const [builtInRules, setBuiltInRules] = useState(settings.builtInRules ?? "");

  // Show the effective built-in rules when the stored value is empty
  useEffect(() => {
    if (!settings.builtInRules) {
      AppBridge.GetDefaultBuiltInRules().then(setBuiltInRules);
    }
  }, [settings.builtInRules]);

  const saveAll = (
    profiles?: typeof draftProfiles,
    newActiveId?: string,
  ) => {
    const p = profiles ?? draftProfiles;
    const a = newActiveId ?? activeId;
    onSave(
      config.Settings.createFrom({
        ...settings,
        transcriptionModel: transModel,
        refinementModel: refModel,
        refinementProfiles: p,
        activeProfileId: a,
        globalRules,
        builtInRules: builtInRules || undefined,
      }),
    );
  };

  const updateDraftProfile = (
    id: string,
    patch: Partial<config.RefinementProfile>,
  ) => {
    setDraftProfiles((prev) =>
      prev.map((p) =>
        p.id === id
          ? config.RefinementProfile.createFrom({ ...p, ...patch })
          : p,
      ),
    );
  };

  const saveProfile = (_id: string) => {
    saveAll(draftProfiles);
  };

  const deleteProfile = (id: string) => {
    const updated = draftProfiles.filter((p) => p.id !== id);
    const newActive = activeId === id ? updated[0]?.id ?? "" : activeId;
    setDraftProfiles(updated);
    setActiveId(newActive);
    saveAll(updated, newActive);
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
    saveAll(undefined, id);
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
    saveAll(updated, newActive);
    toast.success("Default styles restored");
  };

  const restoreProfile = async (id: string) => {
    const defaults = await AppBridge.GetDefaultProfiles();
    const def = defaults.find((d) => d.id === id);
    if (!def) return;
    const updated = draftProfiles.map((p) => (p.id === id ? def : p));
    setDraftProfiles(updated);
    saveAll(updated);
    toast.success(`"${def.name}" restored to default`);
  };

  const restoreBuiltInRules = async () => {
    const rules = await AppBridge.GetDefaultBuiltInRules();
    setBuiltInRules(rules);
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

  // Check for broken model references
  const allModelIds = new Set([
    ...transcriptionModels.map((m) => m.id),
    ...chatModels.map((m) => m.id),
  ]);
  const hasModels = allModelIds.size > 0;

  const isModelBroken = (modelId: string) => {
    if (!hasModels || !modelId) return false;
    return !allModelIds.has(modelId);
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <WarningBanner>Set up your AI connection first.</WarningBanner>
        )}

        {/* Models */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Bot className="size-4" /> Models{" "}
              <InfoTip text="Select which models to use for transcription and refinement. Models are fetched from your configured AI provider." />
            </CardTitle>
            <CardDescription>
              Global defaults — profiles can override the refinement model.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1">
              <Label>Transcription model</Label>
              <ModelSelect
                value={transModel}
                onChange={setTransModel}
                models={transcriptionModels}
              />
              {isModelBroken(transModel) && (
                <p className="text-xs text-amber-500 flex items-center gap-1 mt-1">
                  <AlertTriangle className="size-3" /> Model not found in provider
                </p>
              )}
            </div>
            <div className="space-y-1">
              <Label>Refinement model</Label>
              <ModelSelect
                value={refModel}
                onChange={setRefModel}
                models={chatModels}
              />
              {isModelBroken(refModel) && (
                <p className="text-xs text-amber-500 flex items-center gap-1 mt-1">
                  <AlertTriangle className="size-3" /> Model not found in provider
                </p>
              )}
            </div>
            <Button size="sm" onClick={() => saveAll()}>
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
              <InfoTip text="Rules applied to every refinement style. Use this for preferences like 'don't use em dashes' or 'use British English' so you don't have to repeat them in each profile." />
            </CardTitle>
            <CardDescription>
              Applied to all styles, appended after the style prompt.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <textarea
              value={globalRules}
              onChange={(e) => setGlobalRules(e.target.value)}
              rows={3}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
              placeholder={
                "- Don't use em dashes\n- Use British English spelling\n- Keep sentences under 20 words"
              }
            />

            {/* Built-in rules (advanced) */}
            <AdvancedToggle
              label="Built-in rules"
              open={globalAdvancedOpen}
              onToggle={setGlobalAdvancedOpen}
            >
              <p className="text-xs text-muted-foreground">
                These core rules are always applied. Edit with care — they
                prevent the model from adding content or dropping meaning.
              </p>
              <textarea
                value={builtInRules}
                onChange={(e) => setBuiltInRules(e.target.value)}
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

            <Button size="sm" onClick={() => saveAll()}>
              Save
            </Button>
          </CardContent>
        </Card>

        <Separator />

        {/* Profiles */}
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            Refinement Styles{" "}
            <InfoTip text="Each style defines how your speech gets rewritten. Choose a style before recording, or set one as active." />
          </h3>
          <div className="flex items-center gap-1.5">
            <ConfirmDialog
              trigger={
                <Button variant="ghost" size="sm">
                  <RotateCcw className="size-3.5" /> Restore Defaults
                </Button>
              }
              title="Restore default styles?"
              description="This will replace the built-in styles (Casual, Professional, Concise) with their defaults. Custom styles will be kept."
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
          const profileModelBroken = isModelBroken(profile.model);
          const displayModel = profile.model || settings.refinementModel || "default";

          return (
            <Card
              key={profile.id}
              className={isActive ? "border-primary" : ""}
            >
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
                      {profile.name}
                      {profileModelBroken && (
                        <AlertTriangle className="size-3 text-amber-500" />
                      )}
                    </CardTitle>
                    <CardDescription className="text-xs">
                      {displayModel}
                    </CardDescription>
                  </div>
                  <div className="flex items-center gap-1">
                    {!isActive && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setActive(profile.id)}
                      >
                        <Check className="size-3.5" /> Use
                      </Button>
                    )}
                    {isActive && (
                      <span className="text-xs font-medium text-primary">
                        active
                      </span>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7"
                      onClick={() =>
                        setExpandedProfile(isExpanded ? null : profile.id)
                      }
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
                  <div className="space-y-1">
                    <Label>Name</Label>
                    <Input
                      value={profile.name}
                      onChange={(e) =>
                        updateDraftProfile(profile.id, {
                          name: e.target.value,
                        })
                      }
                    />
                  </div>
                  <div className="space-y-1">
                    <Label>
                      Model override{" "}
                      <InfoTip text="Override the global refinement model for this style. Leave on 'Use global default' to inherit." />
                    </Label>
                    <ModelSelect
                      value={profile.model}
                      onChange={(v) =>
                        updateDraftProfile(profile.id, { model: v })
                      }
                      models={chatModels}
                      allowDefault
                      defaultLabel={`Use global default (${settings.refinementModel})`}
                    />
                    {profileModelBroken && (
                      <p className="text-xs text-amber-500 flex items-center gap-1 mt-1">
                        <AlertTriangle className="size-3" /> Model not found in
                        provider
                      </p>
                    )}
                  </div>
                  <div className="space-y-1">
                    <Label>Prompt</Label>
                    <textarea
                      value={profile.prompt}
                      onChange={(e) =>
                        updateDraftProfile(profile.id, {
                          prompt: e.target.value,
                        })
                      }
                      rows={6}
                      className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                    />
                  </div>

                  {/* Advanced Options */}
                  <AdvancedToggle
                    open={advancedOpen === profile.id}
                    onToggle={(v) => setAdvancedOpen(v ? profile.id : null)}
                  >
                    <div className="space-y-4">
                      {/* Temperature */}
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <Label className="text-xs flex items-center gap-1">
                            Temperature{" "}
                            <InfoTip text="Controls randomness. Lower values produce more consistent output; higher values add variety." />
                          </Label>
                          <Input
                            type="number"
                            min={0}
                            max={1}
                            step={0.05}
                            value={profile.temperature ?? 0.3}
                            onChange={(e) => {
                              const v = parseFloat(e.target.value);
                              if (!isNaN(v) && v >= 0 && v <= 1)
                                updateDraftProfile(profile.id, {
                                  temperature: v,
                                });
                            }}
                            className="h-6 w-16 text-xs tabular-nums px-1.5 text-right"
                          />
                        </div>
                        <Slider
                          min={0}
                          max={1}
                          step={0.05}
                          value={[profile.temperature ?? 0.3]}
                          onValueChange={([v]) =>
                            updateDraftProfile(profile.id, { temperature: v })
                          }
                        />
                        <div className="flex justify-between text-[10px] text-muted-foreground">
                          <span>Consistent</span>
                          <span>Creative</span>
                        </div>
                      </div>

                      {/* Top P */}
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <Label className="text-xs flex items-center gap-1">
                            Top P{" "}
                            <InfoTip text="Nucleus sampling. Limits token selection to the most probable tokens whose cumulative probability reaches this threshold." />
                          </Label>
                          <Input
                            type="number"
                            min={0.1}
                            max={1}
                            step={0.05}
                            value={profile.topP ?? 0.9}
                            onChange={(e) => {
                              const v = parseFloat(e.target.value);
                              if (!isNaN(v) && v >= 0.1 && v <= 1)
                                updateDraftProfile(profile.id, { topP: v });
                            }}
                            className="h-6 w-16 text-xs tabular-nums px-1.5 text-right"
                          />
                        </div>
                        <Slider
                          min={0.1}
                          max={1}
                          step={0.05}
                          value={[profile.topP ?? 0.9]}
                          onValueChange={([v]) =>
                            updateDraftProfile(profile.id, { topP: v })
                          }
                        />
                        <div className="flex justify-between text-[10px] text-muted-foreground">
                          <span>Focused</span>
                          <span>Diverse</span>
                        </div>
                      </div>

                      {/* Few-shot Examples */}
                      <div className="space-y-2">
                        <Label className="text-xs flex items-center gap-1">
                          Examples{" "}
                          <InfoTip text="Before/after pairs that teach the model your preferred style. These are sent as conversation history before your transcript." />
                        </Label>
                        {(profile.examples ?? []).map((ex, i) => (
                          <div
                            key={i}
                            className="space-y-1 rounded border border-border bg-background p-2"
                          >
                            <div className="flex items-center justify-between">
                              <span className="text-[10px] font-medium text-muted-foreground">
                                Example {i + 1}
                              </span>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="size-5"
                                onClick={() => {
                                  const updated = [
                                    ...(profile.examples ?? []),
                                  ];
                                  updated.splice(i, 1);
                                  updateDraftProfile(profile.id, {
                                    examples: updated,
                                  });
                                }}
                              >
                                <Trash2 className="size-2.5" />
                              </Button>
                            </div>
                            <textarea
                              value={ex.input}
                              onChange={(e) => {
                                const updated = [
                                  ...(profile.examples ?? []),
                                ];
                                updated[i] = {
                                  ...updated[i],
                                  input: e.target.value,
                                };
                                updateDraftProfile(profile.id, {
                                  examples: updated,
                                });
                              }}
                              rows={2}
                              placeholder="Speech transcript (before)..."
                              className="w-full rounded border border-input bg-background px-2 py-1 text-xs leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                            />
                            <textarea
                              value={ex.output}
                              onChange={(e) => {
                                const updated = [
                                  ...(profile.examples ?? []),
                                ];
                                updated[i] = {
                                  ...updated[i],
                                  output: e.target.value,
                                };
                                updateDraftProfile(profile.id, {
                                  examples: updated,
                                });
                              }}
                              rows={2}
                              placeholder="Desired output (after)..."
                              className="w-full rounded border border-input bg-background px-2 py-1 text-xs leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                            />
                          </div>
                        ))}
                        <Button
                          variant="outline"
                          size="sm"
                          className="w-full text-xs"
                          onClick={() => {
                            const updated = [
                              ...(profile.examples ?? []),
                              { input: "", output: "" },
                            ];
                            updateDraftProfile(profile.id, {
                              examples: updated,
                            });
                          }}
                        >
                          <Plus className="size-3" /> Add Example
                        </Button>
                      </div>
                    </div>
                  </AdvancedToggle>

                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      onClick={() => saveProfile(profile.id)}
                    >
                      Save
                    </Button>
                    {isPreset && (
                      <ConfirmDialog
                        trigger={
                          <Button variant="outline" size="sm">
                            <RotateCcw className="size-3.5" /> Restore Default
                          </Button>
                        }
                        title={`Restore "${profile.name}" to default?`}
                        description="This will reset the prompt, examples, and sampling parameters to their defaults."
                        confirmLabel="Restore"
                        onConfirm={() => restoreProfile(profile.id)}
                      />
                    )}
                    {!isPreset && (
                      <ConfirmDialog
                        trigger={
                          <Button variant="destructive" size="sm">
                            <Trash2 className="size-3.5" /> Delete
                          </Button>
                        }
                        title={`Delete "${profile.name}"?`}
                        description="This style will be permanently removed."
                        confirmLabel="Delete"
                        onConfirm={() => deleteProfile(profile.id)}
                      />
                    )}
                  </div>
                </CardContent>
              )}
            </Card>
          );
        })}

        <Separator />

        {/* Test */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              Test Playground{" "}
              <InfoTip text="Paste sample text to preview how the active style transforms it, without recording audio." />
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
