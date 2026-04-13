import { useEffect, useRef, useState } from "react";
import { Keyboard } from "lucide-react";

interface ShortcutRecorderProps {
  value: string;
  onChange: (spec: string) => void;
  disabled?: boolean;
}

const MOD_LABELS: Record<string, string> = {
  ctrl: "Ctrl",
  alt: "Alt",
  shift: "Shift",
  super: "Super",
};

function isModifierKey(key: string): string | null {
  switch (key) {
    case "Control":
      return "ctrl";
    case "Alt":
      return "alt";
    case "Shift":
      return "shift";
    case "Meta":
    case "OS":
      return "super";
    default:
      return null;
  }
}

function canonicalKey(e: KeyboardEvent): string | null {
  if (isModifierKey(e.key)) return null;
  if (e.key === " ") return "Space";
  if (e.key === "Escape") return "Escape";
  if (e.key === "Enter") return "Enter";
  if (e.key === "Tab") return "Tab";
  if (/^F\d{1,2}$/.test(e.key)) return e.key;
  if (/^Arrow/.test(e.key)) return e.key.replace("Arrow", "");
  if (e.key.length === 1) return e.key.toUpperCase();
  return e.key;
}

export function ShortcutRecorder({
  value,
  onChange,
  disabled,
}: ShortcutRecorderProps) {
  const [recording, setRecording] = useState(false);
  const [draft, setDraft] = useState<string>(value);
  const inputRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    setDraft(value);
  }, [value]);

  useEffect(() => {
    if (!recording) return;
    const handler = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();
      const key = canonicalKey(e);
      if (!key) return;
      const mods: string[] = [];
      if (e.ctrlKey) mods.push(MOD_LABELS.ctrl);
      if (e.altKey) mods.push(MOD_LABELS.alt);
      if (e.shiftKey) mods.push(MOD_LABELS.shift);
      if (e.metaKey) mods.push(MOD_LABELS.super);
      if (mods.length === 0) {
        // require at least one modifier
        return;
      }
      const spec = [...mods, key].join("+");
      setDraft(spec);
      setRecording(false);
      onChange(spec);
    };
    window.addEventListener("keydown", handler, true);
    return () => window.removeEventListener("keydown", handler, true);
  }, [recording, onChange]);

  const display = recording ? "Press a combination..." : draft || "Not set";

  return (
    <button
      ref={inputRef}
      type="button"
      disabled={disabled}
      onClick={() => setRecording(true)}
      onBlur={() => setRecording(false)}
      className={`flex w-full items-center gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-50 ${
        recording ? "ring-1 ring-primary border-primary" : ""
      }`}
    >
      <Keyboard className="size-3.5 shrink-0 text-muted-foreground" />
      <span
        className={`flex-1 text-left ${
          draft && !recording ? "" : "text-muted-foreground"
        }`}
      >
        {display}
      </span>
      <span className="text-xs text-muted-foreground">
        {recording ? "Esc to cancel" : "Click to record"}
      </span>
    </button>
  );
}
