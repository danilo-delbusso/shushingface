import type { platform } from "../../wailsjs/go/models";

interface ShortcutGuideProps {
  platform: platform.Info | null;
  /** Compact mode for the wizard — no Card wrapper, just inline content. */
  compact?: boolean;
}

const command = "shushingface --toggle";

interface Tip {
  env: string;
  path: string;
  notes?: string;
}

function getTips(platform: platform.Info): Tip[] {
  const de = platform.desktop?.toUpperCase() || "";

  if (platform.os === "darwin") {
    return [
      {
        env: "macOS",
        path: "System Settings → Keyboard → Keyboard Shortcuts → App Shortcuts",
        notes: "Global shortcut support is coming soon.",
      },
    ];
  }

  if (platform.os === "windows") {
    return [
      {
        env: "Windows",
        path: "Create a shortcut and assign a hotkey in its Properties",
        notes: "Global shortcut support is coming soon.",
      },
    ];
  }

  // Linux — detect desktop environment
  if (de.includes("COSMIC")) {
    return [
      {
        env: "COSMIC",
        path: "Settings → Keyboard → Custom Shortcuts → Add Shortcut",
      },
    ];
  }

  if (de.includes("POP") || de.includes("GNOME")) {
    return [
      {
        env: de.includes("POP") ? "Pop!_OS / GNOME" : "GNOME",
        path: "Settings → Keyboard → Custom Shortcuts → Add Custom Shortcut",
      },
    ];
  }

  if (de.includes("KDE") || de.includes("PLASMA")) {
    return [
      {
        env: "KDE Plasma",
        path: "System Settings → Shortcuts → Custom Shortcuts → Edit → New → Global Shortcut → Command",
      },
    ];
  }

  if (de.includes("XFCE")) {
    return [
      {
        env: "XFCE",
        path: "Settings → Keyboard → Application Shortcuts → Add",
      },
    ];
  }

  if (de.includes("CINNAMON")) {
    return [
      {
        env: "Cinnamon",
        path: "System Settings → Keyboard → Shortcuts → Custom Shortcuts → Add",
      },
    ];
  }

  // Fallback
  return [
    {
      env: "Linux",
      path: "Your desktop's keyboard shortcut settings",
      notes: "Look for Custom Shortcuts or Application Shortcuts.",
    },
  ];
}

export function ShortcutGuide({ platform, compact }: ShortcutGuideProps) {
  if (!platform) return null;

  const tips = getTips(platform);

  if (compact) {
    return (
      <div className="space-y-2 text-left">
        {tips.map((tip) => (
          <div key={tip.env} className="space-y-1">
            <p className="text-xs text-muted-foreground">
              Bind{" "}
              <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
                {command}
              </code>{" "}
              to a key in <strong>{tip.path}</strong>
            </p>
            {tip.notes && (
              <p className="text-[11px] text-muted-foreground/70">
                {tip.notes}
              </p>
            )}
          </div>
        ))}
      </div>
    );
  }

  // Standard mode (for settings page CardDescription)
  const tip = tips[0];
  return (
    <div className="space-y-1">
      <p className="text-sm text-muted-foreground">
        Bind{" "}
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
          {command}
        </code>{" "}
        to a key in <strong>{tip.path}</strong>
      </p>
      {tip.notes && (
        <p className="text-xs text-muted-foreground/70">{tip.notes}</p>
      )}
    </div>
  );
}
