import { Mic, Eye } from "lucide-react";
import { InfoTip } from "@/components/info-tip";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import type { config } from "../../wailsjs/go/models";

interface SettingsRecordingProps {
  settings: config.Settings;
  onSave: (settings: config.Settings) => void;
}

export function SettingsRecording({ settings, onSave }: SettingsRecordingProps) {
  const mode = settings.recordingMode || "toggle";
  const opacity = settings.overlayOpacity ?? 0.4;

  const setMode = (m: "toggle" | "push_to_talk") =>
    onSave({ ...settings, recordingMode: m } as config.Settings);

  const setOverlay = (v: boolean) =>
    onSave({ ...settings, overlayEnabled: v } as config.Settings);

  const setOpacity = (v: number) =>
    onSave({ ...settings, overlayOpacity: v } as config.Settings);

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-sm">
          <Mic className="size-4" /> Recording{" "}
          <InfoTip text="Choose how the shortcut starts and stops recording, and whether a translucent indicator floats above the focused window." />
        </CardTitle>
        <CardDescription>
          Toggle starts on first press and stops on second. Push-to-talk records only while the shortcut is held.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label>Mode</Label>
          <div className="grid grid-cols-2 gap-2">
            <button
              type="button"
              onClick={() => setMode("toggle")}
              className={`rounded-md border px-3 py-2 text-sm ${
                mode === "toggle"
                  ? "border-primary bg-primary/10"
                  : "border-input hover:bg-muted"
              }`}
            >
              Toggle
            </button>
            <button
              type="button"
              onClick={() => setMode("push_to_talk")}
              className={`rounded-md border px-3 py-2 text-sm ${
                mode === "push_to_talk"
                  ? "border-primary bg-primary/10"
                  : "border-input hover:bg-muted"
              }`}
            >
              Push-to-talk
            </button>
          </div>
        </div>

        <Separator />

        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="overlay-toggle" className="flex items-center gap-2">
              <Eye className="size-3.5" /> Floating overlay
            </Label>
            <p className="text-xs text-muted-foreground">
              Show a translucent "Recording" pill at the bottom of the focused window
            </p>
          </div>
          <Switch
            id="overlay-toggle"
            checked={settings.overlayEnabled ?? true}
            onCheckedChange={setOverlay}
          />
        </div>

        {settings.overlayEnabled !== false && (
          <div className="space-y-2">
            <Label htmlFor="overlay-opacity">Opacity ({Math.round(opacity * 100)}%)</Label>
            <input
              id="overlay-opacity"
              type="range"
              min={0.1}
              max={1}
              step={0.05}
              value={opacity}
              onChange={(e) => setOpacity(Number(e.target.value))}
              className="w-full"
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}
