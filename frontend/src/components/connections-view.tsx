import { useState, useEffect } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  connectionSchema,
  type ConnectionFormData,
} from "@/lib/schemas";
import { FormField } from "@/components/ui/form-field";
import {
  Plug,
  Loader2,
  RefreshCw,
  Plus,
  Trash2,
  ChevronDown,
  ChevronUp,
  AlertTriangle,
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
import { PasswordInput } from "@/components/ui/password-input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { WarningBanner } from "@/components/ui/warning-banner";
import { AdvancedToggle } from "@/components/ui/advanced-toggle";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
import { providerPresets } from "@/lib/providers";
import { ExternalLink } from "@/components/ui/external-link";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";
import type { ai } from "../../wailsjs/go/models";

interface ConnectionsViewProps {
  settings: config.Settings;
  configured: boolean;
  onSave: (settings: config.Settings) => void;
}

export function ConnectionsView({
  settings,
  configured,
  onSave,
}: ConnectionsViewProps) {
  const [providers, setProviders] = useState<ai.ProviderInfo[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  useEffect(() => {
    AppBridge.ListProviders().then(setProviders);
  }, []);

  const connections = settings.connections ?? [];

  const saveConnections = (updated: config.Connection[]) => {
    const patch: Partial<config.Settings> = { connections: updated };
    // Auto-assign first connection as default if none set
    if (
      updated.length > 0 &&
      !updated.some((c) => c.id === settings.transcriptionConnectionId)
    ) {
      patch.transcriptionConnectionId = updated[0].id;
    }
    if (
      updated.length > 0 &&
      !updated.some((c) => c.id === settings.refinementConnectionId)
    ) {
      patch.refinementConnectionId = updated[0].id;
    }
    onSave(config.Settings.createFrom({ ...settings, ...patch }));
  };

  const saveConnection = (updated: config.Connection) => {
    const list = connections.map((c) => (c.id === updated.id ? updated : c));
    saveConnections(list);
  };

  const addConnection = () => {
    const id = `conn_${Date.now()}`;
    const provId = providers[0]?.id ?? "groq";
    const preset = providerPresets[provId];
    const conn = config.Connection.createFrom({
      id,
      name: preset?.name ?? "New Connection",
      providerId: provId,
      apiKey: "",
    });
    const updated = [...connections, conn];
    const patch: Partial<config.Settings> = { connections: updated };
    if (connections.length === 0) {
      patch.transcriptionConnectionId = id;
      patch.refinementConnectionId = id;
    }
    onSave(config.Settings.createFrom({ ...settings, ...patch }));
    setExpandedId(id);
  };

  const deleteConnection = (id: string) => {
    const updated = connections.filter((c) => c.id !== id);
    const patch: Partial<config.Settings> = { connections: updated };
    if (settings.transcriptionConnectionId === id) {
      patch.transcriptionConnectionId = updated[0]?.id ?? "";
    }
    if (settings.refinementConnectionId === id) {
      patch.refinementConnectionId = updated[0]?.id ?? "";
    }
    onSave(config.Settings.createFrom({ ...settings, ...patch }));
  };

  const isInUse = (id: string) => {
    if (settings.transcriptionConnectionId === id) return true;
    if (settings.refinementConnectionId === id) return true;
    return (settings.refinementProfiles ?? []).some(
      (p) => p.connectionId === id,
    );
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <WarningBanner>
            Add a connection and configure your API key to get started.
          </WarningBanner>
        )}

        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            <Plug className="size-4" /> Connections{" "}
            <InfoTip text="Named AI provider connections. You can add multiple and assign them independently to transcription, refinement, or individual styles." />
          </h3>
          <Button variant="outline" size="sm" onClick={addConnection}>
            <Plus className="size-3.5" /> Add
          </Button>
        </div>

        {connections.length === 0 && (
          <Card>
            <CardContent className="py-8 text-center text-sm text-muted-foreground">
              No connections yet. Add one to get started.
            </CardContent>
          </Card>
        )}

        {connections.map((conn) => (
          <ConnectionCard
            key={conn.id}
            conn={conn}
            providers={providers}
            isExpanded={expandedId === conn.id}
            onToggleExpand={() =>
              setExpandedId(expandedId === conn.id ? null : conn.id)
            }
            onSave={saveConnection}
            onDelete={() => deleteConnection(conn.id)}
            inUse={isInUse(conn.id)}
          />
        ))}
      </div>
    </div>
  );
}

