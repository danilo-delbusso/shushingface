import { useState, useEffect } from "react";
import { useForm, Controller, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { profileSchema, type ProfileFormData } from "@/lib/schemas";
import { FormField } from "@/components/ui/form-field";
import {
  Check,
  Trash2,
  Plus,
  ChevronDown,
  ChevronUp,
  RotateCcw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { AdvancedToggle } from "@/components/ui/advanced-toggle";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
import { textareaClass, textareaCompactClass } from "@/lib/utils";
import { useModelsForConnection } from "@/lib/hooks";
import type { config, ai } from "../../wailsjs/go/models";

export function ConnectionSelect({
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

export function ModelSelect({
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

export function ProfileCard({
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

  const {
    fields: exampleFields,
    append: appendExample,
    remove: removeExample,
  } = useFieldArray({ control, name: "examples" });

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
    ? (connections.find((c) => c.id === watchedConnId)?.name ?? "custom")
    : "default";

  return (
    <Card className={isActive ? "border-primary" : ""}>
      <CardContent className="pb-2 pt-4">
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
          <div className="flex-1 min-w-0">
            <CardTitle className="flex items-center gap-2 text-sm">
              {watchedName || "Untitled"}
            </CardTitle>
            {isExpanded && (
              <CardDescription className="text-xs truncate">
                {displayConn} / {displayModel}
              </CardDescription>
            )}
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
      </CardContent>
      {isExpanded && (
        <CardContent className="space-y-3 pt-0">
          <FormField label="Name" error={errors.name?.message}>
            <Input {...register("name")} />
          </FormField>

          <FormField
            label={
              <>
                Connection override{" "}
                <InfoTip text="Use a different AI connection for this style. Leave on default to inherit." />
              </>
            }
          >
            <Controller
              name="connectionId"
              control={control}
              render={({ field }) => (
                <ConnectionSelect
                  value={field.value ?? ""}
                  onChange={field.onChange}
                  connections={connections}
                  allowDefault
                  defaultLabel={`Use default (${defaultConnName})`}
                />
              )}
            />
          </FormField>

          <FormField
            label={
              <>
                Model override{" "}
                <InfoTip text="Override the refinement model for this style." />
              </>
            }
          >
            <Controller
              name="model"
              control={control}
              render={({ field }) => (
                <ModelSelect
                  value={field.value}
                  onChange={field.onChange}
                  models={profileModels}
                  allowDefault
                  defaultLabel={`Use default (${defaultRefModel})`}
                />
              )}
            />
          </FormField>

          <FormField label="Prompt" error={errors.prompt?.message}>
            <textarea
              {...register("prompt")}
              rows={6}
              className={textareaClass}
            />
          </FormField>

          <AdvancedToggle open={advancedOpen} onToggle={onToggleAdvanced}>
            <div className="space-y-4">
              {/* Temperature */}
              <Controller
                name="temperature"
                control={control}
                render={({ field }) => (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs flex items-center gap-1">
                        Temperature{" "}
                        <InfoTip text="Lower = consistent, higher = creative." />
                      </Label>
                      <Input
                        type="number"
                        min={0}
                        max={1}
                        step={0.05}
                        value={field.value ?? 0.3}
                        onChange={(e) => {
                          const v = parseFloat(e.target.value);
                          if (!Number.isNaN(v) && v >= 0 && v <= 1)
                            field.onChange(v);
                        }}
                        className="h-6 w-16 text-xs tabular-nums px-1.5 text-right"
                      />
                    </div>
                    <Slider
                      min={0}
                      max={1}
                      step={0.05}
                      value={[field.value ?? 0.3]}
                      onValueChange={([v]) => field.onChange(v)}
                    />
                    <div className="flex justify-between text-[10px] text-muted-foreground">
                      <span>Consistent</span>
                      <span>Creative</span>
                    </div>
                  </div>
                )}
              />

              {/* Top P */}
              <Controller
                name="topP"
                control={control}
                render={({ field }) => (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs flex items-center gap-1">
                        Top P <InfoTip text="Nucleus sampling threshold." />
                      </Label>
                      <Input
                        type="number"
                        min={0.1}
                        max={1}
                        step={0.05}
                        value={field.value ?? 0.9}
                        onChange={(e) => {
                          const v = parseFloat(e.target.value);
                          if (!Number.isNaN(v) && v >= 0.1 && v <= 1)
                            field.onChange(v);
                        }}
                        className="h-6 w-16 text-xs tabular-nums px-1.5 text-right"
                      />
                    </div>
                    <Slider
                      min={0.1}
                      max={1}
                      step={0.05}
                      value={[field.value ?? 0.9]}
                      onValueChange={([v]) => field.onChange(v)}
                    />
                    <div className="flex justify-between text-[10px] text-muted-foreground">
                      <span>Focused</span>
                      <span>Diverse</span>
                    </div>
                  </div>
                )}
              />

              {/* Examples */}
              <div className="space-y-2">
                <Label className="text-xs flex items-center gap-1">
                  Examples{" "}
                  <InfoTip text="Before/after pairs that anchor the model's style." />
                </Label>
                {exampleFields.map((field, i) => (
                  <div
                    key={field.id}
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
                        onClick={() => removeExample(i)}
                      >
                        <Trash2 className="size-2.5" />
                      </Button>
                    </div>
                    <textarea
                      {...register(`examples.${i}.input`)}
                      rows={4}
                      placeholder="Speech transcript (before)..."
                      className={textareaCompactClass}
                    />
                    <textarea
                      {...register(`examples.${i}.output`)}
                      rows={4}
                      placeholder="Desired output (after)..."
                      className={textareaCompactClass}
                    />
                  </div>
                ))}
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full text-xs"
                  onClick={() => appendExample({ input: "", output: "" })}
                >
                  <Plus className="size-3" /> Add Example
                </Button>
              </div>
            </div>
          </AdvancedToggle>

          <div className="flex items-center gap-2">
            <Button
              size="sm"
              onClick={handleSubmit((data) => {
                onSave(data);
                reset(data);
              })}
              disabled={!isDirty}
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
                title={`Restore "${watchedName}" to default?`}
                description="This will reset the prompt, examples, and sampling parameters."
                confirmLabel="Restore"
                onConfirm={onRestore}
              />
            )}
            {!isPreset && (
              <ConfirmDialog
                trigger={
                  <Button variant="destructive" size="sm">
                    <Trash2 className="size-3.5" /> Delete
                  </Button>
                }
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
