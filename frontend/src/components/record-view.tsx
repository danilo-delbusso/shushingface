import {
  Mic,
  AlertTriangle,
  Settings,
  ChevronDown,
  Check,
  Globe,
} from "lucide-react";
import { useState } from "react";
import { Popover } from "radix-ui";
import { Button } from "@/components/ui/button";
import { HomeStats } from "@/components/home-stats";
import { getProfileIcon } from "@/lib/icons";
import { whisperLanguages, getLanguage } from "@/lib/languages";
import type { config, history } from "../../wailsjs/go/models";

interface RecordViewProps {
  configured: boolean;
  isRecording: boolean;
  isProcessing: boolean;
  profiles: config.RefinementProfile[];
  activeProfile: config.RefinementProfile | null;
  language: string;
  historyEnabled: boolean;
  historyItems: history.Record[];
  onToggle: () => void;
  onGoToSettings: () => void;
  onSwitchProfile: (id: string) => void;
  onSwitchLanguage: (code: string) => void;
}

export function RecordView({
  configured,
  isRecording,
  isProcessing,
  profiles,
  activeProfile,
  language,
  historyEnabled,
  historyItems,
  onToggle,
  onGoToSettings,
  onSwitchProfile,
  onSwitchLanguage,
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

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <header className="flex items-start justify-between gap-4 border-b px-6 py-4">
        <div>
          <h1 className="text-lg font-semibold">Welcome back</h1>
          <p className="text-xs text-muted-foreground">
            Hold your shortcut and start speaking.
          </p>
        </div>
        {historyEnabled && <HomeStats items={historyItems} />}
      </header>
      <div className="flex flex-1 flex-col items-center justify-center gap-6 p-6 overflow-y-auto">
      <div className="flex items-center gap-2">
        {activeProfile && ProfileIcon && (
          <ProfileSwitcher
            profiles={profiles}
            activeProfile={activeProfile}
            onSwitch={onSwitchProfile}
          />
        )}
        <LanguageSwitcher language={language} onSwitch={onSwitchLanguage} />
      </div>
      <button
        type="button"
        onClick={onToggle}
        disabled={isProcessing}
        className={`relative flex size-28 items-center justify-center rounded-full border-2 transition-all duration-200 hover:scale-105 ${
          isRecording
            ? "bg-red-500 border-red-500 text-white hover:bg-red-600"
            : isProcessing
              ? "bg-muted border-muted text-muted-foreground cursor-wait"
              : "bg-secondary border-primary text-foreground hover:bg-accent"
        }`}
      >
        <Mic className="size-12" />
        {isRecording && (
          <span className="absolute inset-0 animate-ping rounded-full border-4 border-red-500 opacity-30" />
        )}
      </button>
      <p className="text-xs text-muted-foreground">
        {isRecording
          ? "recording... click to stop"
          : isProcessing
            ? "processing..."
            : "click to start recording"}
      </p>
      </div>
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

function LanguageSwitcher({
  language,
  onSwitch,
}: {
  language: string;
  onSwitch: (code: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const current = getLanguage(language);

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger asChild>
        <button
          type="button"
          className="flex items-center gap-1.5 rounded-md border bg-card px-2.5 py-1 text-xs text-muted-foreground hover:bg-accent hover:text-foreground transition-colors cursor-pointer"
        >
          {current ? (
            <span>{current.flag}</span>
          ) : (
            <Globe className="size-3 shrink-0" />
          )}
          {current?.name ?? "Auto"}
          <ChevronDown className="size-3 opacity-50" />
        </button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          sideOffset={4}
          className="z-50 max-h-60 w-48 overflow-y-auto rounded-md border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
        >
          <button
            type="button"
            onClick={() => {
              onSwitch("");
              setOpen(false);
            }}
            className={`flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-xs transition-colors ${
              !language
                ? "bg-accent text-accent-foreground"
                : "hover:bg-accent hover:text-accent-foreground"
            }`}
          >
            <Globe className="size-3.5 shrink-0" />
            <span className="flex-1 text-left">Auto-detect</span>
            {!language && <Check className="size-3 shrink-0 opacity-50" />}
          </button>
          {whisperLanguages.map((lang) => {
            const isActive = lang.code === language;
            return (
              <button
                key={lang.code}
                type="button"
                onClick={() => {
                  onSwitch(isActive ? "" : lang.code);
                  setOpen(false);
                }}
                className={`flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-xs transition-colors ${
                  isActive
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent hover:text-accent-foreground"
                }`}
              >
                <span className="shrink-0">{lang.flag}</span>
                <span className="flex-1 text-left">{lang.name}</span>
                {isActive && <Check className="size-3 shrink-0 opacity-50" />}
              </button>
            );
          })}
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
