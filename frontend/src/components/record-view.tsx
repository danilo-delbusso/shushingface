import { Mic, AlertTriangle, Settings, Copy, ChevronDown, ChevronUp, Coffee, Briefcase, Zap, PenTool } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { desktop, config } from "../../wailsjs/go/models";

const profileIcons: Record<string, React.FC<{ className?: string }>> = {
  coffee: Coffee,
  briefcase: Briefcase,
  zap: Zap,
  "pen-tool": PenTool,
};

interface RecordViewProps {
  configured: boolean;
  isRecording: boolean;
  isProcessing: boolean;
  results: desktop.ProcessResult[];
  activeProfile: config.RefinementProfile | null;
  onToggle: () => void;
  onGoToSettings: () => void;
}

export function RecordView({
  configured,
  isRecording,
  isProcessing,
  results,
  activeProfile,
  onToggle,
  onGoToSettings,
}: RecordViewProps) {
  if (!configured) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-4 text-center">
        <AlertTriangle className="size-12 text-amber-500" />
        <h2 className="text-xl font-semibold">Setup Required</h2>
        <p className="max-w-sm text-muted-foreground">
          Configure your API key before you can start transcribing.
        </p>
        <Button onClick={onGoToSettings}>
          <Settings className="size-4" />
          Go to Settings
        </Button>
      </div>
    );
  }

  const ProfileIcon = activeProfile ? profileIcons[activeProfile.icon] || PenTool : null;
  const hasResults = results.length > 0;

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      {/* Mic section */}
      <div className={`flex flex-col items-center gap-3 py-6 shrink-0 ${hasResults ? "" : "flex-1 justify-center"}`}>
        {activeProfile && ProfileIcon && (
          <div className="flex items-center gap-1.5 rounded-md border bg-card px-2.5 py-1 text-xs text-muted-foreground">
            <ProfileIcon className="size-3" />
            {activeProfile.name}
          </div>
        )}
        <button
          type="button"
          onClick={onToggle}
          disabled={isProcessing}
          className={`relative flex items-center justify-center rounded-full border-2 transition-all duration-200 hover:scale-105 ${
            hasResults ? "size-16" : "size-28"
          } ${
            isRecording
              ? "bg-red-500 border-red-500 text-white hover:bg-red-600"
              : isProcessing
                ? "bg-muted border-muted text-muted-foreground cursor-wait"
                : "bg-secondary border-primary text-foreground hover:bg-accent"
          }`}
        >
          <Mic className={hasResults ? "size-7" : "size-12"} />
          {isRecording && (
            <span className="absolute inset-0 animate-ping rounded-full border-4 border-red-500 opacity-30" />
          )}
        </button>
        <p className="text-xs text-muted-foreground">
          {isRecording
            ? "recording... click to stop"
            : isProcessing
              ? "processing..."
              : hasResults
                ? "click to record again"
                : "click to start recording"}
        </p>
      </div>

      {/* Results */}
      {hasResults && (
        <div className="flex-1 overflow-y-auto border-t">
          <div className="space-y-3 p-4 max-w-2xl mx-auto">
            {results.map((result, i) => (
              <ResultCard key={`${result.refined}-${i}`} result={result} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function ResultCard({ result }: { result: desktop.ProcessResult }) {
  const [showTranscript, setShowTranscript] = useState(false);

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-xs text-muted-foreground">
          Refined
        </CardTitle>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 text-xs"
          onClick={() => {
            navigator.clipboard.writeText(result.refined);
            toast.success("Copied");
          }}
        >
          <Copy className="size-3" />
          Copy
        </Button>
      </CardHeader>
      <CardContent className="space-y-2 pt-0">
        <p className="text-sm leading-relaxed whitespace-pre-wrap">
          {result.refined}
        </p>
        <button
          type="button"
          className="flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground transition-colors"
          onClick={() => setShowTranscript(!showTranscript)}
        >
          {showTranscript ? <ChevronUp className="size-3" /> : <ChevronDown className="size-3" />}
          transcript
        </button>
        {showTranscript && (
          <p className="text-xs text-muted-foreground italic">
            &ldquo;{result.transcript}&rdquo;
          </p>
        )}
      </CardContent>
    </Card>
  );
}
