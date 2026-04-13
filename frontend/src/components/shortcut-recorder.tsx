import { useEffect, useState } from "react";
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

function isModifierKey(key: string): boolean {
  return (
    key === "Control" ||
    key === "Alt" ||
    key === "Shift" ||
    key === "Meta" ||
    key === "OS"
  );
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

interface LiveMods {
  ctrl: boolean;
  alt: boolean;
  shift: boolean;
  super: boolean;
}

const NO_MODS: LiveMods = {
  ctrl: false,
  alt: false,
  shift: false,
  super: false,
};

export function ShortcutRecorder({
  value,
  onChange,
  disabled,
}: ShortcutRecorderProps) {
  const [recording, setRecording] = useState(false);
  const [draft, setDraft] = useState<string>(value);
  const [liveMods, setLiveMods] = useState<LiveMods>(NO_MODS);

  useEffect(() => {
    setDraft(value);
  }, [value]);

  useEffect(() => {
    if (!recording) return;
    setLiveMods(NO_MODS);

    const updateMods = (e: KeyboardEvent) =>
      setLiveMods({
        ctrl: e.ctrlKey,
        alt: e.altKey,
        shift: e.shiftKey,
        super: e.metaKey,
      });

    const onDown = (e: KeyboardEvent) => {
      // Always swallow the event while recording so the OS doesn't act on it.
      e.preventDefault();
      e.stopPropagation();

      if (e.key === "Escape") {
        setRecording(false);
        return;
      }

      updateMods(e);
      const key = canonicalKey(e);
      if (!key) return; // modifier-only keypress — keep waiting

      const mods: string[] = [];
      if (e.ctrlKey) mods.push(MOD_LABELS.ctrl);
      if (e.altKey) mods.push(MOD_LABELS.alt);
      if (e.shiftKey) mods.push(MOD_LABELS.shift);
      if (e.metaKey) mods.push(MOD_LABELS.super);
      if (mods.length === 0) return; // require at least one modifier

      const spec = [...mods, key].join("+");
      setDraft(spec);
      onChange(spec);
      setRecording(false);
    };

    const onUp = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();
      updateMods(e);
    };

    window.addEventListener("keydown", onDown, true);
    window.addEventListener("keyup", onUp, true);
    return () => {
      window.removeEventListener("keydown", onDown, true);
      window.removeEventListener("keyup", onUp, true);
    };
  }, [recording, onChange]);

  const liveDisplay = (() => {
    const parts: string[] = [];
    if (liveMods.ctrl) parts.push("Ctrl");
    if (liveMods.alt) parts.push("Alt");
    if (liveMods.shift) parts.push("Shift");
    if (liveMods.super) parts.push("Super");
    return parts.length ? parts.join("+") + "+…" : "Press a combination…";
  })();

  const display = recording ? liveDisplay : draft || "Not set";

  return (
    <div
      className={`flex w-full items-center gap-2 rounded-md border bg-background px-3 py-2 text-sm transition-colors ${
        recording ? "border-primary ring-1 ring-primary" : "border-input"
      } ${disabled ? "opacity-50" : ""}`}
    >
      <Keyboard className="size-3.5 shrink-0 text-muted-foreground" />
      <span
        className={`flex-1 text-left ${
          draft && !recording ? "" : "text-muted-foreground"
        }`}
      >
        {display}
      </span>
      <button
        type="button"
        disabled={disabled}
        onClick={() => setRecording((r) => !r)}
        className="rounded-md border border-input px-2 py-0.5 text-xs hover:bg-muted disabled:opacity-50"
      >
        {recording ? "Cancel" : draft ? "Change" : "Record"}
      </button>
    </div>
  );
}
