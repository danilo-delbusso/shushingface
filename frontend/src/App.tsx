import { useState, useEffect, useMemo } from "react";
import { EventsOn } from "../wailsjs/runtime/runtime";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import { AppSidebar, type View } from "@/components/app-sidebar";
import { RecordView } from "@/components/record-view";
import { HistoryView } from "@/components/history-view";
import { PlaygroundView } from "@/components/playground-view";
import {
  SettingsDialog,
  type SettingsSection,
} from "@/components/settings-dialog";
import { WelcomeWizard } from "@/components/welcome-wizard";
import { ErrorBoundary } from "@/components/error-boundary";
import {
  useSettings,
  useHistory,
  useRecording,
  useTheme,
  usePlatform,
  usePasteStatus,
  useCapabilities,
  isConfigured,
} from "@/lib/hooks";
import * as AppBridge from "../wailsjs/go/desktop/App";
import { config } from "../wailsjs/go/models";

function App() {
  const [view, setView] = useState<View>("home");
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [settingsSection, setSettingsSection] =
    useState<SettingsSection>("connections");
  const { settings, saveSettings } = useSettings();
  const {
    historyList,
    refresh: refreshHistory,
    clear: clearHistory,
  } = useHistory();
  const platform = usePlatform();
  const pasteStatus = usePasteStatus();
  const capabilities = useCapabilities();
  const [appVersion, setAppVersion] = useState("");
  const configured = isConfigured(settings);
  useTheme(settings?.theme);

  const [updateAvailable, setUpdateAvailable] = useState<{
    version: string;
    url: string;
  } | null>(null);

  useEffect(() => {
    AppBridge.GetVersion().then(setAppVersion);
    const cleanup = EventsOn(
      "update-available",
      (data: { version: string; url: string }) => {
        setUpdateAvailable(data);
      },
    );
    return cleanup;
  }, []);

  const { isRecording, isProcessing, toggle } = useRecording(
    configured,
    refreshHistory,
  );

  const activeProfile = useMemo(
    () =>
      settings?.refinementProfiles?.find(
        (p) => p.id === settings.activeProfileId,
      ) ?? null,
    [settings?.refinementProfiles, settings?.activeProfileId],
  );

  const openSettings = (section: SettingsSection = "connections") => {
    setSettingsSection(section);
    setSettingsOpen(true);
  };

  useEffect(() => {
    if (settings && !settings.setupComplete) return;
    if (settings && !isConfigured(settings)) {
      setSettingsSection("connections");
      setSettingsOpen(true);
    }
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
        <SidebarProvider defaultOpen={false}>
          <AppSidebar
            view={view}
            onNavigate={setView}
            onOpenSettings={() => openSettings()}
            historyEnabled={settings.enableHistory}
            hasWarnings={!configured}
            version={appVersion}
            updateAvailable={updateAvailable}
          />
          <SidebarInset className="flex flex-col h-screen overflow-hidden">
            {view === "home" && (
              <RecordView
                configured={configured}
                isRecording={isRecording}
                isProcessing={isProcessing}
                profiles={settings.refinementProfiles ?? []}
                activeProfile={activeProfile}
                language={settings.transcriptionLanguage ?? ""}
                historyEnabled={settings.enableHistory}
                historyItems={historyList}
                onToggle={toggle}
                onGoToSettings={() => openSettings("connections")}
                onSwitchProfile={(id) =>
                  saveSettings(
                    config.Settings.createFrom({
                      ...settings,
                      activeProfileId: id,
                    }),
                  )
                }
                onSwitchLanguage={(code) =>
                  saveSettings(
                    config.Settings.createFrom({
                      ...settings,
                      transcriptionLanguage: code || undefined,
                    }),
                  )
                }
              />
            )}
            {view === "history" && (
              <HistoryView items={historyList} onClear={clearHistory} />
            )}
            {view === "playground" && (
              <PlaygroundView
                settings={settings}
                onSwitchProfile={(id) =>
                  saveSettings(
                    config.Settings.createFrom({
                      ...settings,
                      activeProfileId: id,
                    }),
                  )
                }
                onEditStyles={() => openSettings("styles")}
              />
            )}
          </SidebarInset>
          <SettingsDialog
            open={settingsOpen}
            onOpenChange={setSettingsOpen}
            section={settingsSection}
            onSectionChange={setSettingsSection}
            settings={settings}
            configured={configured}
            hasWarnings={!configured}
            platform={platform}
            pasteAvailable={pasteStatus?.available ?? true}
            pasteInstallCmd={pasteStatus?.installCmd ?? ""}
            capabilities={capabilities}
            appVersion={appVersion}
            onSave={saveSettings}
          />
        </SidebarProvider>
        <Toaster position="bottom-right" richColors />
      </TooltipProvider>
    </ErrorBoundary>
  );
}

export default App;
