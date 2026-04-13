import { useEffect, useState } from "react";
import { AlertTriangle, Keyboard } from "lucide-react";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
import { ShortcutGuide } from "@/components/shortcut-guide";
import { ShortcutRecorder } from "@/components/shortcut-recorder";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { platform } from "../../wailsjs/go/models";

interface SettingsShortcutProps {
  platform: platform.Info | null;
}

export function SettingsShortcut({ platform }: SettingsShortcutProps) {
  const [caps, setCaps] = useState<platform.Capability | null>(null);
  const [current, setCurrent] = useState<string>("");
  const [draft, setDraft] = useState<string>("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    AppBridge.HotkeyCapabilities()
      .then(setCaps)
      .catch((err) => toast.error(`Hotkey check failed: ${err}`));
    AppBridge.GetShortcut()
      .then((s) => {
        setCurrent(s);
        setDraft(s);
      })
      .catch(() => {});
  }, []);

  if (!caps) return null;

  if (!caps.supported) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-sm">
            <Keyboard className="size-4" /> Shortcut{" "}
            <InfoTip text="In-app hotkey binding is unavailable on this platform — bind it from your desktop's keyboard settings." />
          </CardTitle>
          <CardDescription>
            <ShortcutGuide platform={platform} />
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3 rounded-md border border-amber-600/30 bg-amber-600/10 p-3 text-sm text-amber-500">
            <AlertTriangle className="size-4 shrink-0" />
            {caps.reason || "This platform does not expose a global hotkey API."}
          </div>
        </CardContent>
      </Card>
    );
  }

  const save = async () => {
    if (!draft) return;
    setSaving(true);
    try {
      await AppBridge.SetShortcut(draft);
      setCurrent(draft);
      toast.success(`Bound ${draft}`);
    } catch (err) {
      const msg = String(err);
      if (msg.includes("already registered")) {
        toast.error(`'${draft}' is already taken by another app`);
      } else {
        toast.error(`Failed to bind: ${msg}`);
      }
    } finally {
      setSaving(false);
    }
  };

  const clear = async () => {
    try {
      await AppBridge.ClearShortcut();
      setCurrent("");
      setDraft("");
      toast.success("Shortcut cleared");
    } catch (err) {
      toast.error(`Failed to clear: ${err}`);
    }
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-sm">
          <Keyboard className="size-4" /> Shortcut{" "}
          <InfoTip text="Global hotkey to toggle recording from any app." />
        </CardTitle>
        <CardDescription>
          Press the combination you want, then save. Modifier required.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <ShortcutRecorder value={draft} onChange={setDraft} disabled={saving} />
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={save}
            disabled={saving || !draft || draft === current}
          >
            Save
          </Button>
          {current && (
            <ConfirmDialog
              title="Clear shortcut?"
              description={`This will remove the global ${current} binding.`}
              onConfirm={clear}
              trigger={
                <Button variant="destructive" size="sm">
                  Clear
                </Button>
              }
            />
          )}
        </div>
      </CardContent>
    </Card>
  );
}
