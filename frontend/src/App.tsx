import { useState, useEffect } from "react";
import "./App.css";
import {
  Mic,
  History,
  Settings as SettingsIcon,
  Home,
  Copy,
  ChevronDown,
  ChevronUp,
  Trash2,
} from "lucide-react";
import * as AppBridge from "../wailsjs/go/desktop/App";
import { config, history, desktop } from "../wailsjs/go/models";

type View = "home" | "history" | "settings";

function App() {
  const [view, setView] = useState<View>("home");
  const [isRecording, setIsRecording] = useState(false);
  const [status, setStatus] = useState("Ready");
  const [result, setResult] = useState<desktop.ProcessResult | null>(null);
  const [showTranscript, setShowTranscript] = useState(false);

  const [settings, setSettings] = useState<config.Settings | null>(null);
  const [historyList, setHistoryList] = useState<history.Record[]>([]);

  useEffect(() => {
    // Load initial data
    AppBridge.GetSettings().then(setSettings);
    refreshHistory();
  }, []);

  const refreshHistory = () => {
    AppBridge.GetHistory(50, 0)
      .then(setHistoryList)
      .catch(() => setHistoryList([]));
  };

  const toggleRecording = async () => {
    if (!isRecording) {
      try {
        await AppBridge.StartRecording();
        setIsRecording(true);
        setStatus("Recording...");
        setResult(null);
      } catch (err) {
        setStatus(`Error: ${err}`);
      }
    } else {
      setIsRecording(false);
      setStatus("Processing with AI...");
      const res = await AppBridge.StopAndProcess();
      setResult(res);
      setStatus(res.error ? "Error" : "Ready");
      if (!res.error) refreshHistory();
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    // Visual feedback could go here
  };

  const saveSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    if (settings) {
      await AppBridge.SaveSettings(settings);
      setStatus("Settings Saved");
      setTimeout(() => setStatus("Ready"), 2000);
    }
  };

  const clearHistory = async () => {
    if (confirm("Are you sure you want to wipe all history?")) {
      await AppBridge.ClearHistory();
      refreshHistory();
    }
  };

  return (
    <div className="container">
      <nav className="navbar">
        <div
          className={`nav-item ${view === "home" ? "active" : ""}`}
          onClick={() => setView("home")}
        >
          <Home size={18} /> Home
        </div>
        <div
          className={`nav-item ${view === "history" ? "active" : ""}`}
          onClick={() => setView("history")}
        >
          <History size={18} /> History
        </div>
        <div
          className={`nav-item ${view === "settings" ? "active" : ""}`}
          onClick={() => setView("settings")}
        >
          <SettingsIcon size={18} /> Settings
        </div>
      </nav>

      <main className="main-content">
        {view === "home" && (
          <div className="record-container">
            {!result ? (
              <>
                <button
                  className={`record-button ${isRecording ? "recording" : ""}`}
                  onClick={toggleRecording}
                >
                  <Mic size={48} />
                </button>
                <div className="status-text">{status}</div>
              </>
            ) : (
              <div className="result-card">
                <div className="result-header">
                  <span className="status-text" style={{ fontSize: "0.9rem" }}>
                    Refined Message
                  </span>
                  <button
                    className="copy-button"
                    onClick={() => copyToClipboard(result.refined)}
                  >
                    <Copy size={16} /> Copy
                  </button>
                </div>
                <div className="refined-text">{result.refined}</div>

                <div className="transcript-section">
                  <div
                    className="transcript-header"
                    onClick={() => setShowTranscript(!showTranscript)}
                  >
                    {showTranscript ? (
                      <ChevronUp size={16} />
                    ) : (
                      <ChevronDown size={16} />
                    )}
                    Raw Transcript
                  </div>
                  {showTranscript && (
                    <div className="raw-transcript">{result.transcript}</div>
                  )}
                </div>

                <button
                  className="button-primary"
                  style={{ marginTop: "1rem" }}
                  onClick={() => setResult(null)}
                >
                  New Recording
                </button>
              </div>
            )}
          </div>
        )}

        {view === "history" && (
          <div className="list-container">
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: "1rem",
              }}
            >
              <h2>History</h2>
              <button className="button-danger" onClick={clearHistory}>
                <Trash2 size={16} /> Clear All
              </button>
            </div>
            {historyList.map((item) => (
              <div key={item.id} className="history-item">
                <div className="history-meta">
                  <span>{new Date(item.timestamp).toLocaleString()}</span>
                  <span className="history-app">{item.activeApp}</span>
                </div>
                <div style={{ fontWeight: 500 }}>{item.refinedMessage}</div>
                <div
                  style={{
                    fontSize: "0.85rem",
                    color: "var(--text-dim)",
                    fontStyle: "italic",
                  }}
                >
                  "{item.rawTranscript}"
                </div>
              </div>
            ))}
            {historyList.length === 0 && (
              <div className="status-text">No history found.</div>
            )}
          </div>
        )}

        {view === "settings" && settings && (
          <form
            className="list-container settings-form"
            onSubmit={saveSettings}
          >
            <h2>Settings</h2>

            <div className="settings-group">
              <label>Global Hotkey</label>
              <input
                type="text"
                value={settings.globalHotkey}
                onChange={(e) =>
                  setSettings(
                    config.Settings.createFrom({
                      ...settings,
                      globalHotkey: e.target.value,
                    }),
                  )
                }
              />
            </div>

            <div className="settings-group">
              <label>Refinement Model (Groq)</label>
              <input
                type="text"
                value={settings.refinementModel}
                onChange={(e) =>
                  setSettings(
                    config.Settings.createFrom({
                      ...settings,
                      refinementModel: e.target.value,
                    }),
                  )
                }
              />
            </div>

            <div
              className="settings-group"
              style={{
                flexDirection: "row",
                alignItems: "center",
                gap: "1rem",
              }}
            >
              <input
                type="checkbox"
                checked={settings.autoCopy}
                onChange={(e) =>
                  setSettings(
                    config.Settings.createFrom({
                      ...settings,
                      autoCopy: e.target.checked,
                    }),
                  )
                }
                style={{ width: "auto" }}
              />
              <label>Auto-copy to clipboard when finished</label>
            </div>

            <div
              className="settings-group"
              style={{
                flexDirection: "row",
                alignItems: "center",
                gap: "1rem",
              }}
            >
              <input
                type="checkbox"
                checked={settings.enableHistory}
                onChange={(e) =>
                  setSettings(
                    config.Settings.createFrom({
                      ...settings,
                      enableHistory: e.target.checked,
                    }),
                  )
                }
                style={{ width: "auto" }}
              />
              <label>Enable local transcription history</label>
            </div>

            <button type="submit" className="button-primary">
              Save Changes
            </button>
            <div
              className="status-text"
              style={{ textAlign: "center", fontSize: "0.9rem" }}
            >
              {status}
            </div>
          </form>
        )}
      </main>
    </div>
  );
}

export default App;
