import { useState, useEffect } from "react";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import { AppSidebar, type View } from "@/components/app-sidebar";
import { RecordView } from "@/components/record-view";
import { HistoryView } from "@/components/history-view";
import { AiView } from "@/components/ai-view";
import { AppearanceView } from "@/components/appearance-view";
import { SettingsView } from "@/components/settings-view";
import { WelcomeWizard } from "@/components/welcome-wizard";
import { useSettings, useHistory, useRecording, useTheme, usePlatform, isConfigured } from "@/lib/hooks";
import { config } from "../wailsjs/go/models";

function App() {
  const [view, setView] = useState<View>("home");
  const { settings, saveSettings } = useSettings();
  const { historyList, refresh: refreshHistory, clear: clearHistory } = useHistory();
  const platform = usePlatform();
  const configured = isConfigured(settings);
  useTheme(settings?.theme);

  const { isRecording, isProcessing, results, toggle } =
    useRecording(configured, refreshHistory);

  useEffect(() => {
    if (settings && !settings.setupComplete) return; // wizard handles this
    if (settings && !isConfigured(settings)) setView("ai");
    if (settings && !settings.enableHistory && view === "history") setView("home");
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
              onGoToSettings={() => setView("ai")}
            />
          )}
          {view === "history" && (
            <HistoryView items={historyList} onClear={clearHistory} />
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
              onSave={saveSettings}
              onRunSetup={() =>
                saveSettings(
                  config.Settings.createFrom({
                    ...settings,
                    setupComplete: false,
                  }),
                )
              }
            />
          )}
        </SidebarInset>
      </SidebarProvider>
      <Toaster position="bottom-right" richColors />
    </TooltipProvider>
  );
}

export default App;
