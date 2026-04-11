import { useState } from "react";
import {
  Key,
  Bot,
  Keyboard,
  SlidersHorizontal,
  Palette,
  Eye,
  EyeOff,
  AlertTriangle,
  Sun,
  Moon,
  Monitor,
} from "lucide-react";
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
import { Switch } from "@/components/ui/switch";
import { config, type desktop } from "../../wailsjs/go/models";

interface SettingsViewProps {
  settings: config.Settings;
  configured: boolean;
  platform: desktop.PlatformInfo | null;
  onSave: (settings: config.Settings) => void;
}

export function SettingsView({
  settings,
  configured,
  platform,
  onSave,
}: SettingsViewProps) {
  const [draft, setDraft] = useState(settings);
  const [showKey, setShowKey] = useState(false);

  const update = (patch: Partial<config.Settings>) => {
    setDraft(config.Settings.createFrom({ ...draft, ...patch }));
  };

  const updateProvider = (field: string, value: string) => {
    const id = draft.transcriptionProviderId;
    setDraft(
      config.Settings.createFrom({
        ...draft,
        providers: {
          ...draft.providers,
          [id]: { ...draft.providers[id], [field]: value },
        },
      }),
    );
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(draft);
  };

  const provider = draft.providers?.[draft.transcriptionProviderId];

  return (
    <div className="flex-1 overflow-y-auto">
      <form onSubmit={handleSubmit} className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <div className="flex items-center gap-3 rounded-lg border border-amber-600/30 bg-amber-600/10 p-3 text-sm text-amber-500">
            <AlertTriangle className="size-4 shrink-0" />
            Add your API key to get started.
          </div>
        )}

        {/* Theme */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Palette className="size-4" /> Appearance
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              {([
                { value: "light", icon: Sun, label: "Light" },
                { value: "dark", icon: Moon, label: "Dark" },
                { value: "system", icon: Monitor, label: "System" },
              ] as const).map(({ value, icon: Icon, label }) => (
                <Button
                  key={value}
                  type="button"
                  variant={draft.theme === value ? "default" : "outline"}
                  size="sm"
                  className="flex-1"
                  onClick={() => update({ theme: value })}
                >
                  <Icon className="size-3.5" />
                  {label}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* API Provider */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Key className="size-4" /> API Provider
            </CardTitle>
            <CardDescription>
              shushingface uses Groq for fast speech-to-text.{" "}
              <button
                type="button"
                className="text-primary underline underline-offset-2 hover:text-primary/80"
                onClick={() =>
                  window.open("https://console.groq.com/keys", "_blank")
                }
              >
                Get a free key
              </button>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label htmlFor="api-key">API Key</Label>
              <div className="flex">
                <Input
                  id="api-key"
                  type={showKey ? "text" : "password"}
                  value={provider?.apiKey ?? ""}
                  placeholder="gsk_..."
                  onChange={(e) => updateProvider("apiKey", e.target.value)}
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
          </CardContent>
        </Card>

        {/* Transcription Model */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Bot className="size-4" /> Transcription
            </CardTitle>
            <CardDescription>
              The speech-to-text model that converts your audio to a raw
              transcript.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label htmlFor="trans-model">Model</Label>
              <Input
                id="trans-model"
                value={draft.transcriptionModel}
                onChange={(e) =>
                  update({ transcriptionModel: e.target.value })
                }
              />
            </div>
          </CardContent>
        </Card>

        {/* Shortcuts */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Keyboard className="size-4" /> Shortcuts
            </CardTitle>
            <ShortcutGuide platform={platform} />
          </CardHeader>
        </Card>

        {/* Preferences */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <SlidersHorizontal className="size-4" /> Preferences
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="auto-copy">Auto-copy to clipboard</Label>
                <p className="text-xs text-muted-foreground">
                  Copy refined text automatically after processing
                </p>
              </div>
              <Switch
                id="auto-copy"
                checked={draft.autoCopy}
                onCheckedChange={(v) => update({ autoCopy: v })}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="history-toggle">Save history</Label>
                <p className="text-xs text-muted-foreground">
                  Store transcriptions locally in SQLite
                </p>
              </div>
              <Switch
                id="history-toggle"
                checked={draft.enableHistory}
                onCheckedChange={(v) => update({ enableHistory: v })}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="indicator-toggle">Panel indicator</Label>
                <p className="text-xs text-muted-foreground">
                  Show a mic icon in the system panel bar
                </p>
              </div>
              <Switch
                id="indicator-toggle"
                checked={draft.enableIndicator}
                onCheckedChange={(v) => update({ enableIndicator: v })}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="notify-toggle">Desktop notifications</Label>
                <p className="text-xs text-muted-foreground">
                  Show notifications when recording starts and stops
                </p>
              </div>
              <Switch
                id="notify-toggle"
                checked={draft.enableNotifications}
                onCheckedChange={(v) => update({ enableNotifications: v })}
              />
            </div>
          </CardContent>
        </Card>

        <Button type="submit" className="w-full">
          Save Changes
        </Button>
      </form>
    </div>
  );
}

function ShortcutGuide({
  platform,
}: { platform: desktop.PlatformInfo | null }) {
  if (!platform) return null;

  const command = (
    <code className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
      shushingface --toggle
    </code>
  );

  if (platform.os === "darwin" || platform.os === "windows") {
    return <CardDescription>Global shortcut support coming soon.</CardDescription>;
  }

  const de = platform.desktop?.toUpperCase() || "";
  const path = de.includes("KDE") || de.includes("PLASMA")
    ? "System Settings → Shortcuts → Custom Shortcuts"
    : "Settings → Keyboard → Custom Shortcuts";

  return (
    <CardDescription>
      Bind {command} to a key in <strong>{path}</strong>
    </CardDescription>
  );
}
