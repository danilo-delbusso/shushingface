import { useState, useEffect } from "react";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import { AppSidebar, type View } from "@/components/app-sidebar";
import { RecordView } from "@/components/record-view";
import { HistoryView } from "@/components/history-view";
import { SettingsView } from "@/components/settings-view";
import { PromptView } from "@/components/prompt-view";
import { useSettings, useHistory, useRecording, isConfigured } from "@/lib/hooks";

function App() {
  const [view, setView] = useState<View>("home");
  const { settings, saveSettings } = useSettings();
  const { historyList, refresh: refreshHistory, clear: clearHistory } = useHistory();
  const configured = isConfigured(settings);

  const { isRecording, isProcessing, result, setResult, toggle } =
    useRecording(configured, refreshHistory);

  useEffect(() => {
    if (settings && !isConfigured(settings)) {
      setView("settings");
    }
    if (settings && !settings.enableHistory && view === "history") {
      setView("home");
    }
  }, [settings, view]);

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar
          view={view}
          onNavigate={setView}
          configured={configured}
          historyEnabled={settings?.enableHistory ?? false}
          hotkey={settings?.globalHotkey}
        />
        <SidebarInset className="flex flex-col h-screen overflow-hidden">
          {view === "home" && (
            <RecordView
              configured={configured}
              isRecording={isRecording}
              isProcessing={isProcessing}
              result={result}
              onToggle={toggle}
              onNewRecording={() => setResult(null)}
              onGoToSettings={() => setView("settings")}
            />
          )}
          {view === "history" && (
            <HistoryView items={historyList} onClear={clearHistory} />
          )}
          {view === "settings" && settings && (
            <SettingsView
              settings={settings}
              configured={configured}
              onSave={saveSettings}
            />
          )}
          {view === "prompt" && settings && (
            <PromptView settings={settings} onSave={saveSettings} />
          )}
        </SidebarInset>
      </SidebarProvider>
      <Toaster position="bottom-right" richColors />
    </TooltipProvider>
  );
}

export default App;
