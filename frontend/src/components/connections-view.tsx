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
  // Draft state — only saved on explicit Save
  const [name, setName] = useState(conn.name);
  const [providerId, setProviderId] = useState(conn.providerId);
  const [apiKey, setApiKey] = useState(conn.apiKey);
  const [baseUrl, setBaseUrl] = useState(conn.baseUrl ?? "");
  const [showKey, setShowKey] = useState(false);
  const [testing, setTesting] = useState(false);

  const preset = providerPresets[providerId];
  const needsBaseUrl = preset?.requiresBaseUrl ?? false;
  const [advOpen, setAdvOpen] = useState(!!baseUrl || needsBaseUrl);

  // Sync draft when the saved connection changes (e.g. after add)
  useEffect(() => {
    setName(conn.name);
    setProviderId(conn.providerId);
    setApiKey(conn.apiKey);
    setBaseUrl(conn.baseUrl ?? "");
  }, [conn.id, conn.name, conn.providerId, conn.apiKey, conn.baseUrl]);

  const save = () => {
    onSave(
      config.Connection.createFrom({
        id: conn.id,
        name,
        providerId,
        apiKey,
        baseUrl: baseUrl || undefined,
      }),
    );
  };

  const testConnection = async () => {
    // Save first so backend has the latest credentials
    save();
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
              <span className="text-xs font-bold">{name[0]}</span>
            )}
          </div>
          <div className="flex-1 min-w-0">
            <CardTitle className="flex items-center gap-2 text-sm">
              {name}
              {!apiKey && !needsBaseUrl && (
                <AlertTriangle className="size-3 text-amber-500 shrink-0" />
              )}
            </CardTitle>
            <CardDescription className="text-xs">
              {preset?.name ?? providerId}
              {apiKey ? " — connected" : needsBaseUrl ? "" : " — needs API key"}
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
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-1">
            <Label>Provider</Label>
            <Select value={providerId} onValueChange={setProviderId}>
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
          {needsBaseUrl && (
            <div className="space-y-1">
              <Label>
                Base URL{" "}
                <InfoTip text="The API endpoint, e.g. http://localhost:11434/v1 for Ollama or https://api.openai.com/v1 for OpenAI." />
              </Label>
              <Input
                value={baseUrl}
                placeholder="http://localhost:11434/v1"
                onChange={(e) => setBaseUrl(e.target.value)}
              />
            </div>
          )}

          <div className="space-y-1">
            <Label className="flex items-center gap-2">
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
            </Label>
            <div className="flex">
              <Input
                type={showKey ? "text" : "password"}
                value={apiKey}
                placeholder={preset?.keyPlaceholder ?? "API key..."}
                onChange={(e) => setApiKey(e.target.value)}
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
          </div>

          {/* Base URL override for known providers (hidden behind Advanced) */}
          {!needsBaseUrl && (
            <AdvancedToggle open={advOpen} onToggle={setAdvOpen}>
              <div className="space-y-1">
                <Label className="text-xs flex items-center gap-1">
                  Base URL{" "}
                  <InfoTip text="Override the default API endpoint for self-hosted or proxy setups." />
                </Label>
                <Input
                  value={baseUrl}
                  placeholder="Leave empty for default"
                  onChange={(e) => setBaseUrl(e.target.value)}
                  className="text-xs"
                />
              </div>
            </AdvancedToggle>
          )}

          <div className="flex items-center gap-2">
            <Button size="sm" onClick={save}>
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
              title={`Delete "${name}"?`}
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
