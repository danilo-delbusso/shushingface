import { useState, useEffect } from "react";
import {
  Eye,
  EyeOff,
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
import { Label } from "@/components/ui/label";
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
  const [connections, setConnections] = useState(settings.connections ?? []);

  useEffect(() => {
    AppBridge.ListProviders().then(setProviders);
  }, []);

  const save = (updated: config.Connection[]) => {
    setConnections(updated);
    onSave(
      config.Settings.createFrom({
        ...settings,
        connections: updated,
      }),
    );
  };

  const updateConnection = (id: string, patch: Partial<config.Connection>) => {
    const updated = connections.map((c) =>
      c.id === id ? config.Connection.createFrom({ ...c, ...patch }) : c,
    );
    save(updated);
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
    setConnections(updated);
    setExpandedId(id);

    if (connections.length === 0) {
      onSave(
        config.Settings.createFrom({
          ...settings,
          connections: updated,
          transcriptionConnectionId: id,
          refinementConnectionId: id,
        }),
      );
    } else {
      save(updated);
    }
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
    setConnections(updated);
    onSave(config.Settings.createFrom({ ...settings, ...patch }));
  };

  const isInUse = (id: string) => {
    if (settings.transcriptionConnectionId === id) return true;
    if (settings.refinementConnectionId === id) return true;
    return (settings.refinementProfiles ?? []).some((p) => p.connectionId === id);
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
            onUpdate={(patch) => updateConnection(conn.id, patch)}
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
  onUpdate,
  onDelete,
  inUse,
}: {
  conn: config.Connection;
  providers: ai.ProviderInfo[];
  isExpanded: boolean;
  onToggleExpand: () => void;
  onUpdate: (patch: Partial<config.Connection>) => void;
  onDelete: () => void;
  inUse: boolean;
}) {
  const preset = providerPresets[conn.providerId];
  const [showKey, setShowKey] = useState(false);
  const [testing, setTesting] = useState(false);
  const [advOpen, setAdvOpen] = useState(!!conn.baseUrl || !!preset?.requiresBaseUrl);

  const testConnection = async () => {
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
              <span className="text-xs font-bold">{conn.name[0]}</span>
            )}
          </div>
          <div className="flex-1 min-w-0">
            <CardTitle className="flex items-center gap-2 text-sm">
              {conn.name}
              {!conn.apiKey && (
                <AlertTriangle className="size-3 text-amber-500 shrink-0" />
              )}
            </CardTitle>
            <CardDescription className="text-xs">
              {preset?.name ?? conn.providerId}
              {conn.apiKey ? " — connected" : " — needs API key"}
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
          <div className="space-y-1">
            <Label>Name</Label>
            <Input
              value={conn.name}
              onChange={(e) => onUpdate({ name: e.target.value })}
            />
          </div>
          <div className="space-y-1">
            <Label>Provider</Label>
            <Select
              value={conn.providerId}
              onValueChange={(v) => onUpdate({ providerId: v })}
            >
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
          </div>
          {/* Base URL — shown inline for providers that need it */}
          {preset?.requiresBaseUrl && (
            <div className="space-y-1">
              <Label>
                Base URL{" "}
                <InfoTip text="The API endpoint, e.g. http://localhost:11434/v1 for Ollama or https://api.openai.com/v1 for OpenAI." />
              </Label>
              <Input
                value={conn.baseUrl ?? ""}
                placeholder="http://localhost:11434/v1"
                onChange={(e) =>
                  onUpdate({ baseUrl: e.target.value || undefined })
                }
              />
            </div>
          )}

          <div className="space-y-1">
            <Label className="flex items-center gap-2">
              API Key
              {preset?.keyUrl && (
                <ExternalLink
                  href={preset.keyUrl}
                  className="text-xs font-normal"
                >
                  {preset.keyUrlLabel}
                </ExternalLink>
              )}
            </Label>
            <div className="flex">
              <Input
                type={showKey ? "text" : "password"}
                value={conn.apiKey}
                placeholder={preset?.keyPlaceholder ?? "API key..."}
                onChange={(e) => onUpdate({ apiKey: e.target.value })}
                className="rounded-r-none"
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={() => setShowKey(!showKey)}
                className="rounded-l-none border-l-0"
              >
                {showKey ? (
                  <EyeOff className="size-4" />
                ) : (
                  <Eye className="size-4" />
                )}
              </Button>
            </div>
            {preset?.requiresBaseUrl && (
              <p className="text-xs text-muted-foreground">
                Optional for local providers like Ollama.
              </p>
            )}
          </div>

          {/* Base URL override for known providers (hidden behind Advanced) */}
          {!preset?.requiresBaseUrl && (
            <AdvancedToggle open={advOpen} onToggle={setAdvOpen}>
              <div className="space-y-1">
                <Label className="text-xs flex items-center gap-1">
                  Base URL{" "}
                  <InfoTip text="Override the default API endpoint for self-hosted or proxy setups." />
                </Label>
                <Input
                  value={conn.baseUrl ?? ""}
                  placeholder="Leave empty for default"
                  onChange={(e) =>
                    onUpdate({ baseUrl: e.target.value || undefined })
                  }
                  className="text-xs"
                />
              </div>
            </AdvancedToggle>
          )}

          <div className="flex items-center gap-2">
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
              title={`Delete "${conn.name}"?`}
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
