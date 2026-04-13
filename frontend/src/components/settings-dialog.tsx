import {
  Plug,
  Bot,
  ScrollText,
  Wand2,
  Palette,
  SlidersHorizontal,
  Info,
  AlertTriangle,
  X,
} from "lucide-react";
import { Dialog as DialogPrimitive } from "radix-ui";
import { cn } from "@/lib/utils";
import { ConnectionsView } from "@/components/connections-view";
import { ModelsView } from "@/components/models-view";
import { RulesView } from "@/components/rules-view";
import { StylesView } from "@/components/styles-view";
import { AppearanceView } from "@/components/appearance-view";
import { SettingsView } from "@/components/settings-view";
import { AboutView } from "@/components/about-view";
import type { config, desktop, platform } from "../../wailsjs/go/models";

export type SettingsSection =
  | "connections"
  | "models"
  | "rules"
  | "styles"
  | "appearance"
  | "general"
  | "about";

interface SettingsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  section: SettingsSection;
  onSectionChange: (s: SettingsSection) => void;
  settings: config.Settings;
  configured: boolean;
  hasWarnings: boolean;
  platform: platform.Info | null;
  pasteAvailable: boolean;
  pasteInstallCmd: string;
  capabilities: desktop.Capabilities | null;
  appVersion: string;
  onSave: (settings: config.Settings) => void;
}

const NAV: {
  id: SettingsSection;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  group: "general" | "ai" | "info";
}[] = [
  { id: "connections", label: "Connections", icon: Plug, group: "ai" },
  { id: "models", label: "Models", icon: Bot, group: "ai" },
  { id: "rules", label: "Rules", icon: ScrollText, group: "ai" },
  { id: "styles", label: "Styles", icon: Wand2, group: "ai" },
  {
    id: "appearance",
    label: "Appearance",
    icon: Palette,
    group: "general",
  },
  {
    id: "general",
    label: "General",
    icon: SlidersHorizontal,
    group: "general",
  },
  { id: "about", label: "About", icon: Info, group: "info" },
];

export function SettingsDialog({
  open,
  onOpenChange,
  section,
  onSectionChange,
  settings,
  configured,
  hasWarnings,
  platform,
  pasteAvailable,
  pasteInstallCmd,
  capabilities,
  appVersion,
  onSave,
}: SettingsDialogProps) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/50 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:animate-in data-[state=open]:fade-in-0" />
        <DialogPrimitive.Content
          className={cn(
            "fixed top-[50%] left-[50%] z-50 flex h-[85vh] w-[90vw] max-w-4xl translate-x-[-50%] translate-y-[-50%] overflow-hidden rounded-lg border bg-background shadow-lg outline-none",
            "data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95",
          )}
        >
          <DialogPrimitive.Title className="sr-only">
            Settings
          </DialogPrimitive.Title>
          <DialogPrimitive.Description className="sr-only">
            Configure shushing face.
          </DialogPrimitive.Description>

          <SettingsNav
            section={section}
            onSelect={onSectionChange}
            configured={configured}
            hasWarnings={hasWarnings}
            appVersion={appVersion}
          />

          <SettingsBody section={section}>
            {section === "connections" && (
              <ConnectionsView
                settings={settings}
                configured={configured}
                onSave={onSave}
              />
            )}
            {section === "models" && (
              <ModelsView
                settings={settings}
                configured={configured}
                onSave={onSave}
              />
            )}
            {section === "rules" && (
              <RulesView settings={settings} onSave={onSave} />
            )}
            {section === "styles" && (
              <StylesView settings={settings} onSave={onSave} />
            )}
            {section === "appearance" && (
              <AppearanceView settings={settings} onSave={onSave} />
            )}
            {section === "general" && (
              <SettingsView
                settings={settings}
                platform={platform}
                pasteAvailable={pasteAvailable}
                pasteInstallCmd={pasteInstallCmd}
                capabilities={capabilities}
                onSave={onSave}
              />
            )}
            {section === "about" && (
              <AboutView version={appVersion} platform={platform} />
            )}
          </SettingsBody>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

function SettingsNav({
  section,
  onSelect,
  configured,
  hasWarnings,
  appVersion,
}: {
  section: SettingsSection;
  onSelect: (s: SettingsSection) => void;
  configured: boolean;
  hasWarnings: boolean;
  appVersion: string;
}) {
  const aiItems = NAV.filter((n) => n.group === "ai");
  const generalItems = NAV.filter((n) => n.group === "general");
  const infoItems = NAV.filter((n) => n.group === "info");

  const renderItem = (item: (typeof NAV)[number]) => {
    const warn =
      (item.id === "connections" && !configured) ||
      (item.id === "models" && hasWarnings);
    return (
      <NavItem
        key={item.id}
        {...item}
        active={section === item.id}
        warn={warn}
        onClick={() => onSelect(item.id)}
      />
    );
  };

  return (
    <aside className="flex w-52 shrink-0 flex-col border-r bg-sidebar text-sidebar-foreground">
      <div className="flex-1 space-y-4 overflow-y-auto p-3">
        <NavGroup label="AI">{aiItems.map(renderItem)}</NavGroup>
        <NavGroup label="App">{generalItems.map(renderItem)}</NavGroup>
        <NavGroup label="Info">{infoItems.map(renderItem)}</NavGroup>
      </div>
      {appVersion && (
        <div className="border-t px-4 py-2 text-[10px] text-muted-foreground">
          shushing face v{appVersion}
        </div>
      )}
    </aside>
  );
}

function NavGroup({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1">
      <p className="px-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        {label}
      </p>
      <ul className="space-y-0.5">{children}</ul>
    </div>
  );
}

function NavItem({
  label,
  icon: Icon,
  active,
  warn,
  onClick,
}: {
  id: SettingsSection;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  active: boolean;
  warn?: boolean;
  onClick: () => void;
}) {
  return (
    <li>
      <button
        type="button"
        onClick={onClick}
        className={cn(
          "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors",
          active
            ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium"
            : "text-sidebar-foreground hover:bg-sidebar-accent/60",
        )}
      >
        <Icon className="size-3.5 shrink-0" />
        <span className="flex-1 truncate">{label}</span>
        {warn && (
          <AlertTriangle className="size-3 text-amber-500 shrink-0" />
        )}
      </button>
    </li>
  );
}

function SettingsBody({
  section,
  children,
}: {
  section: SettingsSection;
  children: React.ReactNode;
}) {
  const titles: Record<SettingsSection, string> = {
    connections: "Connections",
    models: "Models",
    rules: "Rules",
    styles: "Styles",
    appearance: "Appearance",
    general: "General",
    about: "About",
  };
  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <header className="flex h-12 items-center justify-between border-b px-6">
        <h2 className="text-sm font-semibold">{titles[section]}</h2>
        <DialogPrimitive.Close
          className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          aria-label="Close settings"
        >
          <X className="size-4" />
        </DialogPrimitive.Close>
      </header>
      <div className="flex flex-1 flex-col overflow-hidden">{children}</div>
    </div>
  );
}
