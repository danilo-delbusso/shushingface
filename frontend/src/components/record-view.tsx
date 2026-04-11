import { Mic, AlertTriangle, Settings, Copy, ChevronDown, ChevronUp } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import type { desktop } from "../../wailsjs/go/models";

interface RecordViewProps {
  configured: boolean;
  isRecording: boolean;
  isProcessing: boolean;
  result: desktop.ProcessResult | null;
  onToggle: () => void;
  onNewRecording: () => void;
  onGoToSettings: () => void;
}

export function RecordView({
  configured,
  isRecording,
  isProcessing,
  result,
  onToggle,
  onNewRecording,
  onGoToSettings,
}: RecordViewProps) {
  const [showTranscript, setShowTranscript] = useState(false);

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

  if (result) {
    return (
      <div className="flex flex-1 items-start justify-center overflow-y-auto p-6">
        <Card className="w-full max-w-2xl">
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Refined Message
            </CardTitle>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                navigator.clipboard.writeText(result.refined);
                toast.success("Copied to clipboard");
              }}
            >
              <Copy className="size-3.5" />
              Copy
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-lg leading-relaxed whitespace-pre-wrap">
              {result.refined}
            </p>
            <Separator />
            <button
              type="button"
              className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
              onClick={() => setShowTranscript(!showTranscript)}
            >
              {showTranscript ? (
                <ChevronUp className="size-4" />
              ) : (
                <ChevronDown className="size-4" />
              )}
              Raw Transcript
            </button>
            {showTranscript && (
              <p className="text-sm text-muted-foreground italic leading-relaxed">
                {result.transcript}
              </p>
            )}
            <Button className="w-full" onClick={onNewRecording}>
              New Recording
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-6">
      <button
        type="button"
        onClick={onToggle}
        disabled={isProcessing}
        className={`relative flex size-28 items-center justify-center rounded-full border-0 text-white transition-all duration-200 hover:scale-105 ${
          isRecording
            ? "bg-red-500 hover:bg-red-600"
            : isProcessing
              ? "bg-muted cursor-wait"
              : "bg-primary hover:bg-primary/90"
        }`}
      >
        <Mic className="size-12" />
        {isRecording && (
          <span className="absolute inset-0 animate-ping rounded-full border-4 border-red-500 opacity-30" />
        )}
      </button>
      <p className="text-sm font-medium text-muted-foreground">
        {isRecording
          ? "Recording... click to stop"
          : isProcessing
            ? "Processing with AI..."
            : "Click to start recording"}
      </p>
    </div>
  );
}
