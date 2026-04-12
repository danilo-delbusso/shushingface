import { useState, useEffect } from "react";
import {
  Eye,
  EyeOff,
  Plug,
  Loader2,
  RefreshCw,
  Check,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { WarningBanner } from "@/components/ui/warning-banner";
import { AdvancedToggle } from "@/components/ui/advanced-toggle";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { InfoTip } from "@/components/info-tip";
import { providerPresets } from "@/lib/providers";
import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";
import type { ai } from "../../wailsjs/go/models";

interface ConnectionsViewProps {
  settings: config.Settings;
  configured: boolean;
  onSave: (settings: config.Settings) => void;
  onModelsRefreshed?: () => void;
}

export function ConnectionsView({
  settings,
  configured,
  onSave,
  onModelsRefreshed,
}: ConnectionsViewProps) {
  const [providers, setProviders] = useState<ai.ProviderInfo[]>([]);
  const [providerId, setProviderId] = useState(settings.providerId);
  const [apiKey, setApiKey] = useState(settings.providerApiKey ?? "");
  const [baseUrl, setBaseUrl] = useState(settings.providerBaseUrl ?? "");
  const [showKey, setShowKey] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(!!settings.providerBaseUrl);
  const [testing, setTesting] = useState(false);
  const [modelCount, setModelCount] = useState<number | null>(null);

  useEffect(() => {
    AppBridge.ListProviders().then(setProviders);
  }, []);

  const preset = providerPresets[providerId];

  const save = () => {
    onSave(
      config.Settings.createFrom({
        ...settings,
        providerId,
        providerApiKey: apiKey,
        providerBaseUrl: baseUrl || undefined,
      }),
    );
  };

  const testConnection = async () => {
    onSave(
      config.Settings.createFrom({
        ...settings,
        providerId,
        providerApiKey: apiKey,
        providerBaseUrl: baseUrl || undefined,
      }),
    );
    setTesting(true);
    setModelCount(null);
    try {
      const models = await AppBridge.ListModels();
      setModelCount(models?.length ?? 0);
      toast.success(`Connected — ${models?.length ?? 0} models available`);
      onModelsRefreshed?.();
    } catch (err) {
      toast.error(`Connection failed: ${err}`);
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <WarningBanner>
            Choose a provider and add your API key to get started.
          </WarningBanner>
        )}

        {/* Provider picker — card per provider */}
        <div className="space-y-2">
          <h3 className="text-sm font-semibold flex items-center gap-1.5">
            <Plug className="size-4" /> AI Provider
          </h3>
          <div className="grid gap-2">
            {providers.map((p) => {
              const meta = providerPresets[p.id];
              const active = p.id === providerId;
              return (
                <button
                  key={p.id}
                  type="button"
                  onClick={() => setProviderId(p.id)}
                  className={`flex items-center gap-3 rounded-lg border-2 p-3 text-left transition-colors ${
                    active
                      ? "border-primary bg-primary/5"
                      : "border-border hover:border-muted-foreground/30"
                  }`}
                >
                  <div
                    className={`flex size-9 items-center justify-center rounded-md text-sm font-bold ${
                      active
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted text-muted-foreground"
                    }`}
                  >
                    {p.displayName[0]}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium">{p.displayName}</p>
                    {meta && (
                      <p className="text-xs text-muted-foreground truncate">
                        {meta.description}
                      </p>
                    )}
                  </div>
                  {active && <Check className="size-4 text-primary shrink-0" />}
                </button>
              );
            })}
          </div>
        </div>

        {/* API key + connection */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              API Key
              {preset && (
                <button
                  type="button"
                  className="text-primary underline underline-offset-2 hover:text-primary/80 text-xs font-normal"
                  onClick={() => BrowserOpenURL(preset.keyUrl)}
                >
                  {preset.keyUrlLabel}
                </button>
              )}
            </CardTitle>
            <CardDescription>
              All transcription and refinement use this connection.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
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

            {/* Advanced — base URL (only needed for self-hosted / proxies) */}
            <AdvancedToggle open={showAdvanced} onToggle={setShowAdvanced}>
              <div className="space-y-1">
                <Label className="text-xs flex items-center gap-1">
                  Base URL{" "}
                  <InfoTip text="Override the default API endpoint for self-hosted or proxy setups. Leave empty to use the provider's default." />
                </Label>
                <Input
                  value={baseUrl}
                  placeholder="Leave empty for default"
                  onChange={(e) => setBaseUrl(e.target.value)}
                  className="text-xs"
                />
              </div>
            </AdvancedToggle>

            <div className="flex items-center gap-2">
              <Button size="sm" onClick={save}>
                Save
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={testConnection}
                disabled={testing || !apiKey.trim()}
              >
                {testing ? (
                  <>
                    <Loader2 className="size-3.5 animate-spin" /> Testing...
                  </>
                ) : (
                  <>
                    <RefreshCw className="size-3.5" /> Test Connection
                  </>
                )}
              </Button>
              {modelCount !== null && (
                <span className="text-xs text-muted-foreground">
                  {modelCount} models found
                </span>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
