import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import type {
  config,
  history,
  desktop,
  platform,
  ai,
} from "../../wailsjs/go/models";
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
  return (
    (settings.connections?.length ?? 0) > 0 &&
    Boolean(settings.transcriptionConnectionId)
  );
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

// useCapabilities reports which platform-dependent features work here so
// the settings UI can grey out toggles that would be no-ops. Returns null
// while loading.
export function useCapabilities() {
  const [caps, setCaps] = useState<desktop.Capabilities | null>(null);
  useEffect(() => {
    AppBridge.GetCapabilities().then(setCaps);
  }, []);
  return caps;
}

export function useSettings() {
  const [settings, setSettings] = useState<config.Settings | null>(null);

  useEffect(() => {
    AppBridge.GetSettings().then(setSettings);
  }, []);

  const saveSettings = useCallback(async (updated: config.Settings) => {
    try {
      await AppBridge.SaveSettings(updated);
      setSettings(updated);
      toast.success("Settings saved");
    } catch (err) {
      toast.error(`Failed to save: ${err}`);
    }
  }, []);

  return { settings, setSettings, saveSettings };
}

/** Fetch models for a specific connection ID, with cross-instance cache. */
const modelCache = new Map<string, ai.ModelInfo[]>();

export function useModelsForConnection(connectionId: string | undefined) {
  const [models, setModels] = useState<ai.ModelInfo[]>(
    () => (connectionId ? modelCache.get(connectionId) : undefined) ?? [],
  );
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async () => {
    if (!connectionId) {
      setModels([]);
      return;
    }
    setLoading(true);
    try {
      const result = await AppBridge.ListModelsForConnection(connectionId);
      const list = result ?? [];
      modelCache.set(connectionId, list);
      setModels(list);
    } catch {
      setModels([]);
    } finally {
      setLoading(false);
    }
  }, [connectionId]);

  useEffect(() => {
    const cached = connectionId ? modelCache.get(connectionId) : undefined;
    if (cached) {
      setModels(cached);
    }
    refresh();
  }, [refresh, connectionId]);

  const transcriptionModels = useMemo(
    () => models.filter((m) => m.category === "transcription"),
    [models],
  );
  const chatModels = useMemo(
    () => models.filter((m) => m.category === "chat"),
    [models],
  );

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

export function useRecording(configured: boolean, onResult: () => void) {
  const [isRecording, setIsRecording] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
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
        onResult();
      }
    }
  }, [configured, onResult]);

  const startRecording = useCallback(async () => {
    if (!configured) {
      toast.error("Set up your API key in Settings first");
      return;
    }
    if (isRecordingRef.current) return;
    try {
      await AppBridge.StartRecording();
      setIsRecording(true);
    } catch (err) {
      toast.error(`Recording failed: ${err}`);
    }
  }, [configured]);

  const stopAndProcess = useCallback(async () => {
    if (!isRecordingRef.current) return;
    setIsRecording(false);
    setIsProcessing(true);
    const res = await AppBridge.StopAndProcess();
    setIsProcessing(false);
    if (res.error) {
      toast.error(res.error);
    } else {
      onResult();
    }
  }, [onResult]);

  useEffect(() => {
    const cleanups = [
      EventsOn("hotkey-toggle", toggle),
      EventsOn("hotkey-press", startRecording),
      EventsOn("hotkey-release", stopAndProcess),
    ];
    return () => {
      for (const c of cleanups) c();
    };
  }, [toggle, startRecording, stopAndProcess]);

  return {
    isRecording,
    isProcessing,
    toggle,
  };
}
