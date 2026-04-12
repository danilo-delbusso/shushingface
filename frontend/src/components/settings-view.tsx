import { Keyboard, SlidersHorizontal, AlertTriangle } from "lucide-react";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
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
import { Switch } from "@/components/ui/switch";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { config, platform } from "../../wailsjs/go/models";

interface SettingsViewProps {
  settings: config.Settings;
  platform: platform.Info | null;
  pasteAvailable: boolean;
  pasteInstallCmd: string;
  onSave: (settings: config.Settings) => void;
  onRunSetup: () => void;
}

export function SettingsView({
  settings,
  platform,
  pasteAvailable,
  pasteInstallCmd,
  onSave,
  onRunSetup,
}: SettingsViewProps) {
  const toggle = (patch: Partial<config.Settings>) => {
    onSave({ ...settings, ...patch } as config.Settings);
  };

  const deleteAllData = async () => {
    try {
      await AppBridge.DeleteAllData();
      toast.success("All data deleted. Restarting setup...");
      // Reload to trigger wizard since setupComplete is now false
      window.location.reload();
    } catch (err) {
      toast.error(`Failed to delete data: ${err}`);
    }
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Keyboard className="size-4" /> Shortcuts{" "}
              <InfoTip text="Configure a system keyboard shortcut to toggle recording from any app without opening the window." />
            </CardTitle>
            <ShortcutGuide platform={platform} />
          </CardHeader>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <SlidersHorizontal className="size-4" /> Preferences{" "}
              <InfoTip text="Control clipboard behavior, history storage, and system integrations." />
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="auto-paste">Auto-paste</Label>
                <p className="text-xs text-muted-foreground">
                  Type refined text into the focused app after processing
                </p>
              </div>
              <Switch
                id="auto-paste"
                checked={settings.autoPaste}
                onCheckedChange={(v) => toggle({ autoPaste: v })}
              />
            </div>
            {settings.autoPaste && !pasteAvailable && pasteInstallCmd && (
              <div className="flex items-start gap-2 rounded-md border border-amber-600/30 bg-amber-600/10 px-3 py-2 text-xs text-amber-500">
                <AlertTriangle className="size-3.5 mt-0.5 shrink-0" />
                <div>
                  <p>Auto-paste requires an external tool. Run:</p>
                  <code className="mt-1 block rounded bg-muted px-2 py-1 font-mono text-foreground">
                    {pasteInstallCmd}
                  </code>
                </div>
              </div>
            )}
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="auto-copy">Auto-copy to clipboard</Label>
                <p className="text-xs text-muted-foreground">
                  Copy refined text to clipboard after processing
                </p>
              </div>
              <Switch
                id="auto-copy"
                checked={settings.autoCopy}
                onCheckedChange={(v) => toggle({ autoCopy: v })}
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
                checked={settings.enableHistory}
                onCheckedChange={(v) => toggle({ enableHistory: v })}
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
                checked={settings.enableIndicator}
                onCheckedChange={(v) => toggle({ enableIndicator: v })}
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
                checked={settings.enableNotifications}
                onCheckedChange={(v) => toggle({ enableNotifications: v })}
              />
            </div>
          </CardContent>
        </Card>

        <Separator />

        <Card className="border-destructive/30">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm text-destructive">
              <AlertTriangle className="size-4 shrink-0" /> Danger Zone
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <p className="text-sm font-medium">Reset & run setup</p>
                <p className="text-xs text-muted-foreground">
                  Resets settings to defaults and re-runs the wizard. API key is
                  kept.
                </p>
              </div>
              <ConfirmDialog
                trigger={
                  <Button variant="destructive" size="sm">
                    Reset
                  </Button>
                }
                title="Reset and run setup?"
                description="This resets all settings to defaults and re-runs the setup wizard. Your API key will be kept."
                confirmLabel="Reset"
                onConfirm={onRunSetup}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <p className="text-sm font-medium">Delete all data</p>
                <p className="text-xs text-muted-foreground">
                  Erases history, settings, API keys, and profiles. Cannot be
                  undone.
                </p>
              </div>
              <ConfirmDialog
                trigger={
                  <Button variant="destructive" size="sm">
                    Delete All
                  </Button>
                }
                title="Delete all data?"
                description="This permanently erases your history, settings, API keys, and all custom profiles. The app will restart as if freshly installed. This cannot be undone."
                confirmLabel="Delete Everything"
                variant="destructive"
                onConfirm={deleteAllData}
              />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function ShortcutGuide({ platform }: { platform: platform.Info | null }) {
  if (!platform) return null;

  const command = (
    <code className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
      shushingface --toggle
    </code>
  );

  if (platform.os === "darwin" || platform.os === "windows") {
    return (
      <CardDescription>Global shortcut support coming soon.</CardDescription>
    );
  }

  const de = platform.desktop?.toUpperCase() || "";
  const path =
    de.includes("KDE") || de.includes("PLASMA")
      ? "System Settings → Shortcuts → Custom Shortcuts"
      : "Settings → Keyboard → Custom Shortcuts";

  return (
    <CardDescription>
      Bind {command} to a key in <strong>{path}</strong>
    </CardDescription>
  );
}
