import { useState } from "react";
import { Play, Loader2, FlaskConical, Pencil, ChevronDown } from "lucide-react";
import { Popover } from "radix-ui";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { textareaClass, cn } from "@/lib/utils";
import { getProfileIcon } from "@/lib/icons";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { config } from "../../wailsjs/go/models";

const PLACEHOLDER =
  "hey um so I was thinking we should probably move the meeting to thursday because like john cant make it on wednesday and I think it would be better if we all met together you know";

interface PlaygroundViewProps {
  settings: config.Settings;
  onSwitchProfile: (id: string) => void;
  onEditStyles: () => void;
}

export function PlaygroundView({
  settings,
  onSwitchProfile,
  onEditStyles,
}: PlaygroundViewProps) {
  const profiles = settings.refinementProfiles ?? [];
  const activeProfile =
    profiles.find((p) => p.id === settings.activeProfileId) ?? null;

  const [sampleText, setSampleText] = useState("");
  const [result, setResult] = useState("");
  const [running, setRunning] = useState(false);

  const handleRun = async () => {
    const text = sampleText.trim() || PLACEHOLDER;
    if (!activeProfile?.prompt) {
      toast.error("Active style has no prompt — edit it first");
      return;
    }
    setRunning(true);
    setResult("");
    try {
      const res = await AppBridge.TestPrompt(text, activeProfile.prompt);
      if (res.error) toast.error(res.error);
      else setResult(res.refined);
    } catch (err) {
      toast.error(`Test failed: ${err}`);
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <header className="flex items-center justify-between gap-4 border-b px-6 py-4">
        <div className="flex items-center gap-2">
          <FlaskConical className="size-5 text-primary" />
          <div>
            <h1 className="text-lg font-semibold">Playground</h1>
            <p className="text-xs text-muted-foreground">
              Test how a style transforms sample text without recording.
            </p>
          </div>
        </div>
        <StylePicker
          profiles={profiles}
          activeProfile={activeProfile}
          onSwitch={onSwitchProfile}
          onEdit={onEditStyles}
        />
      </header>

      <div className="flex-1 overflow-y-auto">
        <div className="space-y-4 p-6 max-w-3xl">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">Input</CardTitle>
              <CardDescription>
                Paste a transcript or leave empty for a built-in sample.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <textarea
                value={sampleText}
                onChange={(e) => setSampleText(e.target.value)}
                rows={5}
                className={textareaClass}
                placeholder={PLACEHOLDER}
              />
              <Button
                onClick={handleRun}
                disabled={running || !activeProfile}
                className="w-full"
              >
                {running ? (
                  <>
                    <Loader2 className="size-4 animate-spin" /> Running...
                  </>
                ) : (
                  <>
                    <Play className="size-4" /> Run with{" "}
                    {activeProfile?.name ?? "—"}
                  </>
                )}
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">Result</CardTitle>
              <CardDescription>
                Tweak the prompt and re-run to compare.{" "}
                <button
                  type="button"
                  onClick={onEditStyles}
                  className="text-primary underline-offset-2 hover:underline"
                >
                  Edit styles →
                </button>
              </CardDescription>
            </CardHeader>
            <CardContent>
              {result ? (
                <p className="rounded-md bg-muted/50 p-3 text-sm leading-relaxed whitespace-pre-wrap">
                  {result}
                </p>
              ) : (
                <p className="text-sm text-muted-foreground italic">
                  Run the playground to see the refined output here.
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function StylePicker({
  profiles,
  activeProfile,
  onSwitch,
  onEdit,
}: {
  profiles: config.RefinementProfile[];
  activeProfile: config.RefinementProfile | null;
  onSwitch: (id: string) => void;
  onEdit: () => void;
}) {
  const [open, setOpen] = useState(false);
  const ActiveIcon = activeProfile ? getProfileIcon(activeProfile.icon) : null;

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger asChild>
        <button
          type="button"
          className="flex items-center gap-2 rounded-md border bg-card px-3 py-1.5 text-xs hover:bg-accent transition-colors"
        >
          {ActiveIcon && <ActiveIcon className="size-3.5 shrink-0" />}
          <span className="font-medium">
            {activeProfile?.name ?? "No style"}
          </span>
          <ChevronDown className="size-3 opacity-50" />
        </button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          align="end"
          sideOffset={4}
          className="z-50 w-56 rounded-md border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
        >
          {profiles.map((p) => {
            const Icon = getProfileIcon(p.icon);
            const isActive = p.id === activeProfile?.id;
            return (
              <button
                key={p.id}
                type="button"
                onClick={() => {
                  onSwitch(p.id);
                  setOpen(false);
                }}
                className={cn(
                  "flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-xs",
                  isActive
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent hover:text-accent-foreground",
                )}
              >
                <Icon className="size-3.5 shrink-0" />
                <span className="flex-1 truncate">{p.name}</span>
              </button>
            );
          })}
          <div className="my-1 h-px bg-border" />
          <button
            type="button"
            onClick={() => {
              setOpen(false);
              onEdit();
            }}
            className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-xs text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          >
            <Pencil className="size-3.5 shrink-0" />
            <span className="flex-1">Edit styles…</span>
          </button>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