function ConnectionCard({
  conn,
  providers,
  isExpanded,
  onToggleExpand,
  onSave,
  onDelete,
  inUse,
}: {
  conn: config.Connection;
  providers: ai.ProviderInfo[];
  isExpanded: boolean;
  onToggleExpand: () => void;
  onSave: (updated: config.Connection) => void;
  onDelete: () => void;
  inUse: boolean;
}) {
  const {
    register,
    handleSubmit,
    watch,
    reset,
    control,
    formState: { errors, isDirty },
  } = useForm<ConnectionFormData>({
    resolver: zodResolver(connectionSchema),
    defaultValues: {
      name: conn.name,
      providerId: conn.providerId,
      apiKey: conn.apiKey,
      baseUrl: conn.baseUrl ?? "",
    },
  });

  const [testing, setTesting] = useState(false);

  const watchedName = watch("name");
  const watchedProviderId = watch("providerId");
  const watchedApiKey = watch("apiKey");
  const preset = providerPresets[watchedProviderId];
  const needsBaseUrl = preset?.requiresBaseUrl ?? false;
  const [advOpen, setAdvOpen] = useState(
    !!conn.baseUrl || needsBaseUrl,
  );

  // Reset form when saved connection changes externally
  useEffect(() => {
    reset({
      name: conn.name,
      providerId: conn.providerId,
      apiKey: conn.apiKey,
      baseUrl: conn.baseUrl ?? "",
    });
  }, [conn, reset]);

  const onSubmit = (data: ConnectionFormData) => {
    onSave(
      config.Connection.createFrom({
        id: conn.id,
        ...data,
        baseUrl: data.baseUrl || undefined,
      }),
    );
    reset(data);
  };

  const testConnection = async () => {
    handleSubmit(onSubmit)();
    setTesting(true);
    try {
      const models = await AppBridge.ListModelsForConnection(conn.id);
      toast.success(`Connected — ${models?.length ?? 0} models available`);
    } catch (err) {
      toast.error(`Connection failed: ${err}`);
    } finally {
      setTesting(false);
    }
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center gap-3">
          <div className="flex size-8 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
            {preset?.icon ? (
              <img src={preset.icon} alt="" className="size-4" />
            ) : (
              <span className="text-xs font-bold">
                {watchedName?.[0] ?? "?"}
              </span>
            )}
          </div>
          <div className="flex-1 min-w-0">
            <CardTitle className="flex items-center gap-2 text-sm">
              {watchedName || "Untitled"}
              {!watchedApiKey && !needsBaseUrl && (
                <AlertTriangle className="size-3 text-amber-500 shrink-0" />
              )}
            </CardTitle>
            <CardDescription className="text-xs">
              {preset?.name ?? watchedProviderId}
              {watchedApiKey
                ? " — connected"
                : needsBaseUrl
                  ? ""
                  : " — needs API key"}
            </CardDescription>
          </div>
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
      </CardHeader>
      {isExpanded && (
        <CardContent className="space-y-4 pt-0">
          <FormField label="Name" error={errors.name?.message}>
            <Input {...register("name")} />
          </FormField>

          <FormField label="Provider">
            <Controller
              name="providerId"
              control={control}
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {providers.map((p) => (
                      <SelectItem key={p.id} value={p.id}>
                        {p.displayName}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </FormField>

          {/* Base URL — shown inline for providers that need it */}
          {needsBaseUrl && (
            <FormField
              label={
                <>
                  Base URL{" "}
                  <InfoTip text="The API endpoint, e.g. http://localhost:11434/v1 for Ollama or https://api.openai.com/v1 for OpenAI." />
                </>
              }
              error={errors.baseUrl?.message}
            >
              <Input
                {...register("baseUrl")}
                placeholder="http://localhost:11434/v1"
              />
            </FormField>
          )}

          <FormField
            label={
              <span className="flex items-center gap-2">
                API Key
                {needsBaseUrl && (
                  <span className="text-xs text-muted-foreground font-normal">
                    optional for local
                  </span>
                )}
                {preset?.keyUrl && (
                  <ExternalLink
                    href={preset.keyUrl}
                    className="text-xs font-normal"
                  >
                    {preset.keyUrlLabel}
                  </ExternalLink>
                )}
              </span>
            }
            error={errors.apiKey?.message}
          >
            <PasswordInput
              {...register("apiKey")}
              placeholder={preset?.keyPlaceholder ?? "API key..."}
            />
          </FormField>

          {/* Base URL override for known providers (hidden behind Advanced) */}
          {!needsBaseUrl && (
            <AdvancedToggle open={advOpen} onToggle={setAdvOpen}>
              <FormField
                label={
                  <span className="text-xs flex items-center gap-1">
                    Base URL{" "}
                    <InfoTip text="Override the default API endpoint for self-hosted or proxy setups." />
                  </span>
                }
              >
                <Input
                  {...register("baseUrl")}
                  placeholder="Leave empty for default"
                  className="text-xs"
                />
              </FormField>
            </AdvancedToggle>
          )}

          <div className="flex items-center gap-2">
            <Button
              size="sm"
              onClick={handleSubmit(onSubmit)}
              disabled={!isDirty}
            >
              Save
            </Button>
            <Button
              size="sm"
              variant="outline"
              disabled={testing}
              onClick={testConnection}
            >
              {testing ? (
                <>
                  <Loader2 className="size-3.5 animate-spin" /> Testing...
                </>
              ) : (
                <>
                  <RefreshCw className="size-3.5" /> Test
                </>
              )}
            </Button>
            <ConfirmDialog
              trigger={
                <Button variant="destructive" size="sm">
                  <Trash2 className="size-3.5" /> Delete
                </Button>
              }
              title={`Delete "${watchedName}"?`}
              description={
                inUse
                  ? "This connection is currently in use by transcription, refinement, or a style. Deleting it may break things."
                  : "This connection will be permanently removed."
              }
              confirmLabel="Delete"
              onConfirm={onDelete}
            />
          </div>
        </CardContent>
      )}
    </Card>
  );
}
