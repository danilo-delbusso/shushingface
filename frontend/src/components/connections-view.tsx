import { useState, useEffect } from "react";
import {
  Key,
  Eye,
  EyeOff,
  AlertTriangle,
  Plug,
  Loader2,
  RefreshCw,
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
import { InfoTip } from "@/components/info-tip";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { config, ai } from "../../wailsjs/go/models";

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
  const [testing, setTesting] = useState(false);
  const [modelCount, setModelCount] = useState<number | null>(null);

  useEffect(() => {
    AppBridge.ListProviders().then(setProviders);
  }, []);

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
    // Save first to update backend state, then try listing models
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
          <div className="flex items-center gap-3 rounded-lg border border-amber-600/30 bg-amber-600/10 p-3 text-sm text-amber-500">
            <AlertTriangle className="size-4 shrink-0" />
            Add your API key to get started.
          </div>
        )}

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Plug className="size-4" /> AI Provider{" "}
              <InfoTip text="Choose which AI service to use for transcription and refinement." />
            </CardTitle>
            <CardDescription>
              All transcription and refinement models come from this provider.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1">
              <Label>Provider</Label>
              <Select
                value={providerId}
                onValueChange={setProviderId}
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

            <div className="space-y-1">
              <Label>
                API Key{" "}
                {providerId === "groq" && (
                  <button
                    type="button"
                    className="text-primary underline underline-offset-2 hover:text-primary/80 text-xs ml-1"
                    onClick={() =>
                      window.open("https://console.groq.com/keys", "_blank")
                    }
                  >
                    Get a free key
                  </button>
                )}
              </Label>
              <div className="flex">
                <Input
                  type={showKey ? "text" : "password"}
                  value={apiKey}
                  placeholder="gsk_..."
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

            <div className="space-y-1">
              <Label className="flex items-center gap-1">
                Base URL{" "}
                <InfoTip text="Optional. Override the default API endpoint for self-hosted or proxy setups." />
              </Label>
              <Input
                value={baseUrl}
                placeholder="Leave empty for default"
                onChange={(e) => setBaseUrl(e.target.value)}
              />
            </div>

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
                  <><Loader2 className="size-3.5 animate-spin" /> Testing...</>
                ) : (
                  <><RefreshCw className="size-3.5" /> Test Connection</>
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
