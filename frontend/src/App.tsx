import { useState, useEffect } from "react";
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
import { ErrorBoundary } from "@/components/error-boundary";
import {
  useSettings,
  useHistory,
  useRecording,
  useTheme,
  usePlatform,
  usePasteStatus,
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

  const { isRecording, isProcessing, results, toggle } = useRecording(
    configured,
    refreshHistory,
  );

  useEffect(() => {
    if (settings && !settings.setupComplete) return;
    if (settings && !isConfigured(settings)) setView("connections");
    if (settings && !settings.enableHistory && view === "history")
      setView("home");
  }, [settings, view]);

  // Show wizard for first-time setup
  if (settings && !settings.setupComplete) {
    return (
      <ErrorBoundary>
        <TooltipProvider>
          <WelcomeWizard settings={settings} onComplete={saveSettings} />
          <Toaster position="bottom-right" richColors />
        </TooltipProvider>
      </ErrorBoundary>
    );
  }

  if (!settings) return null;

  return (
    <ErrorBoundary>
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar
          view={view}
          onNavigate={setView}
          configured={configured}
          historyEnabled={settings.enableHistory}
        />
        <SidebarInset className="flex flex-col h-screen overflow-hidden">
          {view === "home" && (
            <RecordView
              configured={configured}
              isRecording={isRecording}
              isProcessing={isProcessing}
              results={results}
              profiles={settings.refinementProfiles ?? []}
              activeProfile={
                settings.refinementProfiles?.find(
                  (p) => p.id === settings.activeProfileId,
                ) ?? null
              }
              onToggle={toggle}
              onGoToSettings={() => setView("connections")}
              onSwitchProfile={(id) =>
                saveSettings(
                  config.Settings.createFrom({
                    ...settings,
                    activeProfileId: id,
                  }),
                )
              }
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
            />
          )}
          {view === "ai" && (
            <AiView
              settings={settings}
              configured={configured}
              onSave={saveSettings}
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
                // Keep existing connections but reset everything else
                saveSettings(
                  config.Settings.createFrom({
                    ...defaults,
                    connections: settings.connections,
                    transcriptionConnectionId:
                      settings.transcriptionConnectionId,
                    refinementConnectionId: settings.refinementConnectionId,
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
    </ErrorBoundary>
  );
}

export default App;
