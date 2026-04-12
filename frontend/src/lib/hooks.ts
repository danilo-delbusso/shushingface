import { useState, useEffect, useCallback, useRef } from "react";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import type { config, history, desktop, platform, ai } from "../../wailsjs/go/models";
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
  return Boolean(settings.providerApiKey);
}

export function usePlatform() {
  const [platform, setPlatform] = useState<platform.Info | null>(null);
  useEffect(() => {
    AppBridge.GetPlatform().then(setPlatform);
  }, []);
  return platform;
}

export function usePasteStatus() {
  const [status, setStatus] = useState<desktop.PasteStatus | null>(null);
  useEffect(() => {
    AppBridge.GetPasteStatus().then(setStatus);
  }, []);
  return status;
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

export function useModels(settings: config.Settings | null) {
  const [models, setModels] = useState<ai.ModelInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async () => {
    if (!settings || !isConfigured(settings)) {
      setModels([]);
      return;
    }
    setLoading(true);
    try {
      const result = await AppBridge.ListModels();
      setModels(result ?? []);
    } catch {
      setModels([]);
    } finally {
      setLoading(false);
    }
  }, [settings?.providerId, settings?.providerApiKey]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const transcriptionModels = models.filter((m) => m.category === "transcription");
  const chatModels = models.filter((m) => m.category === "chat");

  return { models, transcriptionModels, chatModels, loading, refresh };
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
  const [results, setResults] = useState<desktop.ProcessResult[]>([]);
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
      } else {
        setResults((prev) => [res, ...prev]);
        onResult();
      }
    }
  }, [configured, onResult]);

  useEffect(() => {
    const cleanup = EventsOn("hotkey-toggle", toggle);
    return cleanup;
  }, [toggle]);

  return { isRecording, isProcessing, results, clearResults: () => setResults([]), toggle };
}
