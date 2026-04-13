import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Bot } from "lucide-react";
import { modelsSchema, type ModelsFormData } from "@/lib/schemas";
import { FormField } from "@/components/ui/form-field";
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
import { InfoTip } from "@/components/info-tip";
import { ConnectionSelect, ModelSelect } from "@/components/profile-card";
import { whisperLanguages } from "@/lib/languages";
import { useModelsForConnection } from "@/lib/hooks";
import { config } from "../../wailsjs/go/models";

interface ModelsViewProps {
  settings: config.Settings;
  configured: boolean;
  onSave: (settings: config.Settings) => void;
}

export function ModelsView({ settings, configured, onSave }: ModelsViewProps) {
  const connections = settings.connections ?? [];

  const form = useForm<ModelsFormData>({
    resolver: zodResolver(modelsSchema),
    defaultValues: {
      transcriptionConnectionId: settings.transcriptionConnectionId,
      transcriptionModel: settings.transcriptionModel,
      transcriptionLanguage: settings.transcriptionLanguage ?? "",
      refinementConnectionId: settings.refinementConnectionId,
      refinementModel: settings.refinementModel,
    },
  });

  const transConnId = form.watch("transcriptionConnectionId");
  const refConnId = form.watch("refinementConnectionId");
  const { transcriptionModels } = useModelsForConnection(transConnId);
  const { chatModels: refChatModels } = useModelsForConnection(refConnId);

  const submit = (data: ModelsFormData) => {
    onSave(config.Settings.createFrom({ ...settings, ...data }));
    form.reset(data);
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <WarningBanner>Set up an AI connection first.</WarningBanner>
        )}

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
            <div className="space-y-2">
              <Label className="text-xs font-semibold">Transcription</Label>
              <FormField
                label={<span className="text-xs">Connection</span>}
                error={
                  form.formState.errors.transcriptionConnectionId?.message
                }
              >
                <Controller
                  name="transcriptionConnectionId"
                  control={form.control}
                  render={({ field }) => (
                    <ConnectionSelect
                      value={field.value}
                      onChange={field.onChange}
                      connections={connections}
                    />
                  )}
                />
              </FormField>
              <FormField label={<span className="text-xs">Model</span>}>
                <Controller
                  name="transcriptionModel"
                  control={form.control}
                  render={({ field }) => (
                    <ModelSelect
                      value={field.value}
                      onChange={field.onChange}
                      models={transcriptionModels}
                    />
                  )}
                />
              </FormField>
              <FormField
                label={
                  <span className="text-xs">
                    Language{" "}
                    <InfoTip text="Tell the model what language you're speaking in. This improves accuracy but will produce gibberish if set to the wrong language. Auto-detect works well for most cases." />
                  </span>
                }
              >
                <Controller
                  name="transcriptionLanguage"
                  control={form.control}
                  render={({ field }) => (
                    <Select
                      value={field.value || "__auto__"}
                      onValueChange={(v) =>
                        field.onChange(v === "__auto__" ? "" : v)
                      }
                    >
                      <SelectTrigger className="text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__auto__" className="text-xs">
                          Auto-detect
                        </SelectItem>
                        <SelectSeparator />
                        {whisperLanguages.map((lang) => (
                          <SelectItem
                            key={lang.code}
                            value={lang.code}
                            className="text-xs"
                          >
                            {lang.flag} {lang.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
              </FormField>
            </div>
            <Separator />
            <div className="space-y-2">
              <Label className="text-xs font-semibold">Refinement</Label>
              <FormField
                label={<span className="text-xs">Connection</span>}
                error={form.formState.errors.refinementConnectionId?.message}
              >
                <Controller
                  name="refinementConnectionId"
                  control={form.control}
                  render={({ field }) => (
                    <ConnectionSelect
                      value={field.value}
                      onChange={field.onChange}
                      connections={connections}
                    />
                  )}
                />
              </FormField>
              <FormField label={<span className="text-xs">Model</span>}>
                <Controller
                  name="refinementModel"
                  control={form.control}
                  render={({ field }) => (
                    <ModelSelect
                      value={field.value}
                      onChange={field.onChange}
                      models={refChatModels}
                    />
                  )}
                />
              </FormField>
            </div>
            <Button
              size="sm"
              onClick={form.handleSubmit(submit)}
              disabled={!form.formState.isDirty}
            >
              Save
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
