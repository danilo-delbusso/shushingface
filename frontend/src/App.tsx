import { useState, useEffect, useMemo } from "react";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import { AppSidebar, type View } from "@/components/app-sidebar";
import { RecordView } from "@/components/record-view";
import { HistoryView } from "@/components/history-view";
import { ConnectionsView } from "@/components/connections-view";
import { AiView } from "@/components/ai-view";
import { AppearanceView } from "@/components/appearance-view";
import { SettingsView } from "@/components/settings-view";
import { WelcomeWizard } from "@/components/welcome-wizard";
import {
  useSettings,
  useHistory,
  useRecording,
  useTheme,
  usePlatform,
  usePasteStatus,
  useModels,
  isConfigured,
} from "@/lib/hooks";
import * as AppBridge from "../wailsjs/go/desktop/App";
import { config } from "../wailsjs/go/models";

function App() {
  const [view, setView] = useState<View>("home");
  const { settings, saveSettings } = useSettings();
  const {
    historyList,
    refresh: refreshHistory,
    clear: clearHistory,
  } = useHistory();
  const platform = usePlatform();
  const pasteStatus = usePasteStatus();
  const configured = isConfigured(settings);
  useTheme(settings?.theme);

  const { transcriptionModels, chatModels, refresh: refreshModels } =
    useModels(settings);

  const { isRecording, isProcessing, results, toggle } = useRecording(
    configured,
    refreshHistory,
  );

  // Detect broken model references for warning icons
  const hasWarnings = useMemo(() => {
    if (!settings || chatModels.length === 0) return false;
    const allIds = new Set([
      ...transcriptionModels.map((m) => m.id),
      ...chatModels.map((m) => m.id),
    ]);
    if (settings.transcriptionModel && !allIds.has(settings.transcriptionModel))
      return true;
    if (settings.refinementModel && !allIds.has(settings.refinementModel))
      return true;
    for (const p of settings.refinementProfiles ?? []) {
      if (p.model && !allIds.has(p.model)) return true;
    }
    return false;
  }, [settings, transcriptionModels, chatModels]);

  useEffect(() => {
    if (settings && !settings.setupComplete) return;
    if (settings && !isConfigured(settings)) setView("connections");
    if (settings && !settings.enableHistory && view === "history")
      setView("home");
  }, [settings, view]);

  // Show wizard for first-time setup
  if (settings && !settings.setupComplete) {
    return (
      <>
        <WelcomeWizard settings={settings} onComplete={saveSettings} />
        <Toaster position="bottom-right" richColors />
      </>
    );
  }

  if (!settings) return null;

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar
          view={view}
          onNavigate={setView}
          configured={configured}
          historyEnabled={settings.enableHistory}
          hasWarnings={hasWarnings}
        />
        <SidebarInset className="flex flex-col h-screen overflow-hidden">
          {view === "home" && (
            <RecordView
              configured={configured}
              isRecording={isRecording}
              isProcessing={isProcessing}
              results={results}
              activeProfile={
                settings.refinementProfiles?.find(
                  (p) => p.id === settings.activeProfileId,
                ) ?? null
              }
              onToggle={toggle}
              onGoToSettings={() => setView("connections")}
            />
          )}
          {view === "history" && (
            <HistoryView items={historyList} onClear={clearHistory} />
          )}
          {view === "connections" && (
            <ConnectionsView
              settings={settings}
              configured={configured}
              onSave={saveSettings}
              onModelsRefreshed={refreshModels}
            />
          )}
          {view === "ai" && (
            <AiView
              settings={settings}
              configured={configured}
              onSave={saveSettings}
              transcriptionModels={transcriptionModels}
              chatModels={chatModels}
            />
          )}
          {view === "appearance" && (
            <AppearanceView settings={settings} onSave={saveSettings} />
          )}
          {view === "general" && (
            <SettingsView
              settings={settings}
              platform={platform}
              pasteAvailable={pasteStatus?.available ?? true}
              pasteInstallCmd={pasteStatus?.installCmd ?? ""}
              onSave={saveSettings}
              onRunSetup={async () => {
                const defaults = await AppBridge.GetDefaultSettings();
                const apiKey = settings.providerApiKey ?? "";
                saveSettings(
                  config.Settings.createFrom({
                    ...defaults,
                    providerApiKey: apiKey,
                    setupComplete: false,
                  }),
                );
              }}
            />
          )}
        </SidebarInset>
      </SidebarProvider>
      <Toaster position="bottom-right" richColors />
    </TooltipProvider>
  );
}

export default App;
