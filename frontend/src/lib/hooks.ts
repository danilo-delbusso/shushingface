import { useState, useEffect, useCallback, useRef } from "react";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import type { config, history, desktop } from "../../wailsjs/go/models";
import { toast } from "sonner";

export function useTheme(theme: string | undefined) {
  useEffect(() => {
    const root = document.documentElement;
    const resolved =
      theme === "system"
        ? window.matchMedia("(prefers-color-scheme: dark)").matches
          ? "dark"
          : "light"
        : theme || "dark";
    root.classList.toggle("dark", resolved === "dark");
  }, [theme]);
}

export function isConfigured(settings: config.Settings | null): boolean {
  if (!settings) return false;
  const provider = settings.providers?.[settings.transcriptionProviderId];
  return Boolean(provider?.apiKey);
}

export function usePlatform() {
  const [platform, setPlatform] = useState<desktop.PlatformInfo | null>(null);
  useEffect(() => {
    AppBridge.GetPlatform().then(setPlatform);
  }, []);
  return platform;
}

export function useSettings() {
  const [settings, setSettings] = useState<config.Settings | null>(null);

  useEffect(() => {
    AppBridge.GetSettings().then(setSettings);
  }, []);

  const saveSettings = useCallback(
    async (updated: config.Settings) => {
      try {
        await AppBridge.SaveSettings(updated);
        setSettings(updated);
        toast.success("Settings saved");
      } catch (err) {
        toast.error(`Failed to save: ${err}`);
      }
    },
    [],
  );

  return { settings, setSettings, saveSettings };
}

export function useHistory() {
  const [historyList, setHistoryList] = useState<history.Record[]>([]);

  const refresh = useCallback(() => {
    AppBridge.GetHistory(50, 0)
      .then((items) => setHistoryList(items ?? []))
      .catch(() => setHistoryList([]));
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const clear = useCallback(async () => {
    await AppBridge.ClearHistory();
    setHistoryList([]);
    toast.success("History cleared");
  }, []);

  return { historyList, refresh, clear };
}

export function useRecording(
  configured: boolean,
  onResult: () => void,
) {
  const [isRecording, setIsRecording] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [result, setResult] = useState<desktop.ProcessResult | null>(null);
  const isRecordingRef = useRef(isRecording);
  isRecordingRef.current = isRecording;

  const toggle = useCallback(async () => {
    if (!configured) {
      toast.error("Set up your API key in Settings first");
      return;
    }
    if (!isRecordingRef.current) {
      try {
        await AppBridge.StartRecording();
        setIsRecording(true);
        setResult(null);
      } catch (err) {
        toast.error(`Recording failed: ${err}`);
      }
    } else {
      setIsRecording(false);
      setIsProcessing(true);
      const res = await AppBridge.StopAndProcess();
      setIsProcessing(false);
      if (res.error) {
        toast.error(res.error);
        setResult(null);
      } else {
        setResult(res);
        onResult();
      }
    }
  }, [configured, onResult]);

  useEffect(() => {
    const cleanup = EventsOn("hotkey-toggle", toggle);
    return cleanup;
  }, [toggle]);

  return { isRecording, isProcessing, result, setResult, toggle };
}
