import { useState } from "react";
import {
  Key,
  Bot,
  Keyboard,
  SlidersHorizontal,
  Eye,
  EyeOff,
  AlertTriangle,
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
import { ScrollArea } from "@/components/ui/scroll-area";
import { config } from "../../wailsjs/go/models";

interface SettingsViewProps {
  settings: config.Settings;
  configured: boolean;
  onSave: (settings: config.Settings) => void;
}

export function SettingsView({
  settings,
  configured,
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
    <ScrollArea className="flex-1">
      <form onSubmit={handleSubmit} className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <div className="flex items-center gap-3 rounded-lg border border-yellow-500/30 bg-yellow-500/10 p-3 text-sm text-yellow-500">
            <AlertTriangle className="size-4 shrink-0" />
            Add your API key to get started.
          </div>
        )}

        {/* API Provider */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Key className="size-4" /> API Provider
            </CardTitle>
            <CardDescription>
              Sussurro uses Groq for fast speech-to-text.{" "}
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
            <CardDescription>
              Global hotkey to toggle recording from any app. Requires restart.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label htmlFor="hotkey">Global Hotkey</Label>
              <Input
                id="hotkey"
                value={draft.globalHotkey}
                placeholder="e.g. Ctrl+Shift+R"
                onChange={(e) => update({ globalHotkey: e.target.value })}
              />
            </div>
          </CardContent>
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
          </CardContent>
        </Card>

        <Button type="submit" className="w-full">
          Save Changes
        </Button>
      </form>
    </ScrollArea>
  );
}
