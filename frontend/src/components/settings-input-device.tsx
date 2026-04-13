import { useEffect, useState } from "react";
import { Mic } from "lucide-react";
import { toast } from "sonner";
import { InfoTip } from "@/components/info-tip";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { audio, config } from "../../wailsjs/go/models";

interface SettingsInputDeviceProps {
  settings: config.Settings;
  onSave: (settings: config.Settings) => void;
}

export function SettingsInputDevice({ settings, onSave }: SettingsInputDeviceProps) {
  const [devices, setDevices] = useState<audio.DeviceInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    AppBridge.ListInputDevices()
      .then((d) => setDevices(d ?? []))
      .catch((err) => toast.error(`Failed to list microphones: ${err}`))
      .finally(() => setLoading(false));
  }, []);

  const change = async (id: string) => {
    try {
      const next = { ...settings, inputDeviceId: id } as config.Settings;
      onSave(next);
    } catch (err) {
      toast.error(`Failed to switch microphone: ${err}`);
    }
  };

  const current = settings.inputDeviceId ?? "";

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-sm">
          <Mic className="size-4" /> Microphone{" "}
          <InfoTip text="Pick which capture device records your voice. The default follows the system default." />
        </CardTitle>
        <CardDescription>
          {loading ? "Detecting devices..." : `${devices.length} device${devices.length === 1 ? "" : "s"} available`}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        <Label htmlFor="input-device">Input device</Label>
        <select
          id="input-device"
          disabled={loading || devices.length === 0}
          value={current}
          onChange={(e) => change(e.target.value)}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-50"
        >
          <option value="">System default</option>
          {devices.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
              {d.isDefault ? " (default)" : ""}
            </option>
          ))}
        </select>
      </CardContent>
    </Card>
  );
}
