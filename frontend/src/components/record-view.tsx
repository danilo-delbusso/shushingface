import { Mic, AlertTriangle, Settings, Copy, ChevronDown, ChevronUp, Check } from "lucide-react";
import { useState } from "react";
import { Popover } from "radix-ui";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getProfileIcon } from "@/lib/icons";
import type { desktop, config } from "../../wailsjs/go/models";

interface RecordViewProps {
  configured: boolean;
  isRecording: boolean;
  isProcessing: boolean;
  results: desktop.ProcessResult[];
  profiles: config.RefinementProfile[];
  activeProfile: config.RefinementProfile | null;
  onToggle: () => void;
  onGoToSettings: () => void;
  onSwitchProfile: (id: string) => void;
}

export function RecordView({
  configured,
  isRecording,
  isProcessing,
  results,
  profiles,
  activeProfile,
  onToggle,
  onGoToSettings,
  onSwitchProfile,
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

  const ProfileIcon = activeProfile ? getProfileIcon(activeProfile.icon) : null;
  const hasResults = results.length > 0;

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      {/* Mic section */}
      <div className={`flex flex-col items-center gap-3 py-6 shrink-0 ${hasResults ? "" : "flex-1 justify-center"}`}>
        {activeProfile && ProfileIcon && (
          <ProfileSwitcher
            profiles={profiles}
            activeProfile={activeProfile}
            onSwitch={onSwitchProfile}
          />
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
            {results.map((result) => (
              <ResultCard key={`${result.refined}-${result.transcript}`} result={result} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function ProfileSwitcher({
  profiles,
  activeProfile,
  onSwitch,
}: {
  profiles: config.RefinementProfile[];
  activeProfile: config.RefinementProfile;
  onSwitch: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const ActiveIcon = getProfileIcon(activeProfile.icon);

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger asChild>
        <button
          type="button"
          className="flex items-center gap-1.5 rounded-md border bg-card px-2.5 py-1 text-xs text-muted-foreground hover:bg-accent hover:text-foreground transition-colors cursor-pointer"
        >
          <ActiveIcon className="size-3 shrink-0" />
          {activeProfile.name}
          <ChevronDown className="size-3 opacity-50" />
        </button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          sideOffset={4}
          className="z-50 w-48 rounded-md border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
        >
          {profiles.map((p) => {
            const Icon = getProfileIcon(p.icon);
            const isActive = p.id === activeProfile.id;
            return (
              <button
                key={p.id}
                type="button"
                onClick={() => {
                  onSwitch(p.id);
                  setOpen(false);
                }}
                className={`flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-xs transition-colors ${
                  isActive
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent hover:text-accent-foreground"
                }`}
              >
                <Icon className="size-3.5 shrink-0" />
                <span className="flex-1 text-left">{p.name}</span>
                {isActive && <Check className="size-3 shrink-0 opacity-50" />}
              </button>
            );
          })}
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
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
